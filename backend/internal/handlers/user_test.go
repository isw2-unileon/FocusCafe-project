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

type mockUserService struct {
	getUserProfileFunc func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
}

func (m *mockUserService) GetUserProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	return m.getUserProfileFunc(ctx, id)
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
			mockBehavior: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
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
			name:            "Error: User not found returns 404",
			userIDInContext: uuid.New(),
			mockBehavior: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
				return nil, errors.New("record not found")
			},
			wantStatusCode: http.StatusNotFound,
			expectedBody:   `{"error":"user not found"}`,
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

			// Simulate Middleware: inject the ID into the context
			// Ensure this key matches what h.getUserID(c) looks for
			c.Set("user_id", tt.userIDInContext)

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
