package app

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"

	"github.com/phenrril/tienda3d/internal/adapters/email/smtp"
	"github.com/phenrril/tienda3d/internal/adapters/httpserver"
	"github.com/phenrril/tienda3d/internal/adapters/payments/mercadopago"
	"github.com/phenrril/tienda3d/internal/adapters/repo/postgres"
	"github.com/phenrril/tienda3d/internal/adapters/storage/localfs"
	"github.com/phenrril/tienda3d/internal/domain"
	"github.com/phenrril/tienda3d/internal/usecase"
	"github.com/phenrril/tienda3d/internal/views"
)

type App struct {
	DB                  *gorm.DB
	Tmpl                *template.Template
	ProductUC           *usecase.ProductUC
	QuoteUC             *usecase.QuoteUC
	OrderUC             *usecase.OrderUC
	PaymentUC           *usecase.PaymentUC
	WhatsAppUC          *usecase.WhatsAppUC
	ModelRepo           domain.UploadedModelRepo
	FeaturedProductRepo domain.FeaturedProductRepo
	ShippingMethod      string  `gorm:"size:30"`
	ShippingCost        float64 `gorm:"type:decimal(12,2)"`
	Storage             domain.FileStorage
	Customers           domain.CustomerRepo
	OAuthConfig         *oauth2.Config
	EmailService        domain.EmailService
}

func NewApp(db *gorm.DB) (*App, error) {

	prodRepo := postgres.NewProductRepo(db)
	orderRepo := postgres.NewOrderRepo(db)
	modelRepo := postgres.NewUploadedModelRepo(db)
	custRepo := postgres.NewCustomerRepo(db)
	featuredRepo := postgres.NewFeaturedProductRepo(db)
	storageDir := os.Getenv("STORAGE_DIR")
	if storageDir == "" {
		storageDir = "uploads"
	}
	_ = os.MkdirAll(storageDir, 0755)
	log.Info().Str("storage_dir", storageDir).Msg("using storage directory")
	storage := localfs.New(storageDir)

	token := os.Getenv("MP_ACCESS_TOKEN")
	appEnv := strings.ToLower(os.Getenv("APP_ENV"))
	if appEnv == "production" || appEnv == "prod" {
		if prodTok := os.Getenv("PROD_ACCESS_TOKEN"); prodTok != "" {
			log.Info().Msg("usando credenciales MP producción")
			token = prodTok
		} else {
			log.Warn().Msg("APP_ENV=production pero falta PROD_ACCESS_TOKEN, usando MP_ACCESS_TOKEN")
		}
	} else {
		if strings.HasPrefix(token, "TEST-") {
			log.Info().Msg("modo sandbox MercadoPago (token TEST-)")
		} else {
			log.Info().Msg("APP_ENV no es production; usando token definido")
		}
	}

	payment := mercadopago.NewGateway(token)

	var oauthCfg *oauth2.Config
	googleID := os.Getenv("GOOGLE_CLIENT_ID")
	googleSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	if googleID != "" && googleSecret != "" {
		oauthCfg = &oauth2.Config{
			ClientID:     googleID,
			ClientSecret: googleSecret,
			RedirectURL:  baseURL + "/auth/google/callback",
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}
		log.Info().Msg("OAuth Google habilitado")
	} else {
		log.Warn().Msg("Google OAuth no configurado (faltan GOOGLE_CLIENT_ID / GOOGLE_CLIENT_SECRET)")
	}

	// Crear repositorio de WhatsApp
	whatsappRepo := postgres.NewWhatsAppRepo(db)

	// Inicializar servicio de email
	emailService := smtp.NewSMTPService()

	app := &App{}
	app.ProductUC = &usecase.ProductUC{Products: prodRepo}
	app.OrderUC = &usecase.OrderUC{Orders: orderRepo, Products: prodRepo}
	app.PaymentUC = &usecase.PaymentUC{Orders: orderRepo, Gateway: payment}
	app.WhatsAppUC = &usecase.WhatsAppUC{
		WhatsAppRepo: whatsappRepo,
		Products:     prodRepo,
		Orders:       orderRepo,
		Payments:     &usecase.PaymentUC{Orders: orderRepo, Gateway: payment},
	}
	app.DB = db
	app.ModelRepo = modelRepo
	app.FeaturedProductRepo = featuredRepo
	app.Storage = storage
	app.Customers = custRepo
	app.OAuthConfig = oauthCfg
	app.EmailService = emailService

	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		// colorhex: convierte un nombre genérico (es) o cualquier string a un hex de swatch
		"colorhex": func(s string) string {
			v := strings.TrimSpace(strings.ToLower(s))
			if v == "" {
				return "#334155"
			}
			if strings.HasPrefix(v, "#") {
				return v
			}
			m := map[string]string{
				"negro":       "#111827",
				"blanco":      "#ffffff",
				"azul":        "#3b82f6",
				"verde":       "#10b981",
				"amarillo":    "#f59e0b",
				"rojo":        "#ef4444",
				"violeta":     "#6366f1",
				"lila":        "#8b5cf6",
				"rosa":        "#ec4899",
				"turquesa":    "#14b8a6",
				"lima":        "#a3e635",
				"gris":        "#64748b",
				"gris oscuro": "#334155",
			}
			if hex, ok := m[v]; ok {
				return hex
			}
			return "#334155"
		},
		// img: normaliza URLs de imagen (agrega / si falta y codifica espacios)
		"img": func(u string) string {
			s := strings.TrimSpace(u)
			if s == "" {
				return s
			}
			if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") && !strings.HasPrefix(s, "/") {
				s = "/" + s
			}
			// codificar espacios para atributos src/srcset
			s = strings.ReplaceAll(s, " ", "%20")
			return s
		},
		// imgw: agrega parámetro ?w= para variantes responsivas
		"imgw": func(u string, w int) string {
			base := strings.TrimSpace(u)
			if base == "" {
				return base
			}
			if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") && !strings.HasPrefix(base, "/") {
				base = "/" + base
			}
			base = strings.ReplaceAll(base, " ", "%20")
			return fmt.Sprintf("%s?w=%d", base, w)
		},
		// formatPrice: formatea un número con puntos de miles (ej: 1000 -> "1.000", 1234.56 -> "1.234.56")
		"formatPrice": func(n float64) string {
			// Formatear con 2 decimales
			str := strconv.FormatFloat(n, 'f', 2, 64)
			parts := strings.Split(str, ".")
			intStr := parts[0]
			decStr := parts[1]
			
			// Agregar puntos de miles a la parte entera
			var result strings.Builder
			for i, r := range intStr {
				if i > 0 && (len(intStr)-i)%3 == 0 {
					result.WriteString(".")
				}
				result.WriteRune(r)
			}
			
			// Si los decimales son "00", no mostrarlos
			if decStr == "00" {
				return result.String()
			}
			return result.String() + "." + decStr
		},
	}
	tmpl, err := template.New("layout").Funcs(funcMap).ParseFS(views.FS, "*.html", "admin/*.html")
	if err != nil {
		return nil, err
	}
	app.Tmpl = tmpl

	return app, nil
}

