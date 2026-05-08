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
		LastName:  m.LastName,
		Username:  m.Username,
		Email:     m.Email,
		Role:      m.Role,
		CreatedAt: m.CreatedAt.Format("2006-01-02"),
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

// UpdateUserProfile updates the user's first and last name in the database
func (r *UserRepository) UpdateUserProfile(ctx context.Context, id uuid.UUID, firstName, lastName string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Updates(models.User{
		FirstName: firstName,
		LastName:  lastName,
	}).Error
}

// GetAllUsers retrieves all users from the database with their progress
func (r *UserRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).Preload("Progress").Find(&users).Error
	return users, err
}

// GetUserByEmail retrieves a user by their email address
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Preload("Progress").Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// DeleteUser performs a hard delete of the user by ID (removes from database permanently)
func (r *UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&models.User{}, id).Error
}

// CreateUser inserts a new user into the database
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}
