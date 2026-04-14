package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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
	return &auth.UserClaims{Email: "test@focus.com", ID: "uuid-123"}, nil
}

func TestAuthMiddleware_TableDriven(t *testing.T) {
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

			// Definimos una ruta de prueba que usa el middleware
			r.GET("/protected", handlers.Auth(mock), func(ctx *gin.Context) {
				ctx.Status(http.StatusOK)
			})

			// Ejecución
			req, _ := http.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			r.ServeHTTP(w, req)

			// Verificación
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

/*
func TestAuth(t *testing.T) {
	secretTest := "secret-quite-secure"
	// cases table
	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Missing Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header missing",
		},
		{
			name:           "Wrong Header format",
			authHeader:     "BadToken123",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid authorization format",
		},
		{
			name:           "Empty Token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Token is empty",
		},
	}
/*
	gin.SetMode(gin.TestMode)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			/*r.Use(handlers.Auth(secretTest))

			r.GET("/ping", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			// Petition
			req, _ := http.NewRequest("GET", "/ping", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			r.ServeHTTP(w, req)

			// Verification
			if w.Code != tt.expectedStatus {
				t.Errorf("Failed test %s expected = %d, got= %d", tt.name, tt.expectedStatus, w.Code)
			}

			// verify JSON response
			if tt.expectedStatus != http.StatusOK {
				var response map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("JSON response could not be parsed: %v", err)
				}

				if response["error"] != tt.expectedError {
					t.Errorf("Failed JSON expected = %s, got= %s", tt.expectedError, response["error"])
				}
			}
		})
	}
}*/
