package models

import (
	"time"

	"github.com/google/uuid"
)

// UserOrder represents a specific order placed by a user.
type UserOrder struct {
	ID          uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      uuid.UUID `json:"user_id" gorm:"not null;type:uuid"` // Changed to UUID to match users table
	CafeOrderID uint64    `json:"cafe_order_id" gorm:"not null"`
	Status      string    `json:"status" gorm:"default:'pending'"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	User      User      `json:"-" gorm:"foreignKey:UserID"`
	CafeOrder CafeOrder `json:"cafe_order" gorm:"foreignKey:CafeOrderID"`
}

// TableName overrides the default table name for the model.
func (UserOrder) TableName() string {
	return "user_orders"
}
