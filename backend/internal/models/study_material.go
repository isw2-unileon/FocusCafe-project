package models

import (
	"time"

	"github.com/google/uuid"
)

// StudyMaterial represents documents uploaded by a user for their studies.
type StudyMaterial struct {
	ID          uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      uuid.UUID `json:"user_id" gorm:"not null;type:uuid"`
	Title       string    `json:"title" gorm:"not null"`
	SubjectName string    `json:"subject_name" gorm:"not null"`
	FilePath    string    `json:"file_path" gorm:"not null"`
	UploadDate  time.Time `json:"upload_date" gorm:"autoCreateTime"`
	Content     string    `json:"content"`

	// Relationships
	User User `json:"-" gorm:"foreignKey:UserID"`
	Sessions []StudySession `json:"sessions,omitempty" gorm:"foreignKey:MaterialID"`
}

// TableName sets the custom table name for this model.
func (StudyMaterial) TableName() string {
	return "study_materials"
}
