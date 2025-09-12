package httpserver

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"regexp"
	"sort" // NUEVO para ordenar top productos y días
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	"github.com/phenrril/tienda3d/internal/adapters/payments/mercadopago"
	"github.com/phenrril/tienda3d/internal/domain"
	"github.com/phenrril/tienda3d/internal/usecase"
)

type Server struct {
	mux       *http.ServeMux
	tmpl      *template.Template
	products  *usecase.ProductUC
	quotes    *usecase.QuoteUC
	orders    *usecase.OrderUC
	payments  *usecase.PaymentUC
	models    domain.UploadedModelRepo
	storage   domain.FileStorage
	customers domain.CustomerRepo
	oauthCfg  *oauth2.Config
	// admin auth
	adminAllowed map[string]struct{}
	adminSecret  []byte
}

func New(t *template.Template, p *usecase.ProductUC, q *usecase.QuoteUC, o *usecase.OrderUC, pay *usecase.PaymentUC, m domain.UploadedModelRepo, fs domain.FileStorage, customers domain.CustomerRepo, oauthCfg *oauth2.Config) http.Handler {
	s := &Server{tmpl: t, products: p, quotes: q, orders: o, payments: pay, models: m, storage: fs, customers: customers, oauthCfg: oauthCfg, mux: http.NewServeMux()}
	// admin auth init
	allowed := map[string]struct{}{}
	if raw := os.Getenv("ADMIN_ALLOWED_EMAILS"); raw != "" {
		for _, e := range strings.Split(raw, ",") {
			e = strings.ToLower(strings.TrimSpace(e))
			if e != "" {
				allowed[e] = struct{}{}
			}
		}
	}
	s.adminAllowed = allowed
	sec := os.Getenv("JWT_ADMIN_SECRET")
	if sec == "" {
		sec = os.Getenv("SECRET_KEY")
	}
	if sec == "" {
		sec = "dev-admin-secret"
	}
	s.adminSecret = []byte(sec)

	s.routes()
	return Chain(s.mux, RateLimit(60), Gzip, RequestID, Recovery, Logging)
}

func (s *Server) routes() {
	// estáticos
	s.mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	// exponer uploads (imágenes guardadas vía API)
	s.mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	// SSR
	s.mux.HandleFunc("/", s.handleHome)
	s.mux.HandleFunc("/products", s.handleProducts)
	s.mux.HandleFunc("/product/", s.handleProduct)
	s.mux.HandleFunc("/quote/", s.handleQuoteView)
	s.mux.HandleFunc("/checkout", s.handleCheckout)
	s.mux.HandleFunc("/pay/", s.handlePaySimulated)
	// Cart
	s.mux.HandleFunc("/cart", s.handleCart) // GET ver / POST agregar
	s.mux.HandleFunc("/cart/update", s.handleCartUpdate)
	s.mux.HandleFunc("/cart/remove", s.handleCartRemove)
	s.mux.HandleFunc("/cart/checkout", s.handleCartCheckout)
	// API JSON
	s.mux.HandleFunc("/api/products", s.apiProducts)
	s.mux.HandleFunc("/api/products/", s.apiProductByID)
	// nuevo endpoint multipart
	s.mux.HandleFunc("/api/products/upload", s.apiProductUpload)
	s.mux.HandleFunc("/api/quote", s.apiQuote)
	s.mux.HandleFunc("/api/checkout", s.apiCheckout)
	s.mux.HandleFunc("/webhooks/mp", s.webhookMP)
	s.mux.HandleFunc("/api/products/delete", s.apiProductsBulkDelete)
	// Auth Google
	s.mux.HandleFunc("/auth/google/login", s.handleGoogleLogin)
	s.mux.HandleFunc("/auth/google/callback", s.handleGoogleCallback)
	s.mux.HandleFunc("/logout", s.handleLogout)
	// Admin auth (JSON legacy) y nuevo formulario
	s.mux.HandleFunc("/admin/login", s.handleAdminLogin) // existente
	s.mux.HandleFunc("/admin/auth", s.handleAdminAuth)   // nuevo formulario user/pass
	s.mux.HandleFunc("/admin/logout", s.handleAdminLogout)
	// Admin vistas
	s.mux.HandleFunc("/admin/orders", s.handleAdminOrders)
	s.mux.HandleFunc("/admin/products", s.handleAdminProducts)
	// NUEVO reporte ventas
	s.mux.HandleFunc("/admin/sales", s.handleAdminSales)
}

// ---------- SSR Handlers ----------
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	list, _, err := s.products.List(r.Context(), domain.ProductFilter{Page: 1, PageSize: 8})
	if err != nil {
		http.Error(w, "err", 500)
		return
	}
	data := map[string]any{"Products": list}
	if u := readUserSession(w, r); u != nil {
		data["User"] = u
	}
	s.render(w, "home.html", data)
}

func (s *Server) handleProducts(w http.ResponseWriter, r *http.Request) {
	qv := r.URL.Query()
	page, _ := strconv.Atoi(qv.Get("page"))
	if page < 1 {
		page = 1
	}
	sort := qv.Get("sort")
	query := qv.Get("q")
	category := qv.Get("category")
	stock := qv.Get("stock") // values: available, all
	var readyPtr *bool
	if stock == "available" {
		b := true
		readyPtr = &b
	}
	list, total, _ := s.products.List(r.Context(), domain.ProductFilter{Page: page, PageSize: 12, Sort: sort, Query: query, Category: category, ReadyToShip: readyPtr})
	pages := (int(total) + 11) / 12
	if pages == 0 {
		pages = 1
	}
	cats, _ := s.products.Categories(r.Context())
	data := map[string]any{
		"Products":    list,
		"Total":       total,
		"Page":        page,
		"Pages":       pages,
		"Query":       query,
		"Sort":        sort,
		"Category":    category,
		"StockFilter": stock,
		"Categories":  cats,
	}
	if u := readUserSession(w, r); u != nil {
		data["User"] = u
	}
	s.render(w, "products.html", data)
}

