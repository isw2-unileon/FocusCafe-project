package domain

import (
	"time"

	"github.com/google/uuid"
)

// StudySession represents a user's study session, including the associated study material, session status, and timing information.
type StudySession struct {
	ID              uint64
	UserID          uuid.UUID
	MaterialID      uint64
	Status          string
	StartTime       time.Time
	EndTime         *time.Time
	DurationMinutes int64
}

// QuizQuestion represents a single quiz question with its text, options, correct answer index, and an explanation for the correct answer.
type QuizQuestion struct {
	Question      string
	Options       []string // Expected to have exactly 4 options (A, B, C, D)
	CorrectAnswer int      // Index of the correct answer in the Options slice (0-3)
	Explanation   string
}
