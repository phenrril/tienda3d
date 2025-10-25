package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/phenrril/tienda3d/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Herramienta para sincronizar productos de Chroma3D con WhatsApp Business
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run whatsapp_sync.go <comando>")
		fmt.Println("Comandos disponibles:")
		fmt.Println("  export-products - Exporta productos en formato JSON para WhatsApp")
		fmt.Println("  sync-product <slug> <whatsapp_id> - Sincroniza un producto específico")
		fmt.Println("  list-products - Lista todos los productos disponibles")
		os.Exit(1)
	}

	command := os.Args[1]

	// Conectar a la base de datos
	db, err := connectDB()
	if err != nil {
		log.Fatal("Error conectando a la base de datos:", err)
	}

	switch command {
	case "export-products":
		exportProducts(db)
	case "sync-product":
		if len(os.Args) < 4 {
			log.Fatal("Uso: sync-product <slug> <whatsapp_id>")
		}
		syncProduct(db, os.Args[2], os.Args[3])
	case "list-products":
		listProducts(db)
	default:
		log.Fatal("Comando no reconocido:", command)
	}
}

func connectDB() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=tienda3d port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func exportProducts(db *gorm.DB) {
	var products []domain.Product
	if err := db.Preload("Images").Preload("Variants").Find(&products).Error; err != nil {
		log.Fatal("Error obteniendo productos:", err)
	}

	whatsappProducts := []map[string]interface{}{}

	for _, product := range products {
		// Obtener la primera imagen disponible
		imageURL := ""
		if len(product.Images) > 0 {
			imageURL = product.Images[0].URL
			// Convertir a URL absoluta si es relativa
			if !strings.HasPrefix(imageURL, "http") {
				baseURL := os.Getenv("BASE_URL")
				if baseURL == "" {
					baseURL = "https://www.chroma3d.com.ar"
				}
				imageURL = baseURL + imageURL
			}
		}

		// Obtener colores disponibles
		colors := []string{}
		seenColors := map[string]bool{}
		for _, variant := range product.Variants {
			if variant.Color != "" && !seenColors[variant.Color] {
				colors = append(colors, variant.Color)
				seenColors[variant.Color] = true
			}
		}

		// Crear descripción
		description := product.ShortDesc
		if description == "" {
			description = product.Name
		}

		// Agregar información de dimensiones si está disponible
		if product.WidthMM > 0 || product.HeightMM > 0 || product.DepthMM > 0 {
			description += fmt.Sprintf("\n\nDimensiones: %.0f x %.0f x %.0f mm",
				product.WidthMM, product.HeightMM, product.DepthMM)
		}

		whatsappProduct := map[string]interface{}{
			"name":          product.Name,
			"description":   description,
			"price":         int(product.BasePrice * 100), // WhatsApp espera precios en centavos
			"currency":      "ARS",
			"image_url":     imageURL,
			"is_available":  product.ReadyToShip,
			"category":      product.Category,
			"url":           fmt.Sprintf("https://www.chroma3d.com.ar/product/%s", product.Slug),
			"chroma3d_slug": product.Slug, // Para referencia interna
		}

		// Agregar variantes de color si están disponibles
		if len(colors) > 0 {
			whatsappProduct["variants"] = map[string]interface{}{
				"color": colors,
			}
		}

		whatsappProducts = append(whatsappProducts, whatsappProduct)
	}

	// Exportar como JSON
	output, err := json.MarshalIndent(whatsappProducts, "", "  ")
	if err != nil {
		log.Fatal("Error serializando productos:", err)
	}

	// Escribir a archivo
	filename := fmt.Sprintf("whatsapp_products_%s.json", time.Now().Format("20060102_150405"))
	if err := os.WriteFile(filename, output, 0644); err != nil {
		log.Fatal("Error escribiendo archivo:", err)
	}

	fmt.Printf("✅ Exportados %d productos a %s\n", len(whatsappProducts), filename)
	fmt.Println("\n📋 Para sincronizar con WhatsApp Business:")
	fmt.Println("1. Abre WhatsApp Business Manager")
	fmt.Println("2. Ve a Catálogo > Productos")
	fmt.Println("3. Importa el archivo JSON generado")
	fmt.Println("4. Usa el comando 'sync-product' para vincular cada producto con su ID de WhatsApp")
}

func syncProduct(db *gorm.DB, slug, whatsappID string) {
	// Buscar el producto
	var product domain.Product
	if err := db.Where("slug = ?", slug).First(&product).Error; err != nil {
		log.Fatal("Producto no encontrado:", err)
	}

	// Crear o actualizar la sincronización
	sync := domain.WhatsAppProductSync{
		ProductID:         product.ID,
		WhatsAppProductID: whatsappID,
		LastSynced:        time.Now(),
		SyncStatus:        "synced",
	}

	if err := db.Where("whatsapp_product_id = ?", whatsappID).FirstOrCreate(&sync).Error; err != nil {
		log.Fatal("Error sincronizando producto:", err)
	}

	fmt.Printf("✅ Producto '%s' sincronizado con WhatsApp ID: %s\n", product.Name, whatsappID)
}

func listProducts(db *gorm.DB) {
	var products []domain.Product
	if err := db.Select("slug, name, base_price, category, ready_to_ship").Find(&products).Error; err != nil {
		log.Fatal("Error obteniendo productos:", err)
	}

	fmt.Printf("📦 Productos disponibles (%d total):\n\n", len(products))
	for _, product := range products {
		status := "❌ No disponible"
		if product.ReadyToShip {
			status = "✅ Disponible"
		}
		fmt.Printf("• %s (%s) - $%.2f - %s\n", product.Name, product.Slug, product.BasePrice, status)
	}
}