func (s *Server) handleProduct(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/product/")
	if slug == "" {
		http.NotFound(w, r)
		return
	}
	p, err := s.products.GetBySlug(r.Context(), slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// Recolectar colores únicos en orden de aparición (hasta 16)
	seen := map[string]struct{}{}
	colors := []string{}
	for _, v := range p.Variants {
		c := strings.TrimSpace(v.Color)
		if c == "" {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		colors = append(colors, c)
		if len(colors) == 16 {
			break
		}
	}
	// Paleta base para completar hasta 16 si faltan
	basePalette := []string{
		"#111827", // gris muy oscuro / fondo
		"#ffffff", // blanco
		"#6366f1", // indigo
		"#10b981", // verde
		"#f59e0b", // ámbar
		"#ef4444", // rojo
		"#3b82f6", // azul
		"#8b5cf6", // violeta
		"#ec4899", // rosa fuerte
		"#14b8a6", // teal
		"#f472b6", // rosa claro
		"#fcd34d", // amarillo claro
		"#a3e635", // lima
		"#dc2626", // rojo oscuro
		"#334155", // slate
		"#64748b", // slate claro
	}
	if len(colors) == 0 {
		// Si no hay variantes con color, usamos directamente la paleta completa
		colors = append([]string{}, basePalette...)
	} else if len(colors) < 16 {
		for _, c := range basePalette {
			if len(colors) == 16 {
				break
			}
			if _, ok := seen[c]; ok {
				continue
			}
			seen[c] = struct{}{}
			colors = append(colors, c)
		}
	}
	if len(colors) > 16 {
		colors = colors[:16]
	}
	added := 0
	if r.URL.Query().Get("added") == "1" {
		added = 1
	}
	data := map[string]any{"Product": p, "Colors": colors, "DefaultColor": colors[0], "Added": added}
	if u := readUserSession(w, r); u != nil {
		data["User"] = u
	}
	s.render(w, "product.html", data)
}

func (s *Server) handleQuoteView(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/quote/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	q, err := s.quotes.Quotes.FindByID(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	data := map[string]any{"Quote": q}
	if u := readUserSession(w, r); u != nil {
		data["User"] = u
	}
	s.render(w, "quote.html", data)
}

func (s *Server) handleCheckout(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{}
	if u := readUserSession(w, r); u != nil {
		data["User"] = u
	}
	s.render(w, "checkout.html", data)
}

// ---------- API ----------
func (s *Server) apiProducts(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	if r.Method == http.MethodGet {
		list, total, _ := s.products.List(r.Context(), domain.ProductFilter{Page: 1, PageSize: 100})
		writeJSON(w, 200, map[string]any{"items": list, "total": total})
		return
	}
	if r.Method == http.MethodPost {
		var req struct {
			Name        string  `json:"name"`
			Category    string  `json:"category"`
			ShortDesc   string  `json:"short_desc"`
			BasePrice   float64 `json:"base_price"`
			ReadyToShip bool    `json:"ready_to_ship"`
			WidthMM     float64 `json:"width_mm"`
			HeightMM    float64 `json:"height_mm"`
			DepthMM     float64 `json:"depth_mm"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "json", 400)
			return
		}
		if req.Name == "" || req.BasePrice < 0 || req.WidthMM < 0 || req.HeightMM < 0 || req.DepthMM < 0 {
			http.Error(w, "datos", 400)
			return
		}
		p := &domain.Product{Name: req.Name, Category: req.Category, ShortDesc: req.ShortDesc, BasePrice: req.BasePrice, ReadyToShip: req.ReadyToShip, WidthMM: req.WidthMM, HeightMM: req.HeightMM, DepthMM: req.DepthMM}
		if err := s.products.Create(r.Context(), p); err != nil {
			http.Error(w, "crear", 500)
			return
		}
		writeJSON(w, 201, p)
		return
	}
	http.Error(w, "method", 405)
}

func (s *Server) apiProductByID(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	if r.Method == http.MethodGet {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/products/")
		p, err := s.products.GetBySlug(r.Context(), idStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, 200, p)
		return
	}
	if r.Method == http.MethodPut { // update básico
		idStr := strings.TrimPrefix(r.URL.Path, "/api/products/")
		if idStr == "" {
			http.Error(w, "slug", 400)
			return
		}
		p, err := s.products.GetBySlug(r.Context(), idStr)
		if err != nil || p == nil {
			http.Error(w, "not found", 404)
			return
		}
		var req struct {
			Name        *string  `json:"name"`
			Category    *string  `json:"category"`
			ShortDesc   *string  `json:"short_desc"`
			BasePrice   *float64 `json:"base_price"`
			ReadyToShip *bool    `json:"ready_to_ship"`
			WidthMM     *float64 `json:"width_mm"`
			HeightMM    *float64 `json:"height_mm"`
			DepthMM     *float64 `json:"depth_mm"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "json", 400)
			return
		}
		if req.Name != nil {
			p.Name = *req.Name
		}
		if req.Category != nil {
			p.Category = *req.Category
		}
		if req.ShortDesc != nil {
			p.ShortDesc = *req.ShortDesc
		}
		if req.BasePrice != nil && *req.BasePrice >= 0 {
			p.BasePrice = *req.BasePrice
		}
		if req.ReadyToShip != nil {
			p.ReadyToShip = *req.ReadyToShip
		}
		if req.WidthMM != nil && *req.WidthMM >= 0 {
			p.WidthMM = *req.WidthMM
		}
		if req.HeightMM != nil && *req.HeightMM >= 0 {
			p.HeightMM = *req.HeightMM
		}
		if req.DepthMM != nil && *req.DepthMM >= 0 {
			p.DepthMM = *req.DepthMM
		}
		if err := s.products.Create(r.Context(), p); err != nil { // Save reutiliza Create()
			http.Error(w, "save", 500)
			return
		}
		writeJSON(w, 200, p)
		return
	}
	if r.Method == http.MethodDelete {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/products/")
		if idStr == "" {
			http.Error(w, "slug", 400)
			return
		}
		// eliminación completa (producto + imágenes + variantes + archivos)
		imgPaths, err := s.products.DeleteFullBySlug(r.Context(), idStr)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				http.Error(w, "not found", 404)
				return
			}
			http.Error(w, "delete", 500)
			return
		}
		removedFiles := []string{}
		for _, pth := range imgPaths {
			sp := strings.TrimSpace(pth)
			if sp == "" {
				continue
			}
			// normalizar ruta local: quitar prefijo '/'
			if strings.HasPrefix(sp, "/") {
				sp = sp[1:]
			}
			// sólo intentar si apunta dentro de uploads
			if !strings.Contains(sp, "uploads") {
				continue
			}
			if _, err := os.Stat(sp); err == nil {
				if err2 := os.Remove(sp); err2 == nil {
					removedFiles = append(removedFiles, sp)
				}
			}
		}
		writeJSON(w, 200, map[string]any{"status": "ok", "slug": idStr, "removed_files": removedFiles})
		return
	}
	http.Error(w, "method", 405)
}

