package models

import (
	"time"

	"github.com/google/uuid"
)

// UserOrder represents a specific order placed by a user or a group.
type UserOrder struct {
	ID          uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      uuid.UUID `json:"user_id" gorm:"type:uuid"` // nullable for group orders
	CafeOrderID uint64    `json:"cafe_order_id" gorm:"not null"`
	Status      string    `json:"status" gorm:"default:'pending'"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	GroupID     *int64    `json:"group_id" gorm:"type:bigint"` // nullable; set for group orders

	// Relationships
	User      User       `json:"-" gorm:"foreignKey:UserID"`
	CafeOrder CafeOrder  `json:"cafe_order" gorm:"foreignKey:CafeOrderID"`
	Group     *Group     `json:"group,omitempty" gorm:"foreignKey:GroupID;references:ID"`
}

// TableName overrides the default table name for the model.
func (UserOrder) TableName() string {
	return "user_orders"
}
