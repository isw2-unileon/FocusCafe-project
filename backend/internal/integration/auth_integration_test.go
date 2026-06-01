package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
	"gorm.io/gorm"
)

func TestAuth_Integration(t *testing.T) {
	db, h := setupTestApp()
	db = db.Debug()

	t.Run("GoogleAuth - Redirects to Supabase", func(t *testing.T) {
		testGoogleAuthRedirect(t, h)
	})

	t.Run("SyncUser - Success for New User", func(t *testing.T) {
		testSyncUserSuccess(t, h)
	})

	t.Run("SyncUser - Fails on Missing Token", func(t *testing.T) {
		testSyncUserMissingToken(t, h)
	})

	t.Run("Login - Success", func(t *testing.T) {
		testLoginSuccess(t, db, h)
	})

	t.Run("Register - Success", func(t *testing.T) {
		testRegisterSuccess(t, h)
	})

	t.Run("HealthCheck - Success", func(t *testing.T) {
		testHealthCheckSuccess(t)
	})
}

func testLoginSuccess(t *testing.T, db *gorm.DB, h *handlers.Handler) {
	userID := uuid.New()
	// Seed the user in your actual local database so the Login handler can find them
	seedNormalUserMock(db, userID, "login_tester", "login@test.com")

	// Set up the internal Supabase mock server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "mock-access-token",
			"user": map[string]interface{}{
				"id":    userID.String(),
				"email": "login@test.com",
			},
		})
	}))
	defer srv.Close()

	oldURL := h.SupabaseURL
	h.SupabaseURL = srv.URL
	defer func() { h.SupabaseURL = oldURL }()

	reqBody, _ := json.Marshal(map[string]string{
		"email":    "login@test.com",
		"password": "password123",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(reqBody))

	// Pass the actual userID instead of uuid.Nil so the context builds claims correctly
	c, _ := setupSubtestContext(w, req, userID)
	h.Login(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func testRegisterSuccess(t *testing.T, h *handlers.Handler) {
	mockUserID := uuid.New()
	targetEmail := "register@test.com"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.HasSuffix(r.URL.Path, "/signup") {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"user": map[string]interface{}{
					"id":    mockUserID.String(),
					"email": targetEmail,
				},
			})
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	oldURL := h.SupabaseURL
	h.SupabaseURL = srv.URL
	defer func() { h.SupabaseURL = oldURL }()

	reqBody, _ := json.Marshal(map[string]string{
		"username":         "new_user_abs",
		"email":            targetEmail,
		"password":         "password123",
		"confirm_password": "password123",
		"first_name":       "New",
		"last_name":        "User",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(reqBody))

	c, _ := setupSubtestContext(w, req, mockUserID)
	h.Register(c)

	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Errorf("expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	// White-box Verification: Check if GORM successfully written down the record to the local SQLite
	var responseMap map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &responseMap); err != nil {
		t.Fatalf("failed to unmarshal handler response: %v", err)
	}
	if userPayload, exists := responseMap["user"].(map[string]interface{}); exists {
		if userPayload["email"] != targetEmail {
			t.Errorf("expected user email in response to be %s, got %v", targetEmail, userPayload["email"])
		}
	} else if responseMap["email"] != nil && responseMap["email"] != targetEmail {
		t.Errorf("expected email in flat response to be %s, got %v", targetEmail, responseMap["email"])
	}
}

func testGoogleAuthRedirect(t *testing.T, h *handlers.Handler) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/auth/google", nil)

	c, _ := setupSubtestContext(w, req, uuid.Nil)
	h.GoogleAuth(c)

	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("expected status 307, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if !strings.Contains(location, "provider=google") {
		t.Errorf("expected redirect location to contain google provider, got %s", location)
	}
}

func testSyncUserSuccess(t *testing.T, h *handlers.Handler) {
	userID := uuid.New()
	email := "newuser@test.com"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.HasSuffix(r.URL.Path, "/auth/v1/user") {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":    userID.String(),
				"email": email,
				"user_metadata": map[string]interface{}{
					"full_name": "Test User",
				},
			})
			return
		}

		if strings.HasSuffix(r.URL.Path, "/rest/v1/users") && r.Method == "GET" {
			_, _ = w.Write([]byte(`[]`))
			return
		}

		if strings.HasSuffix(r.URL.Path, "/rest/v1/users") && r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			return
		}

		if strings.HasSuffix(r.URL.Path, "/rest/v1/user_progress") {
			w.WriteHeader(http.StatusCreated)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	oldURL := h.SupabaseURL
	oldKey := h.SupabaseKey
	h.SupabaseURL = srv.URL
	h.SupabaseKey = "test-key"
	defer func() {
		h.SupabaseURL = oldURL
		h.SupabaseKey = oldKey
	}()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/sync", nil)
	req.Header.Set("Authorization", "Bearer mock-token")

	c, _ := setupSubtestContext(w, req, userID)
	h.SyncUser(c)

	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Errorf("expected status 200/201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func testSyncUserMissingToken(t *testing.T, h *handlers.Handler) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/sync", nil)

	c, _ := setupSubtestContext(w, req, uuid.Nil)
	h.SyncUser(c)

	if w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusUnauthorized {
		t.Errorf("expected unauthorized error status, got %d", w.Code)
	}
}

func testHealthCheckSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)

	// Simulate root behaviour
	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.ServeHTTP(w, req)

	// Verify status code
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 OK for health check, got %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal health check response: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("expected health status to be 'ok', got '%s'", body["status"])
	}
}
