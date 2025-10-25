package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/phenrril/tienda3d/internal/domain"
)

type WhatsAppUC struct {
	WhatsAppRepo domain.WhatsAppRepo
	Products     domain.ProductRepo
	Orders       domain.OrderRepo
	Payments     *PaymentUC
}

// ProcessWhatsAppOrder procesa una orden recibida desde WhatsApp
func (uc *WhatsAppUC) ProcessWhatsAppOrder(ctx context.Context, whatsappOrder *domain.WhatsAppOrder) error {
	// Validar que la orden no haya sido procesada antes
	existing, err := uc.WhatsAppRepo.FindOrderByWhatsAppID(ctx, whatsappOrder.WhatsAppID)
	if err == nil && existing != nil {
		if existing.Status == "processed" {
			return fmt.Errorf("orden ya procesada: %s", whatsappOrder.WhatsAppID)
		}
	}

	// Crear la orden en el sistema Chroma3D
	order := &domain.Order{
		ID:             uuid.New(),
		Status:         domain.OrderStatusAwaitingPay,
		Email:          whatsappOrder.CustomerInfo.Email,
		Name:           whatsappOrder.CustomerInfo.Name,
		Phone:          whatsappOrder.CustomerInfo.Phone,
		Address:        whatsappOrder.CustomerInfo.Address,
		PostalCode:     whatsappOrder.CustomerInfo.PostalCode,
		Province:       whatsappOrder.CustomerInfo.City,
		ShippingMethod: "whatsapp",
		Total:          whatsappOrder.Total,
		ShippingCost:   0, // WhatsApp orders sin costo de envío inicial
		CreatedAt:      time.Now(),
	}

	// Procesar items de la orden
	for _, item := range whatsappOrder.Items {
		// Buscar el producto por slug
		product, err := uc.Products.FindBySlug(ctx, item.ProductSlug)
		if err != nil {
			// Si no se encuentra el producto, crear un item genérico
			order.Items = append(order.Items, domain.OrderItem{
				ID:        uuid.New(),
				Title:     item.Name,
				Color:     item.Color,
				Qty:       item.Quantity,
				UnitPrice: item.Price,
			})
			continue
		}

		// Crear el item con referencia al producto
		order.Items = append(order.Items, domain.OrderItem{
			ID:        uuid.New(),
			ProductID: &product.ID,
			Title:     item.Name,
			Color:     item.Color,
			Qty:       item.Quantity,
			UnitPrice: item.Price,
		})
	}

	// Guardar la orden
	if err := uc.Orders.Save(ctx, order); err != nil {
		return fmt.Errorf("error guardando orden: %w", err)
	}

	// Actualizar el estado de la orden de WhatsApp
	now := time.Now()
	whatsappOrder.Status = "processed"
	whatsappOrder.ProcessedAt = &now
	whatsappOrder.OrderID = &order.ID

	if err := uc.WhatsAppRepo.UpdateOrderStatus(ctx, whatsappOrder.ID, "processed", &order.ID); err != nil {
		return fmt.Errorf("error actualizando estado WhatsApp: %w", err)
	}

	return nil
}

// CreateWhatsAppOrder crea una nueva orden de WhatsApp
func (uc *WhatsAppUC) CreateWhatsAppOrder(ctx context.Context, whatsappID string, customerInfo domain.WhatsAppCustomer, items domain.WhatsAppItems) (*domain.WhatsAppOrder, error) {
	total := 0.0
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	order := &domain.WhatsAppOrder{
		ID:           uuid.New(),
		WhatsAppID:   whatsappID,
		CustomerInfo: customerInfo,
		Items:        items,
		Status:       "pending",
		Total:        total,
		CreatedAt:    time.Now(),
	}

	if err := uc.WhatsAppRepo.SaveOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("error guardando orden WhatsApp: %w", err)
	}

	return order, nil
}

