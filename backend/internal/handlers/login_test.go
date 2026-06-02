package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────
// Helpers & Stubs
// ─────────────────────────────────────────────

// newLoginHandler inicializa el Handler con un nombre único para evitar duplicados
func newLoginHandler(supabaseURL string) *Handler {
	return &Handler{
		SupabaseURL: supabaseURL,
		SupabaseKey: "test-api-key",
		ClientURL:   "http://localhost:5173",
	}
}

// supabaseLoginStub levanta un servidor de mentira único para las pruebas de login
func supabaseLoginStub(t *testing.T, statusCode int, body interface{}) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		if strBody, ok := body.(string); ok {
			_, _ = w.Write([]byte(strBody))
		} else {
			if err := json.NewEncoder(w).Encode(body); err != nil {
				t.Errorf("supabaseLoginStub: encode error: %v", err)
			}
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// loginBody serializa las credenciales rápidamente para las peticiones de test
func loginBody(t *testing.T, email, password string) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(map[string]string{"email": email, "password": password})
	if err != nil {
		t.Fatalf("loginBody: %v", err)
	}
	return bytes.NewBuffer(b)
}

// ─────────────────────────────────────────────
// Login handler tests (Table-Driven)
// ─────────────────────────────────────────────

type loginTestCase struct {
	name             string
	requestBody      interface{}
	supabaseStatus   int
	supabaseResponse interface{}
	closeServer      bool
	expectedStatus   int
	expectedError    string
	containsError    string
	expectedToken    string
	checkUser        bool
}

func runLoginTestCase(t *testing.T, tt loginTestCase) {
	stub := supabaseLoginStub(t, tt.supabaseStatus, tt.supabaseResponse)
	if tt.closeServer {
		stub.Close()
	}

	h := newLoginHandler(stub.URL)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/login", h.Login)

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

	if w.Code != tt.expectedStatus {
		t.Errorf("[%s] expected HTTP %d, got %d", tt.name, tt.expectedStatus, w.Code)
	}

	var respBody map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &respBody)

	if tt.expectedError != "" {
		if respBody["error"] != tt.expectedError {
			t.Errorf("[%s] expected error %q, got %q", tt.name, tt.expectedError, respBody["error"])
		}
	}
	if tt.containsError != "" {
		errStr, _ := respBody["error"].(string)
		if !strings.Contains(errStr, tt.containsError) {
			t.Errorf("[%s] expected error to contain %q, got %q", tt.name, tt.containsError, errStr)
		}
	}
	if tt.expectedToken != "" {
		if respBody["token"] != tt.expectedToken {
			t.Errorf("[%s] expected token %q, got %q", tt.name, tt.expectedToken, respBody["token"])
		}
	}
	if tt.checkUser {
		if respBody["user"] == nil {
			t.Errorf("[%s] expected user to be not nil", tt.name)
		}
	}
}

