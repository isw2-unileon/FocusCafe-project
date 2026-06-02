package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
)

type mockUserService struct {
	getUserProfileFunc      func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
	updateUserProfileFunc   func(ctx context.Context, id uuid.UUID, firstName, lastName string) error
	getAllUsersFunc         func(ctx context.Context) ([]domain.UserProfile, error)
	getUserByEmailFunc      func(ctx context.Context, email string) (*domain.UserProfile, error)
	deleteUserFunc          func(ctx context.Context, id uuid.UUID) error
	getLeaderboardFunc      func(ctx context.Context, limit int) ([]domain.UserProfile, error)
	getUserLeaderboardFunc  func(ctx context.Context, userID uuid.UUID) (int, *domain.UserProfile, error)
}

func (m *mockUserService) GetUserProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	return m.getUserProfileFunc(ctx, id)
}

func (m *mockUserService) UpdateUserProfile(ctx context.Context, id uuid.UUID, firstName, lastName string) error {
	return m.updateUserProfileFunc(ctx, id, firstName, lastName)
}

func (m *mockUserService) GetAllUsers(ctx context.Context) ([]domain.UserProfile, error) {
	return m.getAllUsersFunc(ctx)
}

func (m *mockUserService) GetUserByEmail(ctx context.Context, email string) (*domain.UserProfile, error) {
	return m.getUserByEmailFunc(ctx, email)
}

func (m *mockUserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return m.deleteUserFunc(ctx, id)
}

func (m *mockUserService) GetLeaderboard(ctx context.Context, limit int) ([]domain.UserProfile, error) {
	return m.getLeaderboardFunc(ctx, limit)
}

func (m *mockUserService) GetUserLeaderboard(ctx context.Context, userID uuid.UUID) (int, *domain.UserProfile, error) {
	return m.getUserLeaderboardFunc(ctx, userID)
}

func TestHandler_GetUserProfile(t *testing.T) {
	// Set Gin to Test Mode to keep logs clean
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		userIDInContext uuid.UUID
		mockBehavior    func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
		wantStatusCode  int
		expectedBody    string
	}{
		{
			name:            "Success: Returns 200 and Profile",
			userIDInContext: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			mockBehavior: func(_ context.Context, id uuid.UUID) (*domain.UserProfile, error) {
				if id != uuid.MustParse("550e8400-e29b-41d4-a716-446655440000") {
					return nil, errors.New("wrong id passed")
				}
				return &domain.UserProfile{
					ID:     id,
					Energy: 500,
					XP:     100,
				}, nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `"energy":500`, // We check for key fragments
		},
		{
			name:            "Error: User not found returns 401",
			userIDInContext: uuid.New(),
			mockBehavior: func(_ context.Context, id uuid.UUID) (*domain.UserProfile, error) {
				return nil, errors.New("record not found")
			},
			wantStatusCode: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Setup Mock and Handler
			mService := &mockUserService{getUserProfileFunc: tt.mockBehavior}
			h := &handlers.Handler{UserService: mService}

			// 2. Setup Gin Recorder and Context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Simulate Middleware: inject the UserClaims into the context
			claims := &auth.UserClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: tt.userIDInContext.String(),
				},
			}
			c.Set("user", claims)

			// Create a dummy request to avoid nil pointer panics
			c.Request = httptest.NewRequest("GET", "/api/v1/profile", nil)

			// 3. Execute the Handler
			h.GetUserProfile(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler.GetUserProfile() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			// Check if the response body contains the expected strings
			gotBody := w.Body.String()
			if !strings.Contains(gotBody, tt.expectedBody) {
				t.Errorf("Handler.GetUserProfile() body = %v, want to contain %v", gotBody, tt.expectedBody)
			}
		})
	}
}

