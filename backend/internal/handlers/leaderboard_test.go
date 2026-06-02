package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
)

// mockUserServiceForLeaderboard is a test double for the UserServiceInterface.
type mockUserServiceForLeaderboard struct {
	getLeaderboardFunc     func(ctx context.Context, limit int) ([]domain.UserProfile, error)
	getUserLeaderboardFunc func(ctx context.Context, userID uuid.UUID) (int, *domain.UserProfile, error)
}

func (m *mockUserServiceForLeaderboard) GetUserProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	return nil, nil
}

func (m *mockUserServiceForLeaderboard) UpdateUserProfile(ctx context.Context, id uuid.UUID, firstName, lastName string) error {
	return nil
}

func (m *mockUserServiceForLeaderboard) GetAllUsers(ctx context.Context) ([]domain.UserProfile, error) {
	return nil, nil
}

func (m *mockUserServiceForLeaderboard) GetUserByEmail(ctx context.Context, email string) (*domain.UserProfile, error) {
	return nil, nil
}

func (m *mockUserServiceForLeaderboard) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockUserServiceForLeaderboard) GetLeaderboard(ctx context.Context, limit int) ([]domain.UserProfile, error) {
	if m.getLeaderboardFunc != nil {
		return m.getLeaderboardFunc(ctx, limit)
	}
	return nil, errors.New("getLeaderboard not mocked")
}

func (m *mockUserServiceForLeaderboard) GetUserLeaderboard(ctx context.Context, userID uuid.UUID) (int, *domain.UserProfile, error) {
	if m.getUserLeaderboardFunc != nil {
		return m.getUserLeaderboardFunc(ctx, userID)
	}
	return 0, nil, errors.New("getUserLeaderboard not mocked")
}

// ============================================
// TestHandler_GetLeaderboard
// ============================================

func TestHandler_GetLeaderboard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name             string
		userIDInContext  uuid.UUID
		mockBehavior     func(ctx context.Context, limit int) ([]domain.UserProfile, error)
		wantStatusCode   int
		expectedBody     string
	}{
		{
			name:            "Success: Returns leaderboard top 5",
			userIDInContext: testUUID,
			mockBehavior: func(ctx context.Context, limit int) ([]domain.UserProfile, error) {
				return []domain.UserProfile{
					{
						ID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
						FirstName: "Alice",
						Level:     15,
						XP:        3200,
					},
					{
						ID:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
						FirstName: "Bob",
						Level:     12,
						XP:        2100,
					},
				}, nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `"first_name":"Alice"`,
		},
		{
			name:            "Success: Returns empty leaderboard",
			userIDInContext: testUUID,
			mockBehavior: func(ctx context.Context, limit int) ([]domain.UserProfile, error) {
				return []domain.UserProfile{}, nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `[]`,
		},
		{
			name:            "Error: Service failure returns 500",
			userIDInContext: testUUID,
			mockBehavior: func(ctx context.Context, limit int) ([]domain.UserProfile, error) {
				return nil, errors.New("database connection failed")
			},
			wantStatusCode: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to fetch leaderboard"}`,
		},
		{
			name:            "Error: No user in context returns 401",
			userIDInContext: uuid.Nil,
			mockBehavior:    nil,
			wantStatusCode:  http.StatusUnauthorized,
			expectedBody:    `{"error":"unauthorized"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mService := &mockUserServiceForLeaderboard{getLeaderboardFunc: tt.mockBehavior}
			h := &handlers.Handler{UserService: mService}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.userIDInContext != uuid.Nil {
				injectUserClaims(c, tt.userIDInContext)
			}

			c.Request = httptest.NewRequest("GET", "/api/leaderboard", nil)

			h.GetLeaderboard(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler.GetLeaderboard() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Handler.GetLeaderboard() body = %v, want to contain %v", w.Body.String(), tt.expectedBody)
			}
		})
	}
}

// ============================================
// TestHandler_GetUserLeaderboardRank
// ============================================

func TestHandler_GetUserLeaderboardRank(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name                 string
		userIDInContext      uuid.UUID
		mockBehavior         func(ctx context.Context, userID uuid.UUID) (int, *domain.UserProfile, error)
		wantStatusCode       int
		expectedBodyContains string
	}{
		{
			name:            "Success: Returns user rank and profile",
			userIDInContext: testUUID,
			mockBehavior: func(ctx context.Context, userID uuid.UUID) (int, *domain.UserProfile, error) {
				return 7, &domain.UserProfile{
					ID:        userID,
					FirstName: "Juan",
					Level:     5,
					XP:        800,
				}, nil
			},
			wantStatusCode:       http.StatusOK,
			expectedBodyContains: `"rank":7`,
		},
		{
			name:            "Success: Returns rank 1 for top player",
			userIDInContext: testUUID,
			mockBehavior: func(ctx context.Context, userID uuid.UUID) (int, *domain.UserProfile, error) {
				return 1, &domain.UserProfile{
					ID:        userID,
					FirstName: "Alice",
					Level:     15,
					XP:        3200,
				}, nil
			},
			wantStatusCode:       http.StatusOK,
			expectedBodyContains: `"rank":1`,
		},
		{
			name:            "Error: Service failure returns 500",
			userIDInContext: testUUID,
			mockBehavior: func(ctx context.Context, userID uuid.UUID) (int, *domain.UserProfile, error) {
				return 0, nil, errors.New("database connection failed")
			},
			wantStatusCode:       http.StatusInternalServerError,
			expectedBodyContains: `{"error":"failed to fetch user rank"}`,
		},
		{
			name:                 "Error: No user in context returns 401",
			userIDInContext:      uuid.Nil,
			mockBehavior:         nil,
			wantStatusCode:       http.StatusUnauthorized,
			expectedBodyContains: `{"error":"unauthorized"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mService := &mockUserServiceForLeaderboard{getUserLeaderboardFunc: tt.mockBehavior}
			h := &handlers.Handler{UserService: mService}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.userIDInContext != uuid.Nil {
				injectUserClaims(c, tt.userIDInContext)
			}

			c.Request = httptest.NewRequest("GET", "/api/leaderboard/me", nil)

			h.GetUserLeaderboardRank(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler.GetUserLeaderboardRank() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBodyContains) {
				t.Errorf("Handler.GetUserLeaderboardRank() body = %v, want to contain %v", w.Body.String(), tt.expectedBodyContains)
			}
		})
	}
}
