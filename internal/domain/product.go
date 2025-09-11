package domain

import (
	"time"

	"github.com/google/uuid"
)

// Product representa un modelo listo para imprimir o un ítem fabricable bajo demanda.
type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Slug        string    `gorm:"uniqueIndex;size:140"`
	Name        string    `gorm:"size:180"`           // constraint NOT NULL se aplicará manualmente tras backfill
	BasePrice   float64   `gorm:"type:decimal(12,2)"` // idem
	Category    string    `gorm:"size:100"`
	ShortDesc   string    `gorm:"type:text"`
	ReadyToShip bool      `gorm:"default:true"`
	WidthMM     float64   `gorm:"type:decimal(8,2);default:0"`
	HeightMM    float64   `gorm:"type:decimal(8,2);default:0"`
	DepthMM     float64   `gorm:"type:decimal(8,2);default:0"`
	Images      []Image
	Variants    []Variant
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Variant struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductID     uuid.UUID `gorm:"type:uuid;index"`
	Material      Material  `gorm:"type:varchar(10);not null"`
	Color         string    `gorm:"size:60"`
	LayerHeightMM float64   `gorm:"type:decimal(4,2)"`
	InfillPct     int       `gorm:"type:int"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Image struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductID uuid.UUID `gorm:"type:uuid;index"`
	URL       string    `gorm:"size:255"`
	Alt       string    `gorm:"size:140"`
	CreatedAt time.Time
}
