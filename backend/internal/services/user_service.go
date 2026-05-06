package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
)

// UserServiceInterface defines the methods that the UserService must implement
type UserServiceInterface interface {
	GetUserProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
	UpdateUserProfile(ctx context.Context, id uuid.UUID, firstName, lastName string) error
	GetAllUsers(ctx context.Context) ([]domain.UserProfile, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.UserProfile, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

// UserRepository defines the interface for user-related data operations
type UserRepository interface {
	GetUserProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
	UpdateUserProfile(ctx context.Context, id uuid.UUID, firstName, lastName string) error
	GetAllUsers(ctx context.Context) ([]models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

// UserService provides methods to handle user-related business logic
type UserService struct {
	repo UserRepository
}

// NewUserService creates a new instance of UserService with the given UserRepository
func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetUserProfile retrieves the profile information of a user, including their gamified stats (energy, level, XP)
func (s *UserService) GetUserProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	return s.repo.GetUserProfile(ctx, id)
}

// UpdateUserProfile updates the user's first and last name after validating the input
func (s *UserService) UpdateUserProfile(ctx context.Context, id uuid.UUID, firstName, lastName string) error {
	if firstName == "" || lastName == "" {
		return errors.New("first name and last name are required")
	}
	return s.repo.UpdateUserProfile(ctx, id, firstName, lastName)
}

// GetAllUsers retrieves all users and maps them to domain profiles
func (s *UserService) GetAllUsers(ctx context.Context) ([]domain.UserProfile, error) {
	users, err := s.repo.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	profiles := make([]domain.UserProfile, 0, len(users))
	for _, u := range users {
		profile := domain.UserProfile{
			ID:        u.ID,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Username:  u.Username,
			Email:     u.Email,
			Role:      u.Role,
			Energy:    0,
			MaxEnergy: 500,
			XP:        0,
			Level:     1,
		}
		if u.Progress != nil {
			profile.Energy = u.Progress.Energy
			profile.XP = u.Progress.XP
			profile.Level = u.Progress.Level
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

// GetUserByEmail retrieves a user by email and maps it to a domain profile
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*domain.UserProfile, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	profile := &domain.UserProfile{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Energy:    0,
		MaxEnergy: 500,
		XP:        0,
		Level:     1,
	}
	if user.Progress != nil {
		profile.Energy = user.Progress.Energy
		profile.XP = user.Progress.XP
		profile.Level = user.Progress.Level
	}
	return profile, nil
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteUser(ctx, id)
}
