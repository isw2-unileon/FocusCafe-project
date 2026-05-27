package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents the public.users table in Supabase.
// It is linked to auth.users via ID.
type User struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:uuid"`
	FirstName string    `json:"first_name" gorm:"not null"`
	LastName  string    `json:"last_name" gorm:"not null"`
	Username  string    `json:"username" gorm:"not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Role      string    `json:"role" gorm:"default:'user'"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`

	GroupID *int64 `json:"group_id,omitempty" gorm:"type:bigint"`

	// Relationships
	Progress *UserProgress `json:"progress,omitempty" gorm:"foreignKey:UserID"`
	Orders   []UserOrder   `json:"orders,omitempty" gorm:"foreignKey:UserID"`
	Group    *Group        `json:"group,omitempty" gorm:"foreignKey:GroupID;references:ID"`
}

// TableName overrides the default table name for the model.
func (User) TableName() string {
	return "users"
}
