package handlers

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

// newSyncHandler inicializa el Handler con un nombre único para esta clase de tests
func newSyncHandler(supabaseURL string) *Handler {
	return &Handler{
		SupabaseURL: supabaseURL,
		SupabaseKey: "test-api-key",
		ClientURL:   "http://localhost:5173",
	}
}

// supabaseSyncStub arranca el servidor de simulación exclusivo para SyncUser
func supabaseSyncStub(
	t *testing.T,
	authStatus int, authBody interface{},
	existsBody interface{},
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
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode([]interface{}{})

		case r.Method == http.MethodPost && r.URL.Path == "/rest/v1/user_progress":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode([]interface{}{})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// validAuthBody returns a typical Supabase /auth/v1/user response.
func validAuthBody(email string) map[string]interface{} {
	return map[string]interface{}{
		"id":    "uuid-001",
		"email": email,
		"user_metadata": map[string]interface{}{
			"full_name": "Ada Lovelace",
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
			authBody:       validAuthBody("user@focus.com"),
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
			stub := supabaseSyncStub(t,
				tt.authStatus, tt.authBody,
				[]interface{}{},
			)
			h := newSyncHandler(stub.URL)

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
			expectedError:  "user id not found in token",
		},
		{
			name:           "Auth response with full_name in metadata",
			authBody:       validAuthBody("user@focus.com"),
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
			stub := supabaseSyncStub(t,
				http.StatusOK, tt.authBody,
				[]interface{}{}, // user does not exist yet
			)
			h := newSyncHandler(stub.URL)

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

	stub := supabaseSyncStub(t,
		http.StatusOK, validAuthBody("user@focus.com"),
		[]interface{}{map[string]interface{}{"id": "uuid-001"}}, // user exists
	)
	h := newSyncHandler(stub.URL)

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
// SyncUser – User existence, Profile and Progress failures
// ─────────────────────────────────────────────

func TestSyncUser_Failures_TableDriven(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		existsStatus   int
		profileStatus  int
		progressStatus int
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "User existence check fails (Line 40)",
			existsStatus:   http.StatusInternalServerError,
			profileStatus:  http.StatusCreated,
			progressStatus: http.StatusCreated,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "error while verifying the user",
		},
		{
			name:           "Profile creation fails (Line 53)",
			existsStatus:   http.StatusOK,
			profileStatus:  http.StatusInternalServerError,
			progressStatus: http.StatusCreated,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "error while saving the profile",
		},
		{
			name:           "Progress creation fails (Line 59)",
			existsStatus:   http.StatusOK,
			profileStatus:  http.StatusCreated,
			progressStatus: http.StatusInternalServerError,
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to save initial progress", // error from createUserProgress
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.URL.Path == "/auth/v1/user":
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(validAuthBody("user@focus.com"))
				case r.URL.Path == "/rest/v1/users":
					if r.Method == http.MethodGet {
						w.WriteHeader(tt.existsStatus)
						_ = json.NewEncoder(w).Encode([]interface{}{})
					} else {
						w.WriteHeader(tt.profileStatus)
						_ = json.NewEncoder(w).Encode([]interface{}{})
					}
				case r.URL.Path == "/rest/v1/user_progress":
					w.WriteHeader(tt.progressStatus)
					if tt.progressStatus != http.StatusCreated {
						_ = json.NewEncoder(w).Encode(map[string]string{"message": "failed to save initial progress"})
					}
				}
			}))
			defer srv.Close()

			h := newSyncHandler(srv.URL)
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/sync", h.SyncUser)

			req, _ := http.NewRequest(http.MethodPost, "/sync", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			var body map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			assert.Contains(t, body["error"].(string), tt.expectedError)
		})
	}
}

func TestCreateUserProfileSync_FallbackError(t *testing.T) {
	// Cubre la línea 190: JSON válido pero sin campos 'message' o 'code'
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"something": "else"})
	}))
	defer stub.Close()
	h := newSyncHandler(stub.URL)
	err := h.createUserProfileSync("uuid", "test@test.com", "First", "Last")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create profile (status 400)")
}