// SyncProductToWhatsApp sincroniza un producto con WhatsApp Business
func (uc *WhatsAppUC) SyncProductToWhatsApp(ctx context.Context, product *domain.Product, whatsappProductID string) error {
	sync := &domain.WhatsAppProductSync{
		ID:                uuid.New(),
		ProductID:         product.ID,
		WhatsAppProductID: whatsappProductID,
		LastSynced:        time.Now(),
		SyncStatus:        "synced",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	return uc.WhatsAppRepo.SaveProductSync(ctx, sync)
}

// GetProductsForWhatsAppSync obtiene productos que necesitan sincronización
func (uc *WhatsAppUC) GetProductsForWhatsAppSync(ctx context.Context, page, pageSize int) ([]domain.Product, int64, error) {
	return uc.Products.List(ctx, domain.ProductFilter{
		Page:     page,
		PageSize: pageSize,
	})
}

// ConvertProductToWhatsAppFormat convierte un producto de Chroma3D al formato de WhatsApp
func (uc *WhatsAppUC) ConvertProductToWhatsAppFormat(product *domain.Product) map[string]interface{} {
	// Obtener la primera imagen disponible
	imageURL := ""
	if len(product.Images) > 0 {
		imageURL = product.Images[0].URL
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
		"name":         product.Name,
		"description":  description,
		"price":        int(product.BasePrice * 100), // WhatsApp espera precios en centavos
		"currency":     "ARS",
		"image_url":    imageURL,
		"is_available": product.ReadyToShip,
		"category":     product.Category,
		"url":          fmt.Sprintf("https://www.chroma3d.com.ar/product/%s", product.Slug),
	}

	// Agregar variantes de color si están disponibles
	if len(colors) > 0 {
		whatsappProduct["variants"] = map[string]interface{}{
			"color": colors,
		}
	}

	return whatsappProduct
}

// ProcessWhatsAppWebhook procesa un webhook recibido de WhatsApp
func (uc *WhatsAppUC) ProcessWhatsAppWebhook(ctx context.Context, payload []byte) error {
	var webhook map[string]interface{}
	if err := json.Unmarshal(payload, &webhook); err != nil {
		return fmt.Errorf("error parseando webhook: %w", err)
	}

	// Verificar el tipo de evento
	entry, ok := webhook["entry"].([]interface{})
	if !ok || len(entry) == 0 {
		return fmt.Errorf("webhook sin entry válido")
	}

	changes, ok := entry[0].(map[string]interface{})["changes"].([]interface{})
	if !ok || len(changes) == 0 {
		return fmt.Errorf("webhook sin changes válido")
	}

	change := changes[0].(map[string]interface{})
	value := change["value"].(map[string]interface{})

	// Procesar mensajes
	if messages, ok := value["messages"].([]interface{}); ok {
		for _, msgInterface := range messages {
			msg := msgInterface.(map[string]interface{})
			if err := uc.processWhatsAppMessage(ctx, msg); err != nil {
				// Log error but continue processing other messages
				continue
			}
		}
	}

	return nil
}

// processWhatsAppMessage procesa un mensaje individual de WhatsApp
func (uc *WhatsAppUC) processWhatsAppMessage(ctx context.Context, msg map[string]interface{}) error {
	// Verificar si es un mensaje de orden/carrito
	if interactive, ok := msg["interactive"].(map[string]interface{}); ok {
		if interactive["type"] == "button" {
			return uc.processOrderMessage(ctx, msg, interactive)
		}
	}

	// Verificar si es un mensaje de texto con información de orden
	if text, ok := msg["text"].(map[string]interface{}); ok {
		body := text["body"].(string)
		if strings.Contains(strings.ToLower(body), "orden") || strings.Contains(strings.ToLower(body), "pedido") {
			return uc.processTextOrder(ctx, msg, body)
		}
	}

	return nil
}

// processOrderMessage procesa un mensaje de orden desde botones interactivos
func (uc *WhatsAppUC) processOrderMessage(ctx context.Context, msg map[string]interface{}, interactive map[string]interface{}) error {
	// Extraer información del mensaje
	_ = msg["from"].(string)

	// Aquí implementarías la lógica específica para procesar órdenes desde botones
	// Por ejemplo, extraer información del carrito desde el payload del botón

	return nil
}

// processTextOrder procesa una orden enviada como texto
func (uc *WhatsAppUC) processTextOrder(ctx context.Context, msg map[string]interface{}, body string) error {
	// Extraer información del mensaje
	_ = msg["from"].(string)

	// Aquí implementarías la lógica para parsear órdenes desde texto
	// Por ejemplo, usar expresiones regulares para extraer productos y cantidades

	return nil
}
