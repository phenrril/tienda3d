package domain

import (
    "context"
    "time"

    "github.com/google/uuid"
)

// ExpenseType clasifica el movimiento de caja
type ExpenseType string

const (
    ExpenseTypeOut ExpenseType = "egreso"
    ExpenseTypeIn  ExpenseType = "ingreso"
)

// Expense representa un movimiento manual (egreso o ingreso fuera de la web)
type Expense struct {
    ID            uuid.UUID   `gorm:"type:uuid;primaryKey"`
    Kind          ExpenseType `gorm:"type:varchar(16);index"`
    Date          time.Time   `gorm:"index"`
    Category      string      `gorm:"size:80;index"`
    Description   string      `gorm:"size:255"`
    Amount        float64     `gorm:"type:decimal(12,2)"`
    PaymentMethod string      `gorm:"size:30"`

    // Datos opcionales del cliente para ingresos manuales
    CustomerName  string `gorm:"size:140"`
    CustomerEmail string `gorm:"size:140"`
    CustomerPhone string `gorm:"size:60"`

    CreatedAt time.Time
    UpdatedAt time.Time
}

type ExpenseRepo interface {
    Save(ctx context.Context, e *Expense) error
    Recent(ctx context.Context, limit int) ([]Expense, error)
}
