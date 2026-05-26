package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
)

// UserServiceInterface defines the methods that the UserService must implement
//
//nolint:dupl
type UserServiceInterface interface {
	GetUserProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
	UpdateUserProfile(ctx context.Context, id uuid.UUID, firstName, lastName string) error
	GetAllUsers(ctx context.Context) ([]domain.UserProfile, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.UserProfile, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

// UserRepository defines the interface for user-related data operations
//
//nolint:dupl
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

// mapUserToProfile maps a models.User to domain.UserProfile
func mapUserToProfile(u models.User) domain.UserProfile {
	profile := domain.UserProfile{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt.Format("2006-01-02"),
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
	if u.Group != nil {
		profile.Group = &domain.Group{
			ID:         u.Group.ID,
			Name:       u.Group.Name,
			InviteCode: u.Group.InviteCode,
			LeaderID:   u.Group.LeaderID,
			CreatedAt:  u.Group.CreatedAt,
		}
	}
	return profile
}

// GetAllUsers retrieves all users and maps them to domain profiles
func (s *UserService) GetAllUsers(ctx context.Context) ([]domain.UserProfile, error) {
	users, err := s.repo.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	profiles := make([]domain.UserProfile, 0, len(users))
	for _, u := range users {
		profiles = append(profiles, mapUserToProfile(u))
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

	profile := mapUserToProfile(*user)
	return &profile, nil
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteUser(ctx, id)
}
