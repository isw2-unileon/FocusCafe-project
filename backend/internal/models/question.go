package models

// Question represents a single question within a Quiz.
type Question struct {
	ID            uint64 `json:"id" gorm:"primaryKey;autoIncrement"`
	QuizID        uint64 `json:"quiz_id" gorm:"not null"`
	QuestionText  string `json:"question_text" gorm:"not null"`
	OptionA       string `json:"option_a" gorm:"not null"`
	OptionB       string `json:"option_b" gorm:"not null"`
	OptionC       string `json:"option_c" gorm:"not null"`
	OptionD       string `json:"option_d" gorm:"not null"`
	CorrectAnswer string `json:"correct_answer" gorm:"not null;type:char(1)"`
	Explanation   string `json:"explanation"`

	// Relationship
	Quiz Quiz `json:"-" gorm:"foreignKey:QuizID"`
}

// TableName sets the custom table name for this model.
func (Question) TableName() string {
	return "questions"
}
