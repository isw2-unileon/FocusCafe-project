package services_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
)

type mockUserRepository struct {
	getUserProfileFunc    func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
	updateUserProfileFunc func(ctx context.Context, id uuid.UUID, firstName, lastName string) error
	getAllUsersFunc       func(ctx context.Context) ([]models.User, error)
	getUserByEmailFunc    func(ctx context.Context, email string) (*models.User, error)
	deleteUserFunc        func(ctx context.Context, id uuid.UUID) error
	getLeaderboardFunc    func(ctx context.Context, limit int) ([]models.User, error)
	getUserRankFunc       func(ctx context.Context, userID uuid.UUID) (int, error)
}

func (m *mockUserRepository) GetUserProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	if m.getUserProfileFunc != nil {
		return m.getUserProfileFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepository) UpdateUserProfile(ctx context.Context, id uuid.UUID, firstName, lastName string) error {
	if m.updateUserProfileFunc != nil {
		return m.updateUserProfileFunc(ctx, id, firstName, lastName)
	}
	return nil
}

func (m *mockUserRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	if m.getAllUsersFunc != nil {
		return m.getAllUsersFunc(ctx)
	}
	return nil, nil
}

func (m *mockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if m.deleteUserFunc != nil {
		return m.deleteUserFunc(ctx, id)
	}
	return nil
}

func (m *mockUserRepository) GetLeaderboard(ctx context.Context, limit int) ([]models.User, error) {
	if m.getLeaderboardFunc != nil {
		return m.getLeaderboardFunc(ctx, limit)
	}
	return nil, nil
}

func (m *mockUserRepository) GetUserRank(ctx context.Context, userID uuid.UUID) (int, error) {
	if m.getUserRankFunc != nil {
		return m.getUserRankFunc(ctx, userID)
	}
	return 0, nil
}

