package models

import (
	"time"

	"github.com/google/uuid"
)

// Group represents the public.groups table in Supabase.
type Group struct {
	ID         int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	Name       string    `json:"name" gorm:"not null"`
	InviteCode string    `json:"invite_code" gorm:"unique;not null"`
	LeaderID   uuid.UUID `json:"leader_id" gorm:"not null;type:uuid"`

	// Relationships
	Leader *User  `json:"leader,omitempty" gorm:"foreignKey:LeaderID;references:ID"`
	Users  []User `json:"users,omitempty" gorm:"foreignKey:GroupID"`
}

// TableName overrides the default table name for the model.
func (Group) TableName() string {
	return "groups"
}
