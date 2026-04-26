package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"gorm.io/gorm"
)

// UserRepository provides methods to interact with the database for user-related operations
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository with the given database connection
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetUserProfile retrieves the profile information of a user, including their gamified stats (energy, level, XP)
func (r *UserRepository) GetUserProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	var m models.User

	// Query the database for the user with the given ID
	err := r.db.WithContext(ctx).Preload("Progress").First(&m, id).Error
	if err != nil {
		return nil, err
	}

	// Map the database model to the domain model
	profile := &domain.UserProfile{
		ID:        m.ID,
		FirstName: m.FirstName,
		Username:  m.Username,
		Energy:    0,
		MaxEnergy: 500,
		XP:        0,
		Level:     1,
	}

	if m.Progress != nil {
		profile.Energy = m.Progress.Energy
		profile.XP = m.Progress.XP
		profile.Level = m.Progress.Level
	}

	return profile, nil
}
