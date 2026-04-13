package models

import (
	"gorm.io/gorm"
)

// UserOrder represents a specific order placed by a user.
type UserOrder struct {
	gorm.Model
	UserID  uint   `json:"user_id"`
	OrderID uint   `json:"order_id"`
	Status  string `json:"status" gorm:"default:'pending'"` // "pending" | "completed"

	// Relations
	User  User      `json:"-" gorm:"foreignKey:UserID"`
	Order CafeOrder `json:"order" gorm:"foreignKey:OrderID"`
}
