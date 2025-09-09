package domain

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email     string    `gorm:"size:140;uniqueIndex"`
	Name      string    `gorm:"size:140"`
	Phone     string    `gorm:"size:60"`
	CreatedAt time.Time
}