func TestLogin_TableDriven(t *testing.T) {
	gin.SetMode(gin.TestMode)

	successBody := map[string]interface{}{
		"access_token": "jwt-token-abc123",
		"user":         map[string]string{"id": "uuid-999", "email": "user@focus.com"},
	}
	badCredsBody := map[string]interface{}{
		"error":             "invalid_grant",
		"error_description": "Invalid login credentials",
	}

	tests := []loginTestCase{
		{
			name:             "Success - valid credentials",
			requestBody:      map[string]string{"email": "user@focus.com", "password": "secret"},
			supabaseStatus:   http.StatusOK,
			supabaseResponse: successBody,
			expectedStatus:   http.StatusOK,
			expectedToken:    "jwt-token-abc123",
			checkUser:        true,
		},
		{
			name:             "Unauthorized - wrong password",
			requestBody:      map[string]string{"email": "user@focus.com", "password": "wrong"},
			supabaseStatus:   http.StatusBadRequest,
			supabaseResponse: badCredsBody,
			expectedStatus:   http.StatusUnauthorized,
			expectedError:    "Invalid login credentials",
		},
		{
			name:             "Bad request - malformed JSON body",
			requestBody:      "this-is-not-json",
			supabaseStatus:   http.StatusOK,
			supabaseResponse: successBody,
			expectedStatus:   http.StatusBadRequest,
			expectedError:    "invalid request body",
		},
		{
			name:             "Bad request - missing email field",
			requestBody:      map[string]string{"password": "secret"},
			supabaseStatus:   http.StatusBadRequest,
			supabaseResponse: badCredsBody,
			expectedStatus:   http.StatusUnauthorized,
			expectedError:    "Invalid login credentials",
		},
		{
			name:             "Unauthorized - msg field in Supabase response",
			requestBody:      map[string]string{"email": "user@focus.com", "password": "wrong"},
			supabaseStatus:   http.StatusBadRequest,
			supabaseResponse: map[string]interface{}{"msg": "Specific error message from msg field"},
			expectedStatus:   http.StatusUnauthorized,
			expectedError:    "Specific error message from msg field",
		},
		{
			name:             "Unauthorized - message field in Supabase response",
			requestBody:      map[string]string{"email": "user@focus.com", "password": "wrong"},
			supabaseStatus:   http.StatusBadRequest,
			supabaseResponse: map[string]interface{}{"message": "Specific error message from message field"},
			expectedStatus:   http.StatusUnauthorized,
			expectedError:    "Specific error message from message field",
		},
		{
			name:             "Unauthorized - missing error_description in Supabase response",
			requestBody:      map[string]string{"email": "user@focus.com", "password": "wrong"},
			supabaseStatus:   http.StatusBadRequest,
			supabaseResponse: map[string]interface{}{"error": "unknown"},
			expectedStatus:   http.StatusUnauthorized,
			expectedError:    "invalid credentials",
		},
		{
			name:             "Supabase returns corrupt JSON response",
			requestBody:      map[string]string{"email": "user@focus.com", "password": "pass"},
			supabaseStatus:   http.StatusOK,
			supabaseResponse: "{bad-json-corrupt-data",
			expectedStatus:   http.StatusUnauthorized,
			containsError:    "error processing auth response",
		},
		{
			name:             "Supabase network connection error",
			requestBody:      map[string]string{"email": "user@focus.com", "password": "pass"},
			supabaseStatus:   http.StatusOK,
			supabaseResponse: successBody,
			closeServer:      true,
			expectedStatus:   http.StatusUnauthorized,
			expectedError:    "connection error: could not reach auth service",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			runLoginTestCase(t, tt)
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
			h := newLoginHandler(tt.supabaseURL)

			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.GET("/auth/google", h.GoogleAuth)

			req, _ := http.NewRequest(http.MethodGet, "/auth/google", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("[%s] expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}

			location := w.Header().Get("Location")
			if location != tt.expectedLocation {
				t.Errorf("expected location %q, got %q", tt.expectedLocation, location)
			}
		})
	}
}

// ─────────────────────────────────────────────
// parseAuthResponse specific cases
// ─────────────────────────────────────────────

func TestLogin_ParseAuthResponse_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := supabaseLoginStub(t, http.StatusOK, map[string]interface{}{"user": map[string]string{"id": "1"}})
	h := newLoginHandler(stub.URL)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/login", h.Login)

	req, _ := http.NewRequest(http.MethodPost, "/login", loginBody(t, "u@focus.com", "pass"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var body map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
	if body["error"] != "authentication token not found" {
		t.Errorf("expected error %q, got %q", "authentication token not found", body["error"])
	}
}

// ─────────────────────────────────────────────
// Direct Unit Tests (Para cubrir líneas huérfanas)
// ─────────────────────────────────────────────

func TestCallSupabaseAuth_InvalidMethod(t *testing.T) {
	h := newLoginHandler("http://localhost:8080")

	h.SupabaseURL = "http://[::1]:abcd" // URL inválida para forzar que falle el request interno
	_, err := h.callSupabaseAuth([]byte(`{}`))
	if err == nil {
		t.Error("expected error, got nil")
	}
}
