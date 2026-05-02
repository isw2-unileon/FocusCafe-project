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

	// Relationships
	Progress *UserProgress `json:"progress,omitempty" gorm:"foreignKey:UserID"`
	Orders   []UserOrder   `json:"orders,omitempty" gorm:"foreignKey:UserID"`
}

// TableName overrides the default table name for the model.
func (User) TableName() string {
	return "users"
}
