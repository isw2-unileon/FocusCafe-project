package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

// syncStub starts an httptest.Server that routes requests for SyncUser tests.
// authStatus/authBody  → GET  /auth/v1/user
// existsBody           → GET  /rest/v1/users?id=eq.*
// profileStatus        → POST /rest/v1/users
// progressStatus       → POST /rest/v1/user_progress
func syncStub(
	t *testing.T,
	authStatus int, authBody interface{},
	existsBody interface{},
	profileStatus int,
	progressStatus int,
) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/auth/v1/user":
			w.WriteHeader(authStatus)
			_ = json.NewEncoder(w).Encode(authBody)

		case r.Method == http.MethodGet && r.URL.Path == "/rest/v1/users":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(existsBody)

		case r.Method == http.MethodPost && r.URL.Path == "/rest/v1/users":
			w.WriteHeader(profileStatus)
			_ = json.NewEncoder(w).Encode([]interface{}{})

		case r.Method == http.MethodPost && r.URL.Path == "/rest/v1/user_progress":
			w.WriteHeader(progressStatus)
			_ = json.NewEncoder(w).Encode([]interface{}{})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// validAuthBody returns a typical Supabase /auth/v1/user response.
func validAuthBody(userID, email, fullName string) map[string]interface{} {
	return map[string]interface{}{
		"id":    userID,
		"email": email,
		"user_metadata": map[string]interface{}{
			"full_name": fullName,
		},
	}
}

// ─────────────────────────────────────────────
// SyncUser – Authorization header tests
// ─────────────────────────────────────────────

func TestSyncUser_Auth_TableDriven(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authHeader     string
		authStatus     int
		authBody       interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Missing Authorization header",
			authHeader:     "",
			authStatus:     http.StatusOK,
			authBody:       validAuthBody("uuid-001", "user@focus.com", "Ada Lovelace"),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "required token",
		},
		{
			name:           "Invalid token - Supabase rejects it",
			authHeader:     "Bearer invalid-token",
			authStatus:     http.StatusUnauthorized,
			authBody:       map[string]interface{}{"message": "invalid JWT"},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid token",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			stub := syncStub(t,
				tt.authStatus, tt.authBody,
				[]interface{}{},
				http.StatusCreated,
				http.StatusCreated,
			)
			h := newHandler(stub.URL)

			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/sync", h.SyncUser)

			req, _ := http.NewRequest(http.MethodPost, "/sync", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var body map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			assert.Equal(t, tt.expectedError, body["error"])
		})
	}
}

// ─────────────────────────────────────────────
// SyncUser – User data extraction tests
// ─────────────────────────────────────────────

func TestSyncUser_ExtractUserData_TableDriven(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authBody       interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Auth response missing user id",
			authBody:       map[string]interface{}{"email": "user@focus.com"},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "missing user id",
		},
		{
			name:           "Auth response with full_name in metadata",
			authBody:       validAuthBody("uuid-001", "user@focus.com", "Ada Lovelace"),
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
		{
			name: "Auth response with name instead of full_name",
			authBody: map[string]interface{}{
				"id":    "uuid-002",
				"email": "user@focus.com",
				"user_metadata": map[string]interface{}{
					"name": "Ada Lovelace",
				},
			},
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
		{
			name: "Auth response with no metadata",
			authBody: map[string]interface{}{
				"id":    "uuid-003",
				"email": "user@focus.com",
			},
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			stub := syncStub(t,
				http.StatusOK, tt.authBody,
				[]interface{}{}, // user does not exist yet
				http.StatusCreated,
				http.StatusCreated,
			)
			h := newHandler(stub.URL)

			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/sync", h.SyncUser)

			req, _ := http.NewRequest(http.MethodPost, "/sync", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var body map[string]interface{}
				_ = json.Unmarshal(w.Body.Bytes(), &body)
				assert.Equal(t, tt.expectedError, body["error"])
			}
		})
	}
}

// ─────────────────────────────────────────────
// SyncUser – User already exists
// ─────────────────────────────────────────────

func TestSyncUser_UserAlreadyExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := syncStub(t,
		http.StatusOK, validAuthBody("uuid-001", "user@focus.com", "Ada Lovelace"),
		[]interface{}{map[string]interface{}{"id": "uuid-001"}}, // user exists
		http.StatusCreated,
		http.StatusCreated,
	)
	h := newHandler(stub.URL)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/sync", h.SyncUser)

	req, _ := http.NewRequest(http.MethodPost, "/sync", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, false, body["synced"])
	assert.Equal(t, "usuario ya existe", body["message"])
}

// ─────────────────────────────────────────────
// SyncUser – Profile creation failures
// ─────────────────────────────────────────────

func TestSyncUser_ProfileCreation_TableDriven(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		profileStatus  int
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Profile insertion fails with 500",
			profileStatus:  http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "error while saving the profile",
		},
		{
			name:           "Profile insertion fails with conflict",
			profileStatus:  http.StatusConflict,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "error while saving the profile",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			stub := syncStub(t,
				http.StatusOK, validAuthBody("uuid-001", "user@focus.com", "Ada Lovelace"),
				[]interface{}{}, // user does not exist
				tt.profileStatus,
				http.StatusCreated,
			)
			h := newHandler(stub.URL)

			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/sync", h.SyncUser)

			req, _ := http.NewRequest(http.MethodPost, "/sync", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var body map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			assert.Equal(t, tt.expectedError, body["error"])
		})
	}
}

// ─────────────────────────────────────────────
// SyncUser – Full success path
// ─────────────────────────────────────────────

func TestSyncUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := syncStub(t,
		http.StatusOK, validAuthBody("uuid-001", "ada@focus.com", "Ada Lovelace"),
		[]interface{}{}, // user does not exist yet
		http.StatusCreated,
		http.StatusCreated,
	)
	h := newHandler(stub.URL)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/sync", h.SyncUser)

	req, _ := http.NewRequest(http.MethodPost, "/sync", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var body map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	assert.Equal(t, true, body["synced"])
	assert.Equal(t, "uuid-001", body["id"])
	assert.Equal(t, "ada@focus.com", body["email"])
	assert.Equal(t, "Ada", body["first_name"])
	assert.Equal(t, "Lovelace", body["last_name"])
}
