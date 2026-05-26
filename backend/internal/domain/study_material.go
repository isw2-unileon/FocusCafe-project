package domain

import (
	"time"

	"github.com/google/uuid"
)

// StudyMaterial represents the study material uploaded by a user, including its metadata such as title, subject, file path, content, and upload date.
type StudyMaterial struct {
	ID          uint64    `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	SubjectName string    `json:"subject_name"`
	FilePath    string    `json:"file_path"`
	Content     string    `json:"content"`
	UploadDate  time.Time `json:"upload_date"`
}