func (a *App) HTTPHandler() http.Handler {
	return httpserver.New(a.Tmpl, a.ProductUC, a.QuoteUC, a.OrderUC, a.PaymentUC, a.WhatsAppUC, a.ModelRepo, a.Storage, a.Customers, a.OAuthConfig, a.FeaturedProductRepo, a.EmailService)
}

func (a *App) MigrateAndSeed() error {
	if err := a.DB.AutoMigrate(
		&domain.Product{}, &domain.Variant{}, &domain.Image{}, &domain.Order{}, &domain.OrderItem{}, &domain.UploadedModel{}, &domain.Quote{}, &domain.Page{}, &domain.Customer{}, &domain.WhatsAppOrder{}, &domain.WhatsAppProductSync{}, &domain.FeaturedProduct{},
	); err != nil {
		return err
	}

	if err := backfillSlugs(a.DB); err != nil {
		return err
	}

	return nil
}

func backfillSlugs(db *gorm.DB) error {
	var products []domain.Product
	if err := db.Where("slug IS NULL OR slug = ''").Find(&products).Error; err != nil {
		return err
	}
	for _, p := range products {
		base := strings.ToLower(strings.TrimSpace(p.Name))
		base = strings.ReplaceAll(base, " ", "-")
		if base == "" {
			base = p.ID.String()[:8]
		}
		slug := base

		var count int64
		i := 1
		for {
			if err := db.Model(&domain.Product{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				break
			}
			i++
			slug = fmt.Sprintf("%s-%d", base, i)
		}
		if err := db.Model(&domain.Product{}).Where("id = ?", p.ID).Update("slug", slug).Error; err != nil {
			return err
		}
	}

	if err := db.Exec("UPDATE products SET name = 'Producto' WHERE name IS NULL OR name = ''").Error; err != nil {
		return err
	}
	if err := db.Exec("UPDATE products SET base_price = 0 WHERE base_price IS NULL").Error; err != nil {
		return err
	}

	_ = db.Exec("ALTER TABLE products ALTER COLUMN slug SET NOT NULL").Error
	_ = db.Exec("ALTER TABLE products ALTER COLUMN name SET NOT NULL").Error
	_ = db.Exec("ALTER TABLE products ALTER COLUMN base_price SET NOT NULL").Error
	return nil
}

func seedProducts(db *gorm.DB) {
	prods := []domain.Product{
		{ID: uuid.New(), Slug: "llavero-logo", Name: "Llavero Logo", BasePrice: 1200, Category: "accesorios", ShortDesc: "Llavero impreso", ReadyToShip: true},
		{ID: uuid.New(), Slug: "lampara-luna", Name: "Lámpara Luna", BasePrice: 8500, Category: "iluminacion", ShortDesc: "Lámpara decorativa"},
		{ID: uuid.New(), Slug: "soporte-celular", Name: "Soporte Celular", BasePrice: 2500, Category: "hogar", ShortDesc: "Stand universal"},
		{ID: uuid.New(), Slug: "organizador-cables", Name: "Organizador Cables", BasePrice: 1800, Category: "oficina", ShortDesc: "Ordená tus cables"},
		{ID: uuid.New(), Slug: "maceta-geometrica", Name: "Maceta Geométrica", BasePrice: 3000, Category: "jardin"},
		{ID: uuid.New(), Slug: "clip-bolsa", Name: "Clip Bolsa", BasePrice: 600, Category: "cocina", ReadyToShip: true},
	}
	for _, p := range prods {
		db.Create(&p)
	}
}

func seedPages(db *gorm.DB) {
	pages := []domain.Page{{Slug: "about", Title: "Sobre Chroma3D", BodyMD: "Somos un taller de impresión 3D."}, {Slug: "contact", Title: "Contacto", BodyMD: "Escribinos a hola@example.com"}}
	for _, p := range pages {
		db.Create(&p)
	}
}
