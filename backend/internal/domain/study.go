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
	Options       []string // ["opcion1", "opcion2"...]
	CorrectAnswer int      // 0 para A, 1 para B, 2 para C, 3 para D
	Explanation   string
}
