package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserOrder represents a specific order made to a user
type UserOrder struct {
	ID          uint64    `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	CafeOrderID uint64    `json:"cafe_order_id"`
	Status      string    `json:"status"` // "pending", "completed"
	CreatedAt   time.Time `json:"created_at"`

	CafeOrder *CafeOrder `json:"cafe_order,omitempty"`
}