func (s *Server) apiProductsBulkDelete(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	var req struct {
		Slugs []string `json:"slugs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Slugs) == 0 {
		http.Error(w, "json", 400)
		return
	}
	deleted := []string{}
	errorsMap := map[string]string{}
	for _, sl := range req.Slugs {
		if sl == "" {
			continue
		}
		if err := s.products.DeleteBySlug(r.Context(), sl); err != nil {
			errorsMap[sl] = err.Error()
		} else {
			deleted = append(deleted, sl)
		}
	}
	writeJSON(w, 200, map[string]any{"deleted": deleted, "errors": errorsMap})
}

func (s *Server) apiQuote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	var req struct {
		UploadedModelID string  `json:"uploaded_model_id"`
		Material        string  `json:"material"`
		Layer           float64 `json:"layer_height_mm"`
		Infill          int     `json:"infill_pct"`
		Quality         string  `json:"quality"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "json", 400)
		return
	}
	id, err := uuid.Parse(req.UploadedModelID)
	if err != nil {
		http.Error(w, "model", 400)
		return
	}
	model, err := s.models.FindByID(r.Context(), id)
	if err != nil {
		http.Error(w, "model", 404)
		return
	}
	q, err := s.quotes.CreateFromModel(r.Context(), model, domain.QuoteConfig{Material: domain.Material(req.Material), LayerHeightMM: req.Layer, InfillPct: req.Infill, Quality: domain.PrintQuality(req.Quality)})
	if err != nil {
		http.Error(w, "quote", 500)
		return
	}
	writeJSON(w, 200, q)
}

func (s *Server) apiCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	var req struct {
		QuoteID string `json:"quote_id"`
		Email   string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "json", 400)
		return
	}
	qid, err := uuid.Parse(req.QuoteID)
	if err != nil {
		http.Error(w, "quote", 400)
		return
	}
	q, err := s.quotes.Quotes.FindByID(r.Context(), qid)
	if err != nil {
		http.Error(w, "quote", 404)
		return
	}
	order, err := s.orders.CreateFromQuote(r.Context(), q, req.Email)
	if err != nil {
		http.Error(w, "order", 500)
		return
	}
	payURL, err := s.payments.CreatePreference(r.Context(), order)
	if err != nil {
		http.Error(w, "payment", 500)
		return
	}
	writeJSON(w, 200, map[string]any{"init_point": payURL, "order_id": order.ID})
}

func (s *Server) webhookMP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	body, _ := io.ReadAll(r.Body)
	var evt struct {
		Type   string `json:"type"`
		Action string `json:"action"`
		Data   struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(body, &evt)
	payID := evt.Data.ID
	if payID == "" {
		payID = r.URL.Query().Get("id")
	}
	if payID == "" {
		log.Warn().Msg("webhook sin payment id")
		w.WriteHeader(200)
		return
	}
	status, extRef, err := s.payments.Gateway.PaymentInfo(r.Context(), payID)
	if err != nil {
		log.Error().Err(err).Str("payment_id", payID).Msg("payment info")
		w.WriteHeader(200)
		return
	}
	orderID, ok := mercadopago.VerifyExternalRef(extRef)
	if !ok {
		log.Warn().Str("ext", extRef).Msg("external ref inválido")
		w.WriteHeader(200)
		return
	}
	uid, err := uuid.Parse(orderID)
	if err != nil {
		w.WriteHeader(200)
		return
	}
	o, err := s.orders.Orders.FindByID(r.Context(), uid)
	if err != nil || o == nil {
		log.Error().Err(err).Str("order_id", orderID).Msg("orden no encontrada para webhook")
		w.WriteHeader(200)
		return
	}
	approved := false
	switch status {
	case "approved":
		approved = true
		o.MPStatus = "approved"
		o.Status = domain.OrderStatusFinished
	case "pending", "in_process", "in_mediation":
		o.MPStatus = status
		if o.Status != domain.OrderStatusFinished {
			o.Status = domain.OrderStatusAwaitingPay
		}
	default:
		o.MPStatus = status
		if status == "rejected" {
			o.Status = domain.OrderStatusCancelled
		}
	}
	notify := false
	if approved && !o.Notified {
		o.Notified = true
		notify = true
	}
	if err := s.orders.Orders.Save(r.Context(), o); err != nil {
		log.Error().Err(err).Msg("guardar orden webhook")
	}
	if notify {
		go sendOrderNotify(o, true)
	}
	w.WriteHeader(200)
}

// --- Cart ---
// cartItem define items crudos en cookie
type cartItem struct {
	Slug  string  `json:"slug"`
	Color string  `json:"color"`
	Qty   int     `json:"qty"`
	Price float64 `json:"price"`
}

type cartPayload struct {
	Items []cartItem `json:"items"`
}

type cartLine struct {
	Slug      string
	Color     string
	Qty       int
	UnitPrice float64
	Subtotal  float64
	Name      string
	Image     string
}

func aggregateCart(cp cartPayload, lookup func(slug string) (*domain.Product, error)) []cartLine {
	m := map[string]*cartLine{}
	for _, it := range cp.Items {
		if it.Qty <= 0 {
			continue
		}
		key := it.Slug + "|" + it.Color
		line, ok := m[key]
		if !ok {
			line = &cartLine{Slug: it.Slug, Color: it.Color, Qty: 0, UnitPrice: it.Price}
			m[key] = line
		}
		line.Qty += it.Qty
	}
	res := []cartLine{}
	for _, l := range m {
		p, err := lookup(l.Slug)
		if err == nil && p != nil {
			l.Name = p.Name
			if len(p.Images) > 0 {
				l.Image = p.Images[0].URL
			}
			// usar precio actual base en vez del guardado si difiere
			if p.BasePrice != 0 {
				l.UnitPrice = p.BasePrice
			}
		}
		l.Subtotal = l.UnitPrice * float64(l.Qty)
		res = append(res, *l)
	}
	return res
}

var provinceCosts = map[string]float64{
	"Santa Fe":            9000,
	"Buenos Aires":        9000,
	"CABA":                9000,
	"Cordoba":             9000,
	"Entre Rios":          9000,
	"Corrientes":          9000,
	"Chaco":               9000,
	"Misiones":            9000,
	"Formosa":             9000,
	"Santiago del Estero": 9000,
	"Tucuman":             9000,
	"Salta":               9000,
	"Jujuy":               9000,
	"Catamarca":           9000,
	"La Rioja":            9000,
	"San Juan":            9000,
	"San Luis":            9000,
	"Mendoza":             9000,
	"La Pampa":            9000,
	"Neuquen":             9000,
	"Rio Negro":           9000,
	"Chubut":              9000,
	"Santa Cruz":          9000,
	"Tierra del Fuego":    9000,
}

func shippingCostFor(province string) float64 {
	if v, ok := provinceCosts[province]; ok {
		return v
	}
	if province == "" {
		return 0
	}
	return 9000
}