func TestHandler_UpdateUserProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name                   string
		userIDInContext        interface{}
		requestBody            string
		mockUpdateBehavior     func(ctx context.Context, id uuid.UUID, firstName, lastName string) error
		mockGetProfileBehavior func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
		wantStatusCode         int
		expectedBody           string
	}{
		{
			name:            "Success: Returns 200 and Updated Profile",
			userIDInContext: uuid.New().String(),
			requestBody:     `{"first_name": "John", "last_name": "Doe"}`,
			mockUpdateBehavior: func(_ context.Context, id uuid.UUID, fn, ln string) error {
				return nil
			},
			mockGetProfileBehavior: func(_ context.Context, id uuid.UUID) (*domain.UserProfile, error) {
				return &domain.UserProfile{FirstName: "John", LastName: "Doe"}, nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `"first_name":"John"`,
		},
		{
			name:            "Error: Unauthorized if user claims missing",
			userIDInContext: nil,
			requestBody:     `{"first_name": "John", "last_name": "Doe"}`,
			wantStatusCode:  http.StatusUnauthorized,
			expectedBody:    `{"error":"unauthorized"}`,
		},
		{
			name:            "Error: Invalid JSON body",
			userIDInContext: uuid.New().String(),
			requestBody:     `{invalid json}`,
			wantStatusCode:  http.StatusBadRequest,
			expectedBody:    `{"error":"invalid request body"}`,
		},
		{
			name:            "Error: Empty fields",
			userIDInContext: uuid.New().String(),
			requestBody:     `{"first_name": " ", "last_name": ""}`,
			wantStatusCode:  http.StatusBadRequest,
			expectedBody:    `{"error":"first_name and last_name are required"}`,
		},
		{
			name:            "Error: Internal Server Error on update failure",
			userIDInContext: uuid.New().String(),
			requestBody:     `{"first_name": "John", "last_name": "Doe"}`,
			mockUpdateBehavior: func(_ context.Context, id uuid.UUID, fn, ln string) error {
				return errors.New("update error")
			},
			wantStatusCode: http.StatusInternalServerError,
			expectedBody:    `{"error":"failed to update profile"}`,
		},
		{
			name:            "Error: Internal Server Error on fetch failure",
			userIDInContext: uuid.New().String(),
			requestBody:     `{"first_name": "John", "last_name": "Doe"}`,
			mockUpdateBehavior: func(_ context.Context, id uuid.UUID, fn, ln string) error {
				return nil
			},
			mockGetProfileBehavior: func(_ context.Context, id uuid.UUID) (*domain.UserProfile, error) {
				return nil, errors.New("fetch error")
			},
			wantStatusCode: http.StatusInternalServerError,
			expectedBody:    `{"error":"failed to fetch updated profile"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mService := &mockUserService{
				updateUserProfileFunc: tt.mockUpdateBehavior,
				getUserProfileFunc:    tt.mockGetProfileBehavior,
			}
			h := &handlers.Handler{UserService: mService}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.userIDInContext != nil {
				claims := &auth.UserClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: tt.userIDInContext.(string),
					},
				}
				c.Set("user", claims)
			}

			c.Request = httptest.NewRequest("PUT", "/api/v1/profile", strings.NewReader(tt.requestBody))
			c.Request.Header.Set("Content-Type", "application/json")
			h.UpdateUserProfile(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler.UpdateUserProfile() status = %v, want %v", w.Code, tt.wantStatusCode)
			}
			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Handler.UpdateUserProfile() body = %v, want %v", w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestHandler_GetUserID_EdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userInContext  interface{}
		wantStatusCode int
		expectedBody   string
	}{
		{
			name:           "Error: Invalid claims format",
			userInContext:  "not a UserClaims object",
			wantStatusCode: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized"}`,
		},
		{
			name: "Error: Empty user ID",
			userInContext: &auth.UserClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "",
				},
			},
			wantStatusCode: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized"}`,
		},
		{
			name: "Error: Invalid UUID format",
			userInContext: &auth.UserClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: "not-a-uuid",
				},
			},
			wantStatusCode: http.StatusUnauthorized,
			expectedBody:   `{"error":"unauthorized"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mService := &mockUserService{
				getUserProfileFunc: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
					return nil, errors.New("unauthorized")
				},
			}
			h := &handlers.Handler{UserService: mService}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.userInContext != nil {
				c.Set("user", tt.userInContext)
			}

			c.Request = httptest.NewRequest("GET", "/api/v1/profile", nil)
			h.GetUserProfile(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler.GetUserProfile() status = %v, want %v", w.Code, tt.wantStatusCode)
			}
			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Handler.GetUserProfile() body = %v, want %v", w.Body.String(), tt.expectedBody)
			}
		})
	}
}
