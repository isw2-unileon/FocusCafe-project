package models

import (
	"time"
)

// Quiz represents a test generated after a study session.
type Quiz struct {
	ID          uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	SessionID   uint64    `json:"session_id" gorm:"not null"`
	GeneratedAt time.Time `json:"generated_at" gorm:"default:now()"`

	// Relationships
	Session   StudySession `json:"-" gorm:"foreignKey:SessionID"`
	Questions []Question   `json:"questions,omitempty" gorm:"foreignKey:QuizID"`
}

// TableName sets the custom table name for this model.
func (Quiz) TableName() string {
	return "quizzes"
}