func (s *Server) handleCart(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		cp := readCart(r)
		lines := aggregateCart(cp, func(slug string) (*domain.Product, error) { return s.products.GetBySlug(r.Context(), slug) })
		total := 0.0
		for _, l := range lines {
			total += l.Subtotal
		}
		provs := []string{}
		for p := range provinceCosts {
			provs = append(provs, p)
		}
		data := map[string]any{"Lines": lines, "Total": total, "Provinces": provs, "ProvinceCosts": provinceCosts}
		if u := readUserSession(w, r); u != nil {
			data["User"] = u
		}
		s.render(w, "cart.html", data)
		return
	}
	if r.Method == http.MethodPost { // agregar
		if err := r.ParseForm(); err != nil {
			http.Error(w, "form", 400)
			return
		}
		slug := r.FormValue("slug")
		color := r.FormValue("color")
		if slug == "" {
			http.Error(w, "slug", 400)
			return
		}
		p, err := s.products.GetBySlug(r.Context(), slug)
		if err != nil {
			http.Error(w, "prod", 404)
			return
		}
		cart := readCart(r)
		cart.Items = append(cart.Items, cartItem{Slug: slug, Color: color, Qty: 1, Price: p.BasePrice})
		writeCart(w, cart)
		http.Redirect(w, r, "/product/"+slug+"?added=1", 302)
		return
	}
	http.Error(w, "method", 405)
}

func (s *Server) handleCartUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form", 400)
		return
	}
	slug := r.FormValue("slug")
	color := r.FormValue("color")
	op := r.FormValue("op") // inc / dec / set
	qtyStr := r.FormValue("qty")
	cart := readCart(r)
	// expand aggregated modifications re-creando lista
	agg := map[string]int{}
	for _, it := range cart.Items {
		if it.Qty > 0 {
			agg[it.Slug+"|"+it.Color] += it.Qty
		}
	}
	key := slug + "|" + color
	cur := agg[key]
	switch op {
	case "inc":
		cur++
	case "dec":
		if cur > 1 {
			cur--
		} else {
			cur = 0
		}
	case "set":
		if q, err := strconv.Atoi(qtyStr); err == nil {
			cur = q
		}
	}
	if cur < 0 {
		cur = 0
	}
	agg[key] = cur
	// rebuild
	newCart := cartPayload{}
	for k, q := range agg {
		if q <= 0 {
			continue
		}
		parts := strings.SplitN(k, "|", 2)
		newCart.Items = append(newCart.Items, cartItem{Slug: parts[0], Color: parts[1], Qty: q})
	}
	// need prices refreshed
	for i := range newCart.Items {
		p, _ := s.products.GetBySlug(r.Context(), newCart.Items[i].Slug)
		if p != nil {
			newCart.Items[i].Price = p.BasePrice
		}
	}
	writeCart(w, newCart)
	http.Redirect(w, r, "/cart", 302)
}

func (s *Server) handleCartRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form", 400)
		return
	}
	slug := r.FormValue("slug")
	color := r.FormValue("color")
	cart := readCart(r)
	newItems := []cartItem{}
	for _, it := range cart.Items {
		if !(it.Slug == slug && it.Color == color) {
			newItems = append(newItems, it)
		}
	}
	cart.Items = newItems
	writeCart(w, cart)
	http.Redirect(w, r, "/cart", 302)
}

func (s *Server) handleCartCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form", 400)
		return
	}
	email := r.FormValue("email")
	name := r.FormValue("name")
	phone := r.FormValue("phone")
	dni := r.FormValue("dni")
	postal := r.FormValue("postal_code")
	if email == "" || name == "" {
		http.Redirect(w, r, "/cart?err=datos", 302)
		return
	}
	shippingMethod := r.FormValue("shipping")
	if shippingMethod == "" {
		shippingMethod = "retiro"
	}
	// nuevas variantes de dirección por método
	addrEnvio := r.FormValue("address_envio")
	addrCadete := r.FormValue("address_cadete")
	legacyAddr := r.FormValue("address") // fallback si hubiera formularios viejos en cache
	province := r.FormValue("province")
	address := ""
	switch shippingMethod {
	case "envio":
		address = addrEnvio
	case "cadete":
		address = addrCadete
	default:
		address = legacyAddr
	}
	// Validaciones específicas por método
	if shippingMethod == "envio" {
		if province == "" || address == "" || postal == "" || dni == "" || phone == "" {
			http.Redirect(w, r, "/cart?err=envio", 302)
			return
		}
		dniRe := regexp.MustCompile(`^\d{7,8}$`)
		pcRe := regexp.MustCompile(`^\d{4,5}$`)
		if !dniRe.MatchString(dni) || !pcRe.MatchString(postal) {
			http.Redirect(w, r, "/cart?err=formato", 302)
			return
		}
	} else if shippingMethod == "cadete" { // método local
		if address == "" || phone == "" {
			http.Redirect(w, r, "/cart?err=cadete", 302)
			return
		}
		if province == "" { // fijar provincia default
			province = "Santa Fe"
		}
	}
	cp := readCart(r)
	if len(cp.Items) == 0 {
		http.Redirect(w, r, "/cart?err=vacio", 302)
		return
	}
	lines := aggregateCart(cp, func(slug string) (*domain.Product, error) { return s.products.GetBySlug(r.Context(), slug) })
	if len(lines) == 0 {
		http.Redirect(w, r, "/cart?err=vacio", 302)
		return
	}
	o := &domain.Order{ID: uuid.New(), Status: domain.OrderStatusAwaitingPay, Email: email, Name: name, Phone: phone, DNI: dni, PostalCode: postal, ShippingMethod: shippingMethod}
	itemsTotal := 0.0
	for _, l := range lines {
		p, _ := s.products.GetBySlug(r.Context(), l.Slug)
		var pid *uuid.UUID
		var title string
		if p != nil {
			pid = &p.ID
			title = p.Name
		} else {
			title = "Producto"
		}
		o.Items = append(o.Items, domain.OrderItem{ID: uuid.New(), ProductID: pid, Qty: l.Qty, UnitPrice: l.UnitPrice, Title: title, Color: l.Color})
		itemsTotal += l.UnitPrice * float64(l.Qty)
	}
	shippingCost := 0.0
	if shippingMethod == "envio" {
		shippingCost = shippingCostFor(province)
		if address == "" {
			address = "(sin dirección)"
		}
		o.Address = address
		o.Province = province
	} else if shippingMethod == "cadete" {
		shippingCost = 5000
		if address == "" {
			address = "(sin dirección)"
		}
		o.Address = address
		if province == "" {
			province = "Santa Fe"
		}
		o.Province = province
	}
	o.ShippingCost = shippingCost
	o.Total = itemsTotal + shippingCost
	if err := s.orders.Orders.Save(r.Context(), o); err != nil {
		http.Redirect(w, r, "/cart?err=orden", 302)
		return
	}
	redirURL, err := s.payments.CreatePreference(r.Context(), o)
	if err != nil {
		redirURL = "/pay/" + o.ID.String()
	} else {
		_ = s.orders.Orders.Save(r.Context(), o)
	}
	writeCart(w, cartPayload{})
	http.Redirect(w, r, redirURL, 302)
}

