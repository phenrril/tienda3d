package domain

import (
	"time"

	"github.com/google/uuid"
)

type DiscountType string

const (
	DiscountTypePercentage  DiscountType = "percentage"
	DiscountTypeFixedAmount DiscountType = "fixed_amount"
)

type Coupon struct {
	ID                uuid.UUID    `gorm:"type:uuid;primaryKey"`
	Code              string       `gorm:"size:50;uniqueIndex;not null"`
	DiscountType      DiscountType `gorm:"type:varchar(20);not null"`
	DiscountValue     float64      `gorm:"type:decimal(12,2);not null"`
	MinPurchaseAmount float64      `gorm:"type:decimal(12,2);not null;default:0"`
	MaxUses           *int         `gorm:"type:int"`
	CurrentUses       int          `gorm:"type:int;not null;default:0"`
	ExpiresAt         *time.Time   `gorm:"type:timestamp"`
	Active            bool         `gorm:"not null;default:true;index"`
	Description       string       `gorm:"size:255"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

type CouponUsage struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	CouponID        uuid.UUID `gorm:"type:uuid;index;not null"`
	OrderID         uuid.UUID `gorm:"type:uuid;index;not null"`
	Email           string    `gorm:"size:140;index"`
	DiscountApplied float64   `gorm:"type:decimal(12,2);not null"`
	OrderTotal      float64   `gorm:"type:decimal(12,2);not null"`
	UsedAt          time.Time `gorm:"not null"`

	Coupon *Coupon `gorm:"foreignKey:CouponID"`
	Order  *Order  `gorm:"foreignKey:OrderID"`
}
