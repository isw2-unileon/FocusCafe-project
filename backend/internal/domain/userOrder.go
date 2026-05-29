package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserOrder represents a specific order made to a user or a group.
type UserOrder struct {
	ID          uint64    `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	CafeOrderID uint64    `json:"cafe_order_id"`
	Status      string    `json:"status"` // "pending", "completed"
	CreatedAt   time.Time `json:"created_at"`
	GroupID     *int64    `json:"group_id,omitempty"`

	CafeOrder *CafeOrder `json:"cafe_order,omitempty"`
}