func (s *Server) handlePaySimulated(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/pay/")
	uid, err := uuid.Parse(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	o, err := s.orders.Orders.FindByID(r.Context(), uid)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	q := r.URL.Query()
	status := strings.ToLower(q.Get("status"))
	if status == "" {
		status = strings.ToLower(q.Get("collection_status"))
	}
	success := false
	if status == "approved" {
		success = true
	}
	if status != "" { // hubo callback real
		if success {
			o.MPStatus = "approved"
			o.Status = domain.OrderStatusFinished
			if !o.Notified {
				o.Notified = true
				_ = s.orders.Orders.Save(r.Context(), o)
				go sendOrderNotify(o, true)
			} else {
				_ = s.orders.Orders.Save(r.Context(), o)
			}
		} else {
			o.MPStatus = status
			_ = s.orders.Orders.Save(r.Context(), o)
		}
	}
	msg := "Pago pendiente / simulado"
	if status == "rejected" {
		msg = "El pago fue rechazado."
	}
	if success {
		msg = "Pago aprobado. Gracias por tu compra."
	}
	data := map[string]any{"Order": o, "StatusMsg": msg, "Success": success}
	if u := readUserSession(w, r); u != nil {
		data["User"] = u
	}
	s.render(w, "pay.html", data)
}

// ---------- Helpers ----------
func (s *Server) render(w http.ResponseWriter, name string, data any) {
	if m, ok := data.(map[string]any); ok {
		if _, exists := m["Year"]; !exists {
			m["Year"] = time.Now().Year()
		}
		if _, exists := m["User"]; !exists { // inyectar usuario
			if u := readUserSession(w, nil); u != nil {
				m["User"] = u
			}
		}
		data = m
	} else {
		m2 := map[string]any{"Year": time.Now().Year()}
		if u := readUserSession(w, nil); u != nil {
			m2["User"] = u
		}
		data = m2
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Error().Err(err).Str("tpl", name).Msg("render")
		http.Error(w, "tpl", 500)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func secretKey() []byte {
	k := os.Getenv("SESSION_KEY")
	if k == "" {
		k = "dev-insecure"
	}
	return []byte(k)
}

func readCart(r *http.Request) cartPayload {
	c, err := r.Cookie("cart")
	if err != nil {
		return cartPayload{}
	}
	parts := strings.SplitN(c.Value, ".", 2)
	if len(parts) != 2 {
		return cartPayload{}
	}
	sig, _ := base64.RawURLEncoding.DecodeString(parts[0])
	payload, _ := base64.RawURLEncoding.DecodeString(parts[1])
	h := hmac.New(sha256.New, secretKey())
	h.Write(payload)
	if !hmac.Equal(sig, h.Sum(nil)) {
		return cartPayload{}
	}
	var cp cartPayload
	_ = json.Unmarshal(payload, &cp)
	return cp
}

func writeCart(w http.ResponseWriter, cp cartPayload) {
	b, _ := json.Marshal(cp)
	h := hmac.New(sha256.New, secretKey())
	h.Write(b)
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	val := sig + "." + base64.RawURLEncoding.EncodeToString(b)
	http.SetCookie(w, &http.Cookie{Name: "cart", Value: val, Path: "/", MaxAge: 60 * 60 * 24 * 7, HttpOnly: true})
}

// apiProductUpload maneja creación de producto + imágenes (multipart/form-data)
func (s *Server) apiProductUpload(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	// límite memoria para cabeceras, archivos se almacenan en temp
	if err := r.ParseMultipartForm(25 << 20); err != nil { // 25MB
		http.Error(w, "multipart", 400)
		return
	}
	existingSlug := strings.TrimSpace(r.FormValue("existing_slug"))
	var p *domain.Product
	if existingSlug != "" { // añadir imágenes a producto existente
		if prod, err := s.products.GetBySlug(r.Context(), existingSlug); err == nil {
			p = prod
		} else {
			http.Error(w, "prod", 404)
			return
		}
	}
	if p == nil { // creación nueva con dimensiones
		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			http.Error(w, "name", 400)
			return
		}
		bp, _ := strconv.ParseFloat(r.FormValue("base_price"), 64)
		if bp < 0 {
			http.Error(w, "price", 400)
			return
		}
		cat := r.FormValue("category")
		sdesc := r.FormValue("short_desc")
		ready := r.FormValue("ready_to_ship") == "true" || r.FormValue("ready_to_ship") == "1"
		wm, _ := strconv.ParseFloat(r.FormValue("width_mm"), 64)
		hm, _ := strconv.ParseFloat(r.FormValue("height_mm"), 64)
		dm, _ := strconv.ParseFloat(r.FormValue("depth_mm"), 64)
		if wm < 0 {
			wm = 0
		}
		if hm < 0 {
			hm = 0
		}
		if dm < 0 {
			dm = 0
		}
		p = &domain.Product{Name: name, Category: cat, ShortDesc: sdesc, BasePrice: bp, ReadyToShip: ready, WidthMM: wm, HeightMM: hm, DepthMM: dm}
		if err := s.products.Create(r.Context(), p); err != nil {
			log.Error().Err(err).Msg("crear producto")
			http.Error(w, "crear", 500)
			return
		}
	}
	// procesar archivos (image o images)
	files := []*multipart.FileHeader{}
	if r.MultipartForm != nil {
		if fhArr, ok := r.MultipartForm.File["image"]; ok {
			files = append(files, fhArr...)
		}
		if fhArr, ok := r.MultipartForm.File["images"]; ok {
			files = append(files, fhArr...)
		}
	}
	imgs := []domain.Image{}
	for _, fh := range files {
		if fh.Size == 0 {
			continue
		}
		f, err := fh.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(f)
		_ = f.Close()
		if err != nil || len(data) == 0 {
			continue
		}
		storedPath, err := s.storage.SaveImage(r.Context(), fh.Filename, data)
		if err != nil {
			log.Warn().Err(err).Str("file", fh.Filename).Msg("no se pudo guardar imagen")
			continue
		}
		if !strings.HasPrefix(storedPath, "/") {
			storedPath = "/" + strings.ReplaceAll(storedPath, "\\", "/")
		}
		imgs = append(imgs, domain.Image{URL: storedPath, Alt: p.Name})
	}
	if len(imgs) > 0 {
		if err := s.products.AddImages(r.Context(), p.ID, imgs); err != nil {
			log.Error().Err(err).Msg("add images")
		}
		if rp, err := s.products.GetBySlug(r.Context(), p.Slug); err == nil {
			p = rp
		}
	}
	writeJSON(w, 201, map[string]any{"product": p, "added_images": len(imgs)})
}

// --- Admin vistas nuevas ---
func (s *Server) handleAdminProducts(w http.ResponseWriter, r *http.Request) {
	if !s.isAdminSession(r) {
		http.Redirect(w, r, "/admin/auth", 302)
		return
	}
	list, total, _ := s.products.List(r.Context(), domain.ProductFilter{Page: 1, PageSize: 200})
	// incluir token para JS
	tok := s.readAdminToken(r)
	data := map[string]any{"Products": list, "Total": total, "AdminToken": tok}
	s.render(w, "admin_products.html", data)
}

// reemplazar template faltante admin_orders.html
func (s *Server) handleAdminOrders(w http.ResponseWriter, r *http.Request) {
	if !s.isAdminSession(r) {
		http.Redirect(w, r, "/admin/auth", 302)
		return
	}
	page := 1
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	var mpStatus *string
	filterApproved := false
	if r.URL.Query().Get("approved") == "1" { // checkbox activado
		st := "approved"
		mpStatus = &st
		filterApproved = true
	}
	list, total, err := s.orders.Orders.List(r.Context(), nil, mpStatus, page, 20)
	if err != nil {
		http.Error(w, "err", 500)
		return
	}
	pages := (int(total) + 19) / 20
	data := map[string]any{"Orders": list, "Page": page, "Pages": pages, "AdminToken": s.readAdminToken(r), "FilterApproved": filterApproved}
	s.render(w, "admin_orders.html", data)
}

// NUEVO: Reporte de ventas
func (s *Server) handleAdminSales(w http.ResponseWriter, r *http.Request) {
	if !s.isAdminSession(r) {
		http.Redirect(w, r, "/admin/auth", 302)
		return
	}
	q := r.URL.Query()
	layout := "admin_sales.html"
	// rango de fechas (por defecto últimos 30 días)
	const layoutIn = "2006-01-02"
	var (
		to   time.Time
		from time.Time
		err  error
	)
	if ds := q.Get("to"); ds != "" {
		to, err = time.Parse(layoutIn, ds)
		if err != nil {
			to = time.Now()
		}
	} else {
		to = time.Now()
	}
	if ds := q.Get("from"); ds != "" {
		from, err = time.Parse(layoutIn, ds)
		if err != nil {
			from = to.AddDate(0, 0, -29)
		}
	} else {
		from = to.AddDate(0, 0, -29)
	}
	if from.After(to) {
		from, to = to, from
	}

	ordersAll, err := s.orders.Orders.ListInRange(r.Context(), from, to)
	if err != nil {
		http.Error(w, "err", 500)
		return
	}
	// NUEVO: solo considerar órdenes aprobadas por MP
	orders := make([]domain.Order, 0, len(ordersAll))
	for _, o := range ordersAll {
		if strings.EqualFold(o.MPStatus, "approved") { // considerar sólo aprobadas
			orders = append(orders, o)
		}
	}

	// Agregados (solo aprobadas)
	var totalRevenue, shippingRevenue float64
	statusCounts := map[string]int{}
	mpStatusCounts := map[string]int{}
	shippingMethodCounts := map[string]int{}
	provinceCounts := map[string]int{}
	itemsRevenue := 0.0
	productAgg := map[string]struct {
		Title   string
		Qty     int
		Revenue float64
	}{}
	dayRevenue := map[string]struct {
		Revenue float64
		Orders  int
	}{}

	for _, o := range orders {
		totalRevenue += o.Total
		shippingRevenue += o.ShippingCost
		statusCounts[string(o.Status)]++
		if o.MPStatus != "" {
			mpStatusCounts[o.MPStatus]++
		}
		if o.ShippingMethod != "" {
			shippingMethodCounts[o.ShippingMethod]++
		}
		if o.Province != "" {
			provinceCounts[o.Province]++
		}
		dayKey := o.CreatedAt.Format("2006-01-02")
		dr := dayRevenue[dayKey]
		dr.Revenue += o.Total
		dr.Orders++
		dayRevenue[dayKey] = dr
		lineItems := 0.0
		for _, it := range o.Items {
			lineRev := it.UnitPrice * float64(it.Qty)
			lineItems += lineRev
			key := it.Title
			cur := productAgg[key]
			cur.Title = it.Title
			cur.Qty += it.Qty
			cur.Revenue += lineRev
			productAgg[key] = cur
		}
		itemsRevenue += lineItems
	}
	avgOrderValue := 0.0
	if len(orders) > 0 {
		avgOrderValue = totalRevenue / float64(len(orders))
	}
	// ordenar top productos
	prodList := make([]struct {
		Title   string
		Qty     int
		Revenue float64
	}, 0, len(productAgg))
	for _, v := range productAgg {
		prodList = append(prodList, v)
	}
	sort.Slice(prodList, func(i, j int) bool {
		if prodList[i].Qty == prodList[j].Qty {
			return prodList[i].Revenue > prodList[j].Revenue
		}
		return prodList[i].Qty > prodList[j].Qty
	})
	if len(prodList) > 25 {
		prodList = prodList[:25]
	}
	// ordenar días cronológicamente
	dayKeys := make([]string, 0, len(dayRevenue))
	for k := range dayRevenue {
		dayKeys = append(dayKeys, k)
	}
	sort.Strings(dayKeys)
	daySeries := []struct {
		Day     string
		Revenue float64
		Orders  int
	}{}
	for _, k := range dayKeys {
		v := dayRevenue[k]
		daySeries = append(daySeries, struct {
			Day     string
			Revenue float64
			Orders  int
		}{Day: k, Revenue: v.Revenue, Orders: v.Orders})
	}

	if strings.ToLower(q.Get("format")) == "csv" { // export solo aprobadas
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=ventas_%s_%s.csv", from.Format(layoutIn), to.Format(layoutIn)))
		fmt.Fprintln(w, "order_id,created_at,status,mp_status,total,shipping_method,shipping_cost,province")
		for _, o := range orders {
			fmt.Fprintf(w, "%s,%s,%s,%s,%.2f,%s,%.2f,%s\n", o.ID, o.CreatedAt.Format(time.RFC3339), o.Status, o.MPStatus, o.Total, o.ShippingMethod, o.ShippingCost, strings.ReplaceAll(o.Province, ",", " "))
		}
		return
	}

	data := map[string]any{
		"From":                 from.Format(layoutIn),
		"To":                   to.Format(layoutIn),
		"OrdersCount":          len(orders),
		"TotalRevenue":         totalRevenue,
		"ItemsRevenue":         itemsRevenue,
		"ShippingRevenue":      shippingRevenue,
		"AvgOrderValue":        avgOrderValue,
		"StatusCounts":         statusCounts,
		"MPStatusCounts":       mpStatusCounts,
		"ShippingMethodCounts": shippingMethodCounts,
		"ProvinceCounts":       provinceCounts,
		"TopProducts":          prodList,
		"DailySeries":          daySeries,
		"AdminToken":           s.readAdminToken(r),
	}
	// inyectar para layout nav etc
	s.render(w, layout, data)
}

func (s *Server) handleAdminAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if s.isAdminSession(r) {
			http.Redirect(w, r, "/admin/products", 302)
			return
		}
		data := map[string]any{}
		s.render(w, "admin_auth.html", data)
		return
	}
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "form", 400)
			return
		}
		user := strings.TrimSpace(r.FormValue("user"))
		pass := strings.TrimSpace(r.FormValue("pass"))
		cfgUser := os.Getenv("ADMIN_USER")
		cfgPass := os.Getenv("ADMIN_PASS")
		if cfgUser == "" {
			cfgUser = "admin"
		}
		if cfgPass == "" {
			cfgPass = "admin123"
		}
		if user != cfgUser || pass != cfgPass {
			http.Error(w, "credenciales", 401)
			return
		}
		email := user + "@local"
		if _, ok := s.adminAllowed[email]; !ok {
			if len(s.adminAllowed) == 0 {
				s.adminAllowed[email] = struct{}{}
			} else {
				for k := range s.adminAllowed {
					email = k
					break
				}
			}
		}
		tok, _, err := s.issueAdminToken(email, 12*time.Hour)
		if err != nil {
			http.Error(w, "token", 500)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "admin_token", Value: tok, Path: "/", MaxAge: 60 * 60 * 12})
		http.Redirect(w, r, "/admin/products", 302)
		return
	}
	http.Error(w, "method", 405)
}

