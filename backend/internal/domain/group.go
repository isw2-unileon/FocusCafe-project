package domain

import (
	"time"

	"github.com/google/uuid"
)

// Group represents a study/cafe group in the system.
type Group struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	InviteCode string    `json:"invite_code"`
	LeaderID   uuid.UUID `json:"leader_id"`
	CreatedAt  time.Time `json:"created_at"`
}
