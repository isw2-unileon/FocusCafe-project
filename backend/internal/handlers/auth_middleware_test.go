package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
	"github.com/stretchr/testify/assert"
)

// Mock simulating the validator
type MockValidator struct {
	shouldFail bool
}

func (m *MockValidator) ValidateToken(token string) (*auth.UserClaims, error) {
	if m.shouldFail {
		return nil, assert.AnError
	}
	return &auth.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "uuid-123",
		},
		Email: "test@focus.com",
	}, nil
}

func TestAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authHeader     string
		mockFail       bool
		expectedStatus int
	}{
		{
			name:           "Without header authorization",
			authHeader:     "",
			mockFail:       false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid header format (without Bearer)",
			authHeader:     "Basic dXNlcjpwYXNz",
			mockFail:       false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Empty token",
			authHeader:     "Bearer ",
			mockFail:       false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Rejected token by the validator",
			authHeader:     "Bearer bad token",
			mockFail:       true,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Success",
			authHeader:     "Bearer valid-token",
			mockFail:       false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			mock := &MockValidator{shouldFail: tt.mockFail}

			r.GET("/protected", handlers.Auth(mock), func(ctx *gin.Context) {
				ctx.Status(http.StatusOK)
			})

			req, _ := http.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			r.ServeHTTP(w, req)

			// Replacement for assert.Equal
			if w.Code != tt.expectedStatus {
				t.Errorf("%s: expected status %d but got %d", tt.name, tt.expectedStatus, w.Code)
			}
		})
	}
}