func (s *Server) handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "admin_token", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/admin/auth", 302)
}

func (s *Server) isAdminSession(r *http.Request) bool {
	if tok := s.readAdminToken(r); tok != "" {
		if _, err := s.verifyAdminToken(tok); err == nil {
			return true
		}
	}
	return false
}

func (s *Server) readAdminToken(r *http.Request) string {
	c, err := r.Cookie("admin_token")
	if err != nil || c.Value == "" {
		return ""
	}
	return c.Value
}

func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		tok := strings.TrimSpace(auth[7:])
		if _, err := s.verifyAdminToken(tok); err == nil {
			return true
		}
	}
	// fallback cookie
	if tok := s.readAdminToken(r); tok != "" {
		if _, err := s.verifyAdminToken(tok); err == nil {
			return true
		}
	}
	http.Error(w, "unauthorized", 401)
	return false
}

func sendOrderEmail(o *domain.Order, success bool) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	to := os.Getenv("ORDER_NOTIFY_EMAIL")
	if to == "" {
		to = "chroma3dimpresiones@gmail.com"
	}
	if host == "" || port == "" || user == "" || pass == "" {
		log.Warn().Msg("SMTP no configurado, se omite envío de email")
		return nil
	}
	addr := host + ":" + port
	statusTxt := "PAGO FALLIDO"
	if success {
		statusTxt = "PAGO APROBADO"
	}
	var buf bytes.Buffer
	_, _ = fmt.Fprintf(&buf, "Subject: Nueva orden %s #%s\r\n", statusTxt, o.ID.String())
	_, _ = fmt.Fprintf(&buf, "From: %s\r\n", user)
	_, _ = fmt.Fprintf(&buf, "To: %s\r\n", to)
	buf.WriteString("MIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n")
	_, _ = fmt.Fprintf(&buf, "Estado: %s\n", statusTxt)
	_, _ = fmt.Fprintf(&buf, "Orden: %s\n", o.ID)
	_, _ = fmt.Fprintf(&buf, "Nombre: %s\nEmail: %s\nTel: %s\nDNI: %s\n", o.Name, o.Email, o.Phone, o.DNI)
	if o.ShippingMethod == "envio" || o.ShippingMethod == "cadete" {
		_, _ = fmt.Fprintf(&buf, "Envío (%s) a: %s (%s) CP:%s\n", o.ShippingMethod, o.Address, o.Province, o.PostalCode)
	} else {
		buf.WriteString("Retiro en local\n")
	}
	buf.WriteString("Items:\n")
	for _, it := range o.Items {
		_, _ = fmt.Fprintf(&buf, "- %s x%d $%.2f Color:%s\n", it.Title, it.Qty, it.UnitPrice, it.Color)
	}
	_, _ = fmt.Fprintf(&buf, "Total: $%.2f (Envío: $%.2f)\n", o.Total, o.ShippingCost)
	auth := smtp.PlainAuth("", user, pass, host)
	if err := smtp.SendMail(addr, auth, user, []string{to}, buf.Bytes()); err != nil {
		log.Error().Err(err).Msg("email send")
		return err
	}
	return nil
}

