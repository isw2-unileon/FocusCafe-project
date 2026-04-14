package models

import (
	"github.com/google/uuid"
)

// UserProgress represents the public.user_progress table.
// It stores the gamified stats of a user.
type UserProgress struct {
	UserID uuid.UUID `json:"user_id" gorm:"primaryKey;type:uuid"`
	Energy int       `json:"energy" gorm:"default:0"`
	Level  int       `json:"level" gorm:"default:1"`

	// Relationship
	User User `json:"-" gorm:"foreignKey:UserID"`
}

func (UserProgress) TableName() string {
	return "user_progress"
}
