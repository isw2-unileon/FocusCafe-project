package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
)

// mockUserRepository is a test double for the UserRepository.
type mockUserRepository struct {
	getUserProfileFunc     func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
	updateUserProfileFunc  func(ctx context.Context, id uuid.UUID, firstName, lastName string) error
	getAllUsersFunc        func(ctx context.Context) ([]models.User, error)
	getUserByEmailFunc     func(ctx context.Context, email string) (*models.User, error)
	deleteUserFunc         func(ctx context.Context, id uuid.UUID) error
	getLeaderboardFunc     func(ctx context.Context, limit int) ([]models.User, error)
	getUserRankFunc        func(ctx context.Context, userID uuid.UUID) (int, error)
}

func (m *mockUserRepository) GetUserProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	if m.getUserProfileFunc != nil {
		return m.getUserProfileFunc(ctx, id)
	}
	return nil, errors.New("getUserProfile not mocked")
}

func (m *mockUserRepository) UpdateUserProfile(ctx context.Context, id uuid.UUID, firstName, lastName string) error {
	if m.updateUserProfileFunc != nil {
		return m.updateUserProfileFunc(ctx, id, firstName, lastName)
	}
	return errors.New("updateUserProfile not mocked")
}

func (m *mockUserRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	if m.getAllUsersFunc != nil {
		return m.getAllUsersFunc(ctx)
	}
	return nil, errors.New("getAllUsers not mocked")
}

func (m *mockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, errors.New("getUserByEmail not mocked")
}

func (m *mockUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if m.deleteUserFunc != nil {
		return m.deleteUserFunc(ctx, id)
	}
	return errors.New("deleteUser not mocked")
}

func (m *mockUserRepository) GetLeaderboard(ctx context.Context, limit int) ([]models.User, error) {
	if m.getLeaderboardFunc != nil {
		return m.getLeaderboardFunc(ctx, limit)
	}
	return nil, errors.New("getLeaderboard not mocked")
}

func (m *mockUserRepository) GetUserRank(ctx context.Context, userID uuid.UUID) (int, error) {
	if m.getUserRankFunc != nil {
		return m.getUserRankFunc(ctx, userID)
	}
	return 0, errors.New("getUserRank not mocked")
}

// ============================================
// TestUserService_GetLeaderboard
// ============================================

func TestUserService_GetLeaderboard(t *testing.T) {
	testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name         string
		mockBehavior func(ctx context.Context, limit int) ([]models.User, error)
		wantErr      bool
		wantLen      int
		wantFirstXP  int
	}{
		{
			name: "Success: Returns mapped profiles ordered by XP",
			mockBehavior: func(ctx context.Context, limit int) ([]models.User, error) {
				return []models.User{
					{
						ID:        testUUID,
						FirstName: "Alice",
						Username:  "alice",
						Email:     "alice@test.com",
						Role:      "user",
						Progress:  &models.UserProgress{UserID: testUUID, XP: 1500, Level: 5, Energy: 300},
					},
					{
						ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
						FirstName: "Bob",
						Username:  "bob",
						Email:     "bob@test.com",
						Role:      "user",
						Progress:  &models.UserProgress{UserID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), XP: 800, Level: 3, Energy: 200},
					},
				}, nil
			},
			wantErr:     false,
			wantLen:     2,
			wantFirstXP: 1500,
		},
		{
			name: "Success: Returns empty leaderboard",
			mockBehavior: func(ctx context.Context, limit int) ([]models.User, error) {
				return []models.User{}, nil
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "Error: Repository failure",
			mockBehavior: func(ctx context.Context, limit int) ([]models.User, error) {
				return nil, errors.New("database connection lost")
			},
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockUserRepository{getLeaderboardFunc: tt.mockBehavior}
			s := services.NewUserService(mRepo)

			profiles, err := s.GetLeaderboard(context.Background(), 5)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLeaderboard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(profiles) != tt.wantLen {
				t.Errorf("GetLeaderboard() len = %d, want %d", len(profiles), tt.wantLen)
			}
			if !tt.wantErr && tt.wantLen > 0 && profiles[0].XP != tt.wantFirstXP {
				t.Errorf("GetLeaderboard() first XP = %d, want %d", profiles[0].XP, tt.wantFirstXP)
			}
		})
	}
}

// ============================================
// TestUserService_GetUserLeaderboard
// ============================================

func TestUserService_GetUserLeaderboard(t *testing.T) {
	testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")

	tests := []struct {
		name               string
		mockProfileFunc    func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
		mockRankFunc       func(ctx context.Context, userID uuid.UUID) (int, error)
		wantErr            bool
		wantRank           int
		wantFirstName      string
	}{
		{
			name: "Success: Returns rank 7 and profile for regular user",
			mockProfileFunc: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
				return &domain.UserProfile{
					ID:        id,
					FirstName: "Juan",
					XP:        800,
					Level:     5,
				}, nil
			},
			mockRankFunc: func(ctx context.Context, userID uuid.UUID) (int, error) {
				return 7, nil
			},
			wantErr:       false,
			wantRank:      7,
			wantFirstName: "Juan",
		},
		{
			name: "Success: Returns rank 1 for top player",
			mockProfileFunc: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
				return &domain.UserProfile{
					ID:        id,
					FirstName: "Alice",
					XP:        3200,
					Level:     15,
				}, nil
			},
			mockRankFunc: func(ctx context.Context, userID uuid.UUID) (int, error) {
				return 1, nil
			},
			wantErr:       false,
			wantRank:      1,
			wantFirstName: "Alice",
		},
		{
			name: "Error: GetUserProfile fails",
			mockProfileFunc: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
				return nil, errors.New("user not found")
			},
			mockRankFunc: func(ctx context.Context, userID uuid.UUID) (int, error) {
				return 0, nil
			},
			wantErr: true,
		},
		{
			name: "Error: GetUserRank fails",
			mockProfileFunc: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
				return &domain.UserProfile{ID: id, FirstName: "Juan"}, nil
			},
			mockRankFunc: func(ctx context.Context, userID uuid.UUID) (int, error) {
				return 0, errors.New("database connection lost")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockUserRepository{
				getUserProfileFunc: tt.mockProfileFunc,
				getUserRankFunc:    tt.mockRankFunc,
			}
			s := services.NewUserService(mRepo)

			rank, profile, err := s.GetUserLeaderboard(context.Background(), testUUID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserLeaderboard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if rank != tt.wantRank {
				t.Errorf("GetUserLeaderboard() rank = %d, want %d", rank, tt.wantRank)
			}
			if profile.FirstName != tt.wantFirstName {
				t.Errorf("GetUserLeaderboard() firstName = %v, want %v", profile.FirstName, tt.wantFirstName)
			}
		})
	}
}