// ─────────────────────────────────────────────
// SyncUser – User data extraction edge cases
// ─────────────────────────────────────────────

func TestExtractUserData_EdgeCases(t *testing.T) {
	tests := []struct {
		name              string
		userData          map[string]any
		expectedFirstName string
		expectedLastName  string
	}{
		{
			name: "Only first name in full_name",
			userData: map[string]any{
				"id": "uuid-1",
				"user_metadata": map[string]any{
					"full_name": "Cher",
				},
			},
			expectedFirstName: "Cher",
			expectedLastName:  "",
		},
		{
			name: "Name with multiple spaces",
			userData: map[string]any{
				"id": "uuid-2",
				"user_metadata": map[string]any{
					"name": "Marie Curie Skłodowska",
				},
			},
			expectedFirstName: "Marie",
			expectedLastName:  "Curie Skłodowska",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, fn, ln, err := extractUserData(tt.userData)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedFirstName, fn)
			assert.Equal(t, tt.expectedLastName, ln)
		})
	}
}

// ─────────────────────────────────────────────
// Direct Unit Tests for Helpers
// ─────────────────────────────────────────────

func TestFetchSupabaseUser_Errors(t *testing.T) {
	t.Run("Connection Error", func(t *testing.T) {
		h := newSyncHandler("http://localhost:9999") // invalid URL
		_, err := h.fetchSupabaseUser("token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error connecting to auth service")
	})

	t.Run("Decoding Error", func(t *testing.T) {
		stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{invalid-json"))
		}))
		defer stub.Close()
		h := newSyncHandler(stub.URL)
		_, err := h.fetchSupabaseUser("token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error decoding user data")
	})
}

func TestUserExists_Errors(t *testing.T) {
	t.Run("Connection Error", func(t *testing.T) {
		h := newSyncHandler("http://localhost:9999")
		_, err := h.userExists("uuid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error connecting to database")
	})

	t.Run("Non-200 Status", func(t *testing.T) {
		stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer stub.Close()
		h := newSyncHandler(stub.URL)
		_, err := h.userExists("uuid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error verifying user existence")
	})

	t.Run("Decoding Error", func(t *testing.T) {
		stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{invalid-json"))
		}))
		defer stub.Close()
		h := newSyncHandler(stub.URL)
		_, err := h.userExists("uuid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error decoding database response")
	})
}

func TestCreateUserProfileSync_Errors(t *testing.T) {
	t.Run("Network Error", func(t *testing.T) {
		h := newSyncHandler("http://localhost:9999")
		err := h.createUserProfileSync("uuid", "test@test.com", "First", "Last")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "network error while creating profile")
	})

	t.Run("Duplicate Error (23505)", func(t *testing.T) {
		stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "23505"})
		}))
		defer stub.Close()
		h := newSyncHandler(stub.URL)
		err := h.createUserProfileSync("uuid", "test@test.com", "First", "Last")
		assert.NoError(t, err) // Should return nil for duplicate
	})

	t.Run("Database Error with Message", func(t *testing.T) {
		stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "custom error"})
		}))
		defer stub.Close()
		h := newSyncHandler(stub.URL)
		err := h.createUserProfileSync("uuid", "test@test.com", "First", "Last")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error: custom error")
	})

	t.Run("Decoding error message fails", func(t *testing.T) {
		stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("{invalid-json"))
		}))
		defer stub.Close()
		h := newSyncHandler(stub.URL)
		err := h.createUserProfileSync("uuid", "test@test.com", "First", "Last")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create profile (status 400)")
	})
}

func TestSyncUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := supabaseSyncStub(t,
		http.StatusOK, validAuthBody("ada@focus.com"),
		[]interface{}{}, // user does not exist yet
	)
	h := newSyncHandler(stub.URL)

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