func TestUserService_GetUserProfile(t *testing.T) {
	id := uuid.New()
	mockProfile := &domain.UserProfile{ID: id, FirstName: "Juan", Email: "juan@test.com"}

	tests := []struct {
		name    string
		repo    services.UserRepository
		wantErr bool
		want    *domain.UserProfile
	}{
		{
			name: "Success: Profile retrieved",
			repo: &mockUserRepository{
				getUserProfileFunc: func(ctx context.Context, uid uuid.UUID) (*domain.UserProfile, error) {
					return mockProfile, nil
				},
			},
			wantErr: false,
			want:    mockProfile,
		},
		{
			name: "Error: Repo failure",
			repo: &mockUserRepository{
				getUserProfileFunc: func(ctx context.Context, uid uuid.UUID) (*domain.UserProfile, error) {
					return nil, errors.New("not found")
				},
			},
			wantErr: true,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := services.NewUserService(tt.repo)
			got, err := s.GetUserProfile(context.Background(), id)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetUserProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserProfile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserService_UpdateUserProfile(t *testing.T) {
	id := uuid.New()

	tests := []struct {
		name      string
		firstName string
		lastName  string
		repo      services.UserRepository
		wantErr   bool
	}{
		{
			name:      "Success: Valid inputs forwarded to repo",
			firstName: "Carlos",
			lastName:  "Santana",
			repo: &mockUserRepository{
				updateUserProfileFunc: func(ctx context.Context, uid uuid.UUID, f, l string) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name:      "Error: Missing first name",
			firstName: "",
			lastName:  "Santana",
			repo:      &mockUserRepository{},
			wantErr:   true,
		},
		{
			name:      "Error: Missing last name",
			firstName: "Carlos",
			lastName:  "",
			repo:      &mockUserRepository{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := services.NewUserService(tt.repo)
			err := s.UpdateUserProfile(context.Background(), id, tt.firstName, tt.lastName)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateUserProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserService_GetAllUsers(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	now := time.Now()

	mockDBUsers := []models.User{
		{
			ID:        id1,
			FirstName: "User1",
			CreatedAt: now,
			Progress:  &models.UserProgress{Energy: 200, XP: 50, Level: 2},
		},
		{
			ID:        id2,
			FirstName: "User2",
			CreatedAt: now,
			Group:     &models.Group{ID: 10, Name: "Alfa Team", InviteCode: "ALFA"},
		},
	}

	expectedProfiles := []domain.UserProfile{
		{
			ID:        id1,
			FirstName: "User1",
			CreatedAt: now.Format("2006-01-02"),
			Energy:    200,
			MaxEnergy: 500,
			XP:        50,
			Level:     2,
		},
		{
			ID:        id2,
			FirstName: "User2",
			CreatedAt: now.Format("2006-01-02"),
			Energy:    0,
			MaxEnergy: 500,
			XP:        0,
			Level:     1,
			Group:     &domain.Group{ID: 10, Name: "Alfa Team", InviteCode: "ALFA"},
		},
	}

	tests := []struct {
		name    string
		repo    services.UserRepository
		wantErr bool
		want    []domain.UserProfile
	}{
		{
			name: "Success: Retrieve and map multiple users",
			repo: &mockUserRepository{
				getAllUsersFunc: func(ctx context.Context) ([]models.User, error) {
					return mockDBUsers, nil
				},
			},
			wantErr: false,
			want:    expectedProfiles,
		},
		{
			name: "Error: Repo failure",
			repo: &mockUserRepository{
				getAllUsersFunc: func(ctx context.Context) ([]models.User, error) {
					return nil, errors.New("db error")
				},
			},
			wantErr: true,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := services.NewUserService(tt.repo)
			got, err := s.GetAllUsers(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetAllUsers() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserService_GetUserByEmail(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	mockDBUser := &models.User{
		ID:        id,
		Email:     "test@focus.com",
		CreatedAt: now,
	}

	expectedProfile := &domain.UserProfile{
		ID:        id,
		Email:     "test@focus.com",
		CreatedAt: now.Format("2006-01-02"),
		MaxEnergy: 500,
		Level:     1,
	}

	tests := []struct {
		name    string
		email   string
		repo    services.UserRepository
		wantErr bool
		want    *domain.UserProfile
	}{
		{
			name:  "Success: User found by email",
			email: "test@focus.com",
			repo: &mockUserRepository{
				getUserByEmailFunc: func(ctx context.Context, e string) (*models.User, error) {
					return mockDBUser, nil
				},
			},
			wantErr: false,
			want:    expectedProfile,
		},
		{
			name:    "Error: Email parameter is empty",
			email:   "",
			repo:    &mockUserRepository{},
			wantErr: true,
			want:    nil,
		},
		{
			name:  "Error: Repo returns error",
			email: "missing@test.com",
			repo: &mockUserRepository{
				getUserByEmailFunc: func(ctx context.Context, e string) (*models.User, error) {
					return nil, errors.New("record not found")
				},
			},
			wantErr: true,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := services.NewUserService(tt.repo)
			got, err := s.GetUserByEmail(context.Background(), tt.email)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetUserByEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserByEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	id := uuid.New()

	t.Run("Success: Deletion call forwarded", func(t *testing.T) {
		repo := &mockUserRepository{
			deleteUserFunc: func(ctx context.Context, uid uuid.UUID) error {
				return nil
			},
		}
		s := services.NewUserService(repo)
		err := s.DeleteUser(context.Background(), id)
		if err != nil {
			t.Errorf("unexpected error on DeleteUser: %v", err)
		}
	})
}

func TestGetLeaderboard_Success(t *testing.T) {
	repo := &mockUserRepository{
		getLeaderboardFunc: func(ctx context.Context, limit int) ([]models.User, error) {
			return []models.User{
				{ID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), FirstName: "Alice", Email: "alice@test.com", Progress: &models.UserProgress{XP: 3200, Level: 15}},
				{ID: uuid.MustParse("22222222-2222-2222-2222-222222222222"), FirstName: "Bob", Email: "bob@test.com", Progress: &models.UserProgress{XP: 2100, Level: 12}},
			}, nil
		},
	}
	s := services.NewUserService(repo)

	result, err := s.GetLeaderboard(context.Background(), 5)
	if err != nil {
		t.Fatalf("Unexpected error fetching leaderboard: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 leaderboard entries, got %d", len(result))
	}

	if result[0].FirstName != "Alice" {
		t.Errorf("Expected first entry to be Alice, got %s", result[0].FirstName)
	}

	if result[0].XP != 3200 {
		t.Errorf("Expected Alice to have 3200 XP, got %d", result[0].XP)
	}

	if result[1].FirstName != "Bob" {
		t.Errorf("Expected second entry to be Bob, got %s", result[1].FirstName)
	}
}

func TestGetLeaderboard_RepoError(t *testing.T) {
	expectedErr := errors.New("leaderboard query timeout")
	repo := &mockUserRepository{
		getLeaderboardFunc: func(ctx context.Context, limit int) ([]models.User, error) {
			return nil, expectedErr
		},
	}
	s := services.NewUserService(repo)

	_, err := s.GetLeaderboard(context.Background(), 5)
	if err == nil {
		t.Fatal("Expected error when repository fails, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestGetUserLeaderboard_Success(t *testing.T) {
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	repo := &mockUserRepository{
		getUserProfileFunc: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
			return &domain.UserProfile{ID: userID, FirstName: "Juan", XP: 800, Level: 5}, nil
		},
		getUserRankFunc: func(ctx context.Context, uid uuid.UUID) (int, error) {
			return 7, nil
		},
	}
	s := services.NewUserService(repo)

	rank, profile, err := s.GetUserLeaderboard(context.Background(), userID)
	if err != nil {
		t.Fatalf("Unexpected error getting user leaderboard: %v", err)
	}

	if rank != 7 {
		t.Errorf("Expected rank 7, got %d", rank)
	}

	if profile == nil {
		t.Fatal("Expected profile to be returned, got nil")
	}

	if profile.FirstName != "Juan" {
		t.Errorf("Expected profile name Juan, got %s", profile.FirstName)
	}
}

func TestGetUserLeaderboard_UserNotFound(t *testing.T) {
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	expectedErr := errors.New("user profile not found")
	repo := &mockUserRepository{
		getUserProfileFunc: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
			return nil, expectedErr
		},
	}
	s := services.NewUserService(repo)

	_, _, err := s.GetUserLeaderboard(context.Background(), userID)
	if err == nil {
		t.Fatal("Expected error when user profile is missing, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestGetUserLeaderboard_RankError(t *testing.T) {
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
	expectedErr := errors.New("rank calculation failed")
	repo := &mockUserRepository{
		getUserProfileFunc: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
			return &domain.UserProfile{ID: userID, FirstName: "Ana"}, nil
		},
		getUserRankFunc: func(ctx context.Context, uid uuid.UUID) (int, error) {
			return 0, expectedErr
		},
	}
	s := services.NewUserService(repo)

	_, _, err := s.GetUserLeaderboard(context.Background(), userID)
	if err == nil {
		t.Fatal("Expected error when rank calculation fails, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestDeleteUser_Failure(t *testing.T) {
	id := uuid.New()
	expectedErr := errors.New("foreign key constraint violation on delete")
	repo := &mockUserRepository{
		deleteUserFunc: func(ctx context.Context, uid uuid.UUID) error {
			return expectedErr
		},
	}
	s := services.NewUserService(repo)

	err := s.DeleteUser(context.Background(), id)
	if err == nil {
		t.Fatal("Expected error when deletion fails, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}
