package domain

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Slug        string    `gorm:"uniqueIndex;size:140"`
	Name        string    `gorm:"size:180"`
	BasePrice   float64   `gorm:"type:decimal(12,2)"`
	Category    string    `gorm:"size:100"`
	ShortDesc   string    `gorm:"type:text"`
	ReadyToShip bool      `gorm:"default:true"`
	WidthMM     float64   `gorm:"type:decimal(8,2);default:0"`
	HeightMM    float64   `gorm:"type:decimal(8,2);default:0"`
	DepthMM     float64   `gorm:"type:decimal(8,2);default:0"`
	Observation string    `gorm:"type:text"`
	Grams       float64   `gorm:"type:decimal(8,2);default:0"`
	Hours       float64   `gorm:"type:decimal(8,2);default:0"`
	Profit      float64   `gorm:"type:decimal(12,2);default:0"`
	GrossPrice  float64   `gorm:"type:decimal(12,2);default:0"`
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

type FeaturedProduct struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	Product   Product   `gorm:"foreignKey:ProductID;references:ID"`
	Order     int       `gorm:"column:display_order;default:0"`
	Active    bool      `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
