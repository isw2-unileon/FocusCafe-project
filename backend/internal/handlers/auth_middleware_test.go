package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
)

// Mock simulating the validator
type MockValidator struct {
	shouldFail bool
}

func (m *MockValidator) ValidateToken(token string) (*auth.UserClaims, error) {
	if m.shouldFail {
		return nil, errors.New("simulated validation error")
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
			name:           "Token with multiple sequential spaces",
			authHeader:     "Bearer  token-valido",
			mockFail:       false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Rejected token by the validator",
			authHeader:     "Bearer bad-token",
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

			if w.Code != tt.expectedStatus {
				t.Errorf("%s: expected status %d but got %d", tt.name, tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAdminOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(c *gin.Context)
		mockBehavior   func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
		expectedStatus int
	}{
		{
			name: "Forbidden: No user in context",
			setupContext: func(c *gin.Context) {
				// Don't set "user"
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Forbidden: Invalid claims type",
			setupContext: func(c *gin.Context) {
				c.Set("user", "not user claims")
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Forbidden: Invalid UUID in claims",
			setupContext: func(c *gin.Context) {
				c.Set("user", &auth.UserClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: "invalid-uuid",
					},
				})
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Forbidden: User not found",
			setupContext: func(c *gin.Context) {
				c.Set("user", &auth.UserClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: uuid.New().String(),
					},
				})
			},
			mockBehavior: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
				return nil, errors.New("database or validation error")
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Forbidden: User is not admin",
			setupContext: func(c *gin.Context) {
				c.Set("user", &auth.UserClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: uuid.New().String(),
					},
				})
			},
			mockBehavior: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
				return &domain.UserProfile{Role: "user"}, nil
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "Success: User is admin",
			setupContext: func(c *gin.Context) {
				c.Set("user", &auth.UserClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: uuid.New().String(),
					},
				})
			},
			mockBehavior: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
				return &domain.UserProfile{Role: "admin"}, nil
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			mService := &mockUserService{getUserProfileFunc: tt.mockBehavior}
			h := &handlers.Handler{UserService: mService}

			r.GET("/admin", func(c *gin.Context) {
				tt.setupContext(c)
				h.AdminOnly()(c)
			}, func(c *gin.Context) {
				if !c.IsAborted() {
					c.Status(http.StatusOK)
				}
			})

			req, _ := http.NewRequest("GET", "/admin", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("%s: expected status %d but got %d", tt.name, tt.expectedStatus, w.Code)
			}
		})
	}
}
