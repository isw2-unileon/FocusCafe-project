package domain

import (
	"time"

	"github.com/google/uuid"
)

// GroupMember represents a user belonging to a group.
type GroupMember struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Level     int       `json:"level"`
}

// GroupDetail represents a group with its members for admin viewing.
type GroupDetail struct {
	ID         int64         `json:"id"`
	Name       string        `json:"name"`
	InviteCode string        `json:"invite_code"`
	LeaderID   uuid.UUID     `json:"leader_id"`
	CreatedAt  time.Time     `json:"created_at"`
	Members    []GroupMember `json:"members"`
}
