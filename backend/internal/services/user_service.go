package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
)

// UserServiceInterface defines the methods that the UserService must implement
type UserServiceInterface interface {
	GetUserProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
	UpdateUserProfile(ctx context.Context, id uuid.UUID, firstName, lastName string) error
}

// UserRepository defines the interface for user-related data operations
type UserRepository interface {
	GetUserProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
	UpdateUserProfile(ctx context.Context, id uuid.UUID, firstName, lastName string) error
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