func sendOrderTelegram(o *domain.Order, success bool) error {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if token == "" || chatID == "" {
		return fmt.Errorf("telegram vars faltantes")
	}
	statusTxt := "PAGO FALLIDO"
	if success {
		statusTxt = "PAGO APROBADO"
	}
	var b strings.Builder
	b.WriteString("Orden ")
	b.WriteString(o.ID.String())
	b.WriteString(" - ")
	b.WriteString(statusTxt)
	b.WriteString("\n")
	fmt.Fprintf(&b, "Nombre: %s\nEmail: %s\nTel: %s\nDNI: %s\n", o.Name, o.Email, o.Phone, o.DNI)
	if o.ShippingMethod == "envio" || o.ShippingMethod == "cadete" {
		fmt.Fprintf(&b, "Envío (%s) a: %s (%s %s) CP:%s\n", o.ShippingMethod, o.Address, o.Province, o.ShippingMethod, o.PostalCode)
	} else {
		b.WriteString("Retiro en local\n")
	}
	b.WriteString("Items:\n")
	for _, it := range o.Items {
		fmt.Fprintf(&b, "- %s x%d $%.2f %s\n", it.Title, it.Qty, it.UnitPrice, it.Color)
	}
	fmt.Fprintf(&b, "Total: $%.2f (Envio: $%.2f)\n", o.Total, o.ShippingCost)
	apiURL := "https://api.telegram.org/bot" + token + "/sendMessage"
	form := url.Values{}
	form.Set("chat_id", chatID)
	form.Set("text", b.String())
	form.Set("disable_web_page_preview", "1")
	resp, err := http.PostForm(apiURL, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func sendOrderNotify(o *domain.Order, success bool) {
	if err := sendOrderTelegram(o, success); err != nil {
		log.Warn().Err(err).Msg("telegram notif fallo")
		if os.Getenv("SMTP_HOST") != "" {
			_ = sendOrderEmail(o, success)
		}
	}
}

// sesiones usuario (Google OAuth)
type sessionUser struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func writeUserSession(w http.ResponseWriter, u *sessionUser) {
	if u == nil { // clear
		http.SetCookie(w, &http.Cookie{Name: "sess", Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
		return
	}
	b, _ := json.Marshal(u)
	h := hmac.New(sha256.New, secretKey())
	h.Write(b)
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	val := sig + "." + base64.RawURLEncoding.EncodeToString(b)
	http.SetCookie(w, &http.Cookie{Name: "sess", Value: val, Path: "/", MaxAge: 60 * 60 * 24 * 30, HttpOnly: true})
}

func readUserSession(w http.ResponseWriter, r *http.Request) *sessionUser {
	if r == nil {
		return nil
	}
	c, err := r.Cookie("sess")
	if err != nil || c.Value == "" {
		return nil
	}
	parts := strings.SplitN(c.Value, ".", 2)
	if len(parts) != 2 {
		return nil
	}
	sig, _ := base64.RawURLEncoding.DecodeString(parts[0])
	payload, _ := base64.RawURLEncoding.DecodeString(parts[1])
	h := hmac.New(sha256.New, secretKey())
	h.Write(payload)
	if !hmac.Equal(sig, h.Sum(nil)) {
		return nil
	}
	var u sessionUser
	if err := json.Unmarshal(payload, &u); err != nil {
		return nil
	}
	if u.Email == "" {
		return nil
	}
	return &u
}

func (s *Server) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if s.oauthCfg == nil {
		http.Error(w, "oauth no configurado", 500)
		return
	}
	state := uuid.New().String()
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: state, Path: "/", MaxAge: 300, HttpOnly: true, Secure: false})
	loginURL := s.oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
	http.Redirect(w, r, loginURL, 302)
}

func (s *Server) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if s.oauthCfg == nil {
		http.Error(w, "oauth no configurado", 500)
		return
	}
	q := r.URL.Query()
	state := q.Get("state")
	code := q.Get("code")
	c, _ := r.Cookie("oauth_state")
	if c == nil || c.Value == "" || c.Value != state {
		http.Error(w, "state", 400)
		return
	}
	tok, err := s.oauthCfg.Exchange(r.Context(), code)
	if err != nil {
		log.Error().Err(err).Msg("exchange oauth")
		http.Error(w, "oauth", 400)
		return
	}
	client := s.oauthCfg.Client(r.Context(), tok)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil || resp.StatusCode != 200 {
		log.Error().Err(err).Msg("userinfo")
		http.Error(w, "userinfo", 400)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var info struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	_ = json.Unmarshal(body, &info)
	if info.Email == "" {
		http.Error(w, "email", 400)
		return
	}
	if s.customers != nil {
		if cust, err := s.customers.FindByEmail(r.Context(), info.Email); err != nil && err == domain.ErrNotFound {
			_ = s.customers.Save(r.Context(), &domain.Customer{ID: uuid.New(), Email: info.Email, Name: info.Name})
		} else if cust == nil && err == nil {
			_ = s.customers.Save(r.Context(), &domain.Customer{ID: uuid.New(), Email: info.Email, Name: info.Name})
		}
	}
	writeUserSession(w, &sessionUser{Email: info.Email, Name: info.Name})
	http.Redirect(w, r, "/", 302)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	writeUserSession(w, nil)
	http.Redirect(w, r, "/", 302)
}

// --- Admin Auth (legacy API key + email whitelist) ---
func (s *Server) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", 405)
		return
	}
	cfgKey := os.Getenv("ADMIN_API_KEY")
	if cfgKey == "" {
		log.Error().Msg("ADMIN_API_KEY faltante")
		http.Error(w, "config", 500)
		return
	}
	apiKey := r.Header.Get("X-Admin-Key")
	if apiKey == "" || !secureCompare(apiKey, cfgKey) {
		http.Error(w, "unauthorized", 401)
		return
	}
	var req struct {
		Email string `json:"email"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" && len(s.adminAllowed) == 1 {
		for k := range s.adminAllowed {
			email = k
		}
	}
	if _, ok := s.adminAllowed[email]; !ok {
		http.Error(w, "forbidden", 403)
		return
	}
	tok, exp, err := s.issueAdminToken(email, 30*time.Minute)
	if err != nil {
		http.Error(w, "token", 500)
		return
	}
	writeJSON(w, 200, map[string]any{"token": tok, "exp": exp.Unix(), "email": email})
}

func (s *Server) issueAdminToken(email string, dur time.Duration) (string, time.Time, error) {
	head := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	exp := time.Now().Add(dur)
	claims := map[string]any{"sub": email, "email": email, "role": "admin", "exp": exp.Unix(), "iat": time.Now().Unix(), "iss": "tienda3d"}
	b, _ := json.Marshal(claims)
	pay := base64.RawURLEncoding.EncodeToString(b)
	unsigned := head + "." + pay
	h := hmac.New(sha256.New, s.adminSecret)
	h.Write([]byte(unsigned))
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return unsigned + "." + sig, exp, nil
}

func (s *Server) verifyAdminToken(tok string) (string, error) {
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("formato")
	}
	unsigned := parts[0] + "." + parts[1]
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", fmt.Errorf("sig")
	}
	h := hmac.New(sha256.New, s.adminSecret)
	h.Write([]byte(unsigned))
	if !hmac.Equal(sig, h.Sum(nil)) {
		return "", fmt.Errorf("firma")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("payload")
	}
	var m map[string]any
	if err := json.Unmarshal(payload, &m); err != nil {
		return "", fmt.Errorf("json")
	}
	role, _ := m["role"].(string)
	email, _ := m["email"].(string)
	expF, _ := m["exp"].(float64)
	if role != "admin" || email == "" {
		return "", fmt.Errorf("claims")
	}
	if time.Now().Unix() > int64(expF) {
		return "", fmt.Errorf("exp")
	}
	if _, ok := s.adminAllowed[strings.ToLower(email)]; !ok {
		return "", fmt.Errorf("not allowed")
	}
	return email, nil
}

func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}
