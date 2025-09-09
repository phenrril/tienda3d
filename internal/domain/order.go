package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPendingQuote OrderStatus = "pending_quote"
	OrderStatusQuoted       OrderStatus = "quoted"
	OrderStatusAwaitingPay  OrderStatus = "awaiting_payment"
	OrderStatusInPrint      OrderStatus = "in_print"
	OrderStatusFinished     OrderStatus = "finished"
	OrderStatusShipped      OrderStatus = "shipped"
	OrderStatusCancelled    OrderStatus = "cancelled"
)

type Order struct {
	ID             uuid.UUID   `gorm:"type:uuid;primaryKey"`
	Status         OrderStatus `gorm:"type:varchar(30);index"`
	Items          []OrderItem
	Email          string  `gorm:"size:140"`
	Name           string  `gorm:"size:140"`
	Phone          string  `gorm:"size:50"`
	DNI            string  `gorm:"size:30"`
	Address        string  `gorm:"size:255"`
	PostalCode     string  `gorm:"size:20"`
	Province       string  `gorm:"size:80"`
	MPPreferenceID string  `gorm:"size:140"`
	MPStatus       string  `gorm:"size:60"`
	Total          float64 `gorm:"type:decimal(12,2)"`
	ShippingMethod string  `gorm:"size:30"`
	ShippingCost   float64 `gorm:"type:decimal(12,2)"`
	Notified       bool    `gorm:"not null;default:false"`
	// GORM timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

type OrderItem struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	OrderID   uuid.UUID  `gorm:"type:uuid;index"`
	ProductID *uuid.UUID `gorm:"type:uuid;index"`
	QuoteID   *uuid.UUID `gorm:"type:uuid;index"`
	Title     string     `gorm:"size:180"`
	Color     string     `gorm:"size:20"`
	Qty       int        `gorm:"not null"`
	UnitPrice float64    `gorm:"type:decimal(12,2)"`
}
