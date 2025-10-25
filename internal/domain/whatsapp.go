package domain

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// WhatsAppOrder representa una orden recibida desde WhatsApp Business
type WhatsAppOrder struct {
	ID           uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	WhatsAppID   string           `gorm:"size:100;uniqueIndex" json:"whatsapp_id"` // ID del mensaje/orden en WhatsApp
	CustomerInfo WhatsAppCustomer `gorm:"type:jsonb" json:"customer_info"`
	Items        WhatsAppItems    `gorm:"type:jsonb" json:"items"`
	Status       string           `gorm:"size:50;index" json:"status"` // pending, processed, failed
	Total        float64          `gorm:"type:decimal(12,2)" json:"total"`
	Notes        string           `gorm:"type:text" json:"notes"`
	CreatedAt    time.Time        `json:"created_at"`
	ProcessedAt  *time.Time       `json:"processed_at,omitempty"`
	OrderID      *uuid.UUID       `gorm:"type:uuid;index" json:"order_id,omitempty"` // ID de la orden creada en Chroma3D
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

// Value implementa driver.Valuer para GORM
func (wc WhatsAppCustomer) Value() (driver.Value, error) {
	return json.Marshal(wc)
}

// Scan implementa sql.Scanner para GORM
func (wc *WhatsAppCustomer) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into WhatsAppCustomer", value)
	}

	return json.Unmarshal(bytes, wc)
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

// WhatsAppItems es un slice de WhatsAppItem con métodos para GORM
type WhatsAppItems []WhatsAppItem

// Value implementa driver.Valuer para GORM
func (wi WhatsAppItems) Value() (driver.Value, error) {
	return json.Marshal(wi)
}

// Scan implementa sql.Scanner para GORM
func (wi *WhatsAppItems) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into WhatsAppItems", value)
	}

	return json.Unmarshal(bytes, wi)
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
