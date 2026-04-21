package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────
// Shared mock (reutilizable across test files)
// ─────────────────────────────────────────────

// MockValidator already defined in auth_middleware_test.go;
// if both files live in the same package just remove this block.
// Uncomment if this file is compiled standalone:
//
// type MockValidator struct{ shouldFail bool }
// func (m *MockValidator) ValidateToken(token string) (*auth.UserClaims, error) { … }

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

// newHandler creates a Handler with the given Supabase base URL so tests
// can point it at a local httptest.Server.
func newHandler(supabaseURL string) *handlers.Handler {
	return &handlers.Handler{
		SupabaseURL: supabaseURL,
		SupabaseKey: "test-api-key",
		Auth:        &MockValidator{shouldFail: false},
	}
}

// supabaseStub starts an httptest.Server that responds to
// POST /auth/v1/token with the provided status code and body.
func supabaseStub(t *testing.T, statusCode int, body interface{}) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(body); err != nil {
			t.Errorf("supabaseStub: encode error: %v", err)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// loginBody serialises a LoginRequest into a *bytes.Buffer.
func loginBody(t *testing.T, email, password string) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(map[string]string{"email": email, "password": password})
	if err != nil {
		t.Fatalf("loginBody: %v", err)
	}
	return bytes.NewBuffer(b)
}

// ─────────────────────────────────────────────
// Login handler tests
// ─────────────────────────────────────────────

func TestLogin_TableDriven(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Supabase stub responses
	successBody := map[string]interface{}{
		"access_token": "jwt-token-abc123",
		"user":         map[string]string{"id": "uuid-999", "email": "user@focus.com"},
	}
	badCredsBody := map[string]interface{}{
		"error":             "invalid_grant",
		"error_description": "Invalid login credentials",
	}

	tests := []struct {
		name             string
		requestBody      interface{} // sent to the handler
		supabaseStatus   int
		supabaseResponse interface{}
		expectedStatus   int
		checkBody        func(t *testing.T, body map[string]interface{})
	}{
		{
			name:             "Success - valid credentials",
			requestBody:      map[string]string{"email": "user@focus.com", "password": "secret"},
			supabaseStatus:   http.StatusOK,
			supabaseResponse: successBody,
			expectedStatus:   http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "jwt-token-abc123", body["token"])
				assert.NotNil(t, body["user"])
			},
		},
		{
			name:             "Unauthorized - wrong password",
			requestBody:      map[string]string{"email": "user@focus.com", "password": "wrong"},
			supabaseStatus:   http.StatusBadRequest,
			supabaseResponse: badCredsBody,
			expectedStatus:   http.StatusUnauthorized,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Invalid login credentials", body["error"])
			},
		},
		{
			name:             "Bad request - malformed JSON body",
			requestBody:      "this-is-not-json",
			supabaseStatus:   http.StatusOK, // never reached
			supabaseResponse: successBody,
			expectedStatus:   http.StatusBadRequest,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Datos inválidos", body["error"])
			},
		},
		{
			name:             "Bad request - missing email field",
			requestBody:      map[string]string{"password": "secret"},
			supabaseStatus:   http.StatusBadRequest,
			supabaseResponse: badCredsBody,
			expectedStatus:   http.StatusUnauthorized,
			checkBody:        nil,
		},
		{
			name:             "Unauthorized - missing error_description in Supabase response",
			requestBody:      map[string]string{"email": "user@focus.com", "password": "wrong"},
			supabaseStatus:   http.StatusBadRequest,
			supabaseResponse: map[string]interface{}{"error": "unknown"},
			expectedStatus:   http.StatusUnauthorized,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Credenciales incorrectas", body["error"])
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			// --- Supabase stub ---
			stub := supabaseStub(t, tt.supabaseStatus, tt.supabaseResponse)

			// --- Handler + router ---
			h := newHandler(stub.URL)
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/login", h.Login)

			// --- Build request ---
			var reqBody *bytes.Buffer
			switch v := tt.requestBody.(type) {
			case string:
				reqBody = bytes.NewBufferString(v)
			default:
				b, _ := json.Marshal(v)
				reqBody = bytes.NewBuffer(b)
			}

			req, _ := http.NewRequest(http.MethodPost, "/login", reqBody)
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			// --- Assertions ---
			if w.Code != tt.expectedStatus {
				t.Errorf("[%s] expected HTTP %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}

			if tt.checkBody != nil {
				var respBody map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &respBody); err != nil {
					t.Fatalf("[%s] could not parse response body: %v", tt.name, err)
				}
				tt.checkBody(t, respBody)
			}
		})
	}
}

// ─────────────────────────────────────────────
// GoogleAuth handler tests
// ─────────────────────────────────────────────

func TestGoogleAuth_TableDriven(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		supabaseURL      string
		expectedStatus   int
		expectedLocation string
	}{
		{
			name:             "Redirects to Google provider via Supabase",
			supabaseURL:      "https://xyzcompany.supabase.co",
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: "https://xyzcompany.supabase.co/auth/v1/authorize?provider=google&redirect_to=http://localhost:5173/home",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			h := newHandler(tt.supabaseURL)

			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.GET("/auth/google", h.GoogleAuth)

			req, _ := http.NewRequest(http.MethodGet, "/auth/google", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("[%s] expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}

			location := w.Header().Get("Location")
			assert.Equal(t, tt.expectedLocation, location,
				"[%s] unexpected redirect URL", tt.name)
		})
	}
}

// ─────────────────────────────────────────────
// parseAuthResponse unit tests (via Login integration)
// ─────────────────────────────────────────────

func TestLogin_ParseAuthResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		supabaseStatus   int
		supabaseBody     interface{}
		expectedStatus   int
		expectedErrorMsg string
		expectedToken    string
	}{
		{
			name:             "Response without access_token field",
			supabaseStatus:   http.StatusOK,
			supabaseBody:     map[string]interface{}{"user": map[string]string{"id": "1"}},
			expectedStatus:   http.StatusUnauthorized,
			expectedErrorMsg: "error al obtener el token",
		},
		{
			name:           "Supabase returns 200 with valid token",
			supabaseStatus: http.StatusOK,
			supabaseBody: map[string]interface{}{
				"access_token": "valid.jwt.here",
				"user":         map[string]string{"email": "ok@focus.com"},
			},
			expectedStatus: http.StatusOK,
			expectedToken:  "valid.jwt.here",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			stub := supabaseStub(t, tt.supabaseStatus, tt.supabaseBody)
			h := newHandler(stub.URL)

			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/login", h.Login)

			req, _ := http.NewRequest(http.MethodPost, "/login",
				loginBody(t, "u@focus.com", "pass"))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("[%s] expected %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}

			var body map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &body)

			if tt.expectedErrorMsg != "" {
				assert.Equal(t, tt.expectedErrorMsg, body["error"])
			}
			if tt.expectedToken != "" {
				assert.Equal(t, tt.expectedToken, body["token"])
			}
		})
	}
}

// Ensure MockValidator satisfies auth.TokenValidator (compile-time check).
var _ auth.TokenValidator = (*MockValidator)(nil)
