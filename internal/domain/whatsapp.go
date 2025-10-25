package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// WhatsAppOrder representa una orden recibida desde WhatsApp Business
type WhatsAppOrder struct {
	ID           uuid.UUID        `json:"id"`
	WhatsAppID   string           `json:"whatsapp_id"` // ID del mensaje/orden en WhatsApp
	CustomerInfo WhatsAppCustomer `json:"customer_info"`
	Items        []WhatsAppItem   `json:"items"`
	Status       string           `json:"status"` // pending, processed, failed
	Total        float64          `json:"total"`
	Notes        string           `json:"notes"`
	CreatedAt    time.Time        `json:"created_at"`
	ProcessedAt  *time.Time       `json:"processed_at,omitempty"`
	OrderID      *uuid.UUID       `json:"order_id,omitempty"` // ID de la orden creada en Chroma3D
}

// WhatsAppCustomer información del cliente desde WhatsApp
type WhatsAppCustomer struct {
	Phone      string `json:"phone"`
	Name       string `json:"name,omitempty"`
	Email      string `json:"email,omitempty"`
	Address    string `json:"address,omitempty"`
	City       string `json:"city,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
}

// WhatsAppItem representa un producto en la orden de WhatsApp
type WhatsAppItem struct {
	ProductID   string  `json:"product_id"`   // ID del producto en WhatsApp
	ProductSlug string  `json:"product_slug"` // Slug del producto en Chroma3D
	Name        string  `json:"name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Color       string  `json:"color,omitempty"`
	Image       string  `json:"image,omitempty"`
}

// WhatsAppProductSync representa un producto sincronizado con WhatsApp
type WhatsAppProductSync struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductID         uuid.UUID `gorm:"type:uuid;index"`
	WhatsAppProductID string    `gorm:"size:100;uniqueIndex"`
	LastSynced        time.Time `gorm:"index"`
	SyncStatus        string    `gorm:"size:50"` // synced, failed, pending
	ErrorMessage      string    `gorm:"type:text"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// WhatsAppRepo define las operaciones para manejar órdenes de WhatsApp
type WhatsAppRepo interface {
	SaveOrder(ctx context.Context, order *WhatsAppOrder) error
	FindOrderByWhatsAppID(ctx context.Context, whatsappID string) (*WhatsAppOrder, error)
	ListPendingOrders(ctx context.Context) ([]WhatsAppOrder, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string, orderID *uuid.UUID) error

	// Sincronización de productos
	SaveProductSync(ctx context.Context, sync *WhatsAppProductSync) error
	FindProductSyncByWhatsAppID(ctx context.Context, whatsappID string) (*WhatsAppProductSync, error)
	ListProductsToSync(ctx context.Context) ([]WhatsAppProductSync, error)
	UpdateSyncStatus(ctx context.Context, id uuid.UUID, status string, errorMsg string) error
}
