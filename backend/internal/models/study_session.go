package models

import (
	"time"

	"github.com/google/uuid"
)

// StudySession tracks the time a user spends studying a specific material.
type StudySession struct {
	ID              uint64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID          uuid.UUID  `json:"user_id" gorm:"not null;type:uuid"`
	MaterialID      uint64     `json:"material_id" gorm:"not null"`
	DurationMinutes int64      `json:"duration_minutes" gorm:"not null"`
	StartTime       time.Time  `json:"start_time" gorm:"not null"`
	EndTime         *time.Time `json:"end_time"`
	Status          string     `json:"status" gorm:"not null"`

	// Relationships
	User     User          `json:"-" gorm:"foreignKey:UserID"`
	Material StudyMaterial `json:"material" gorm:"foreignKey:MaterialID"`
	Quizzes  []Quiz        `json:"quizzes,omitempty" gorm:"foreignKey:SessionID"`
}

// TableName sets the custom table name for this model.
func (StudySession) TableName() string {
	return "study_sessions"
}
