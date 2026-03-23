package domain

import (
	"time"

	"github.com/google/uuid"
)

type WorkshopOrderStatus string

const (
	WorkshopPendiente    WorkshopOrderStatus = "pendiente"
	WorkshopDisenado     WorkshopOrderStatus = "disenado"
	WorkshopEnImpresion  WorkshopOrderStatus = "en_impresion"
	WorkshopListoEntrega WorkshopOrderStatus = "listo_entrega"
	WorkshopEntregado    WorkshopOrderStatus = "entregado"
)

type WorkshopOrder struct {
	ID           uuid.UUID           `gorm:"type:uuid;primaryKey"`
	ClientSlug   string              `gorm:"size:120;index"`
	RequestedAt  time.Time           `gorm:"type:date"`
	DeliveryDate time.Time           `gorm:"type:date"`
	Detail       string              `gorm:"type:text"`
	TotalAmount  *float64            `gorm:"type:decimal(12,2)"`
	IsBarter     bool                `gorm:"not null;default:false"`
	Status       WorkshopOrderStatus `gorm:"size:30;index"`
	DeliveredAt  *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time

	Deposits  []WorkshopDeposit       `gorm:"foreignKey:WorkshopOrderID"`
	Filaments []WorkshopOrderFilament `gorm:"foreignKey:WorkshopOrderID"`
}

func (WorkshopOrder) TableName() string { return "workshop_orders" }

type WorkshopDeposit struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	WorkshopOrderID uuid.UUID `gorm:"type:uuid;index"`
	Amount          float64   `gorm:"type:decimal(12,2);not null"`
	PaidAt          time.Time `gorm:"type:date"`
	CreatedAt       time.Time
}

func (WorkshopDeposit) TableName() string { return "workshop_deposits" }

type WorkshopOrderFilament struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	WorkshopOrderID uuid.UUID `gorm:"type:uuid;index"`
	ColorSlug       string    `gorm:"size:80;not null"`
	Grams           int       `gorm:"not null"`
	CreatedAt       time.Time
}

func (WorkshopOrderFilament) TableName() string { return "workshop_order_filaments" }

const (
	FilamentEntryPurchase    = "purchase"
	FilamentEntryConsumption = "consumption"
)

type FilamentLedgerEntry struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ColorSlug          string     `gorm:"size:80;index"`
	DeltaGrams         int        `gorm:"not null"`
	EntryType          string     `gorm:"size:20;index"`
	UnitCost           *float64   `gorm:"type:decimal(12,2)"`
	RefWorkshopOrderID *uuid.UUID `gorm:"type:uuid;index"`
	Note               string     `gorm:"size:255"`
	CreatedAt          time.Time
}

func (FilamentLedgerEntry) TableName() string { return "filament_ledger_entries" }

type BusinessExpense struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Amount      float64   `gorm:"type:decimal(12,2);not null"`
	SpentAt     time.Time `gorm:"type:date"`
	Category    string    `gorm:"size:80"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time
}

func (BusinessExpense) TableName() string { return "business_expenses" }

type AppSetting struct {
	Key   string `gorm:"size:64;primaryKey"`
	Value string `gorm:"type:text"`
}

func (AppSetting) TableName() string { return "app_settings" }

const SettingWorkshopDigestLast = "workshop_digest_last_date"
