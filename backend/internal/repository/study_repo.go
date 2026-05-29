package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"gorm.io/gorm"
)

// StudyRepositoryInterface defines the methods that the StudyRepository must implement
type StudyRepositoryInterface interface {
	CreateMaterial(ctx context.Context, material *models.StudyMaterial) error
	CreateSession(ctx context.Context, session *models.StudySession) error
	GetSessionWithMaterial(ctx context.Context, sessionID uint64) (*models.StudySession, error)
	SaveFullQuiz(ctx context.Context, sessionID uint64, quizName string, questions []models.Question) error
	UpdateUserProgress(ctx context.Context, userID uuid.UUID, sessionID uint64, energy int) (int, error)
}

// StudyRepository provides methods to interact with the study-related database tables
type StudyRepository struct {
	db *gorm.DB
}

// NewStudyRepository creates a new instance of StudyRepository with the given gorm DB connection
func NewStudyRepository(db *gorm.DB) *StudyRepository {
	return &StudyRepository{db: db}
}

// CreateMaterial inserts a new study material into the database and returns any error encountered
func (r *StudyRepository) CreateMaterial(ctx context.Context, material *models.StudyMaterial) error {
	return r.db.WithContext(ctx).Create(material).Error
}

// CreateSession inserts a new study session into the database and returns any error encountered
func (r *StudyRepository) CreateSession(ctx context.Context, session *models.StudySession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetSessionWithMaterial retrieves a study session along with its associated material based on the session ID and returns the session or any error encountered
func (r *StudyRepository) GetSessionWithMaterial(ctx context.Context, sessionID uint64) (*models.StudySession, error) {
	var session models.StudySession
	if err := r.db.WithContext(ctx).Preload("Material").Where("id = ?", sessionID).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// SaveFullQuiz inserts a new quiz and its associated questions into the database within a transaction and returns any error encountered
func (r *StudyRepository) SaveFullQuiz(ctx context.Context, sessionID uint64, quizName string, questions []models.Question) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		quizModel := models.Quiz{
			SessionID: sessionID,
		}
		if err := tx.Create(&quizModel).Error; err != nil {
			return err
		}

		for i := range questions {
			questions[i].QuizID = quizModel.ID
			if err := tx.Create(&questions[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateUserProgress updates a user's progress in the database by adding energy points and returns the updated energy or any error encountered
func (r *StudyRepository) UpdateUserProgress(ctx context.Context, userID uuid.UUID, sessionID uint64, energy int) (int, error) {
	var session models.StudySession
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error; err != nil {
		return 0, errors.New("study session not found")
	}

	var progress models.UserProgress
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).First(&progress).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				progress = models.UserProgress{UserID: userID, Energy: 0, Level: 1, XP: 0}
				if err := tx.Create(&progress).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		progress.Energy += energy
		return tx.Save(&progress).Error
	})

	return progress.Energy, err
}
