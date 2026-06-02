package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newSyncHandler(supabaseURL string) *Handler {
	return &Handler{
		SupabaseURL: supabaseURL,
		SupabaseKey: "test-api-key",
		ClientURL:   "http://localhost:5173",
	}
}

func supabaseSyncStub(
	t *testing.T,
	authStatus int, authBody interface{},
	existsBody interface{},
) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/v1/user":
			if r.Method == http.MethodGet {
				w.WriteHeader(authStatus)
				_ = json.NewEncoder(w).Encode(authBody)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		case "/rest/v1/users":
			switch r.Method {
			case http.MethodGet:
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(existsBody)
			case http.MethodPost:
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode([]interface{}{})
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		case "/rest/v1/user_progress":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode([]interface{}{})
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func validAuthBody(email string) map[string]interface{} {
	return map[string]interface{}{
		"id":    "uuid-001",
		"email": email,
		"user_metadata": map[string]interface{}{
			"full_name": "Ada Lovelace",
		},
	}
}

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

			if w.Code != tt.expectedStatus {
				t.Errorf("[%s] expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}

			var body map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if body["error"] != tt.expectedError {
				t.Errorf("[%s] expected error %q, got %q", tt.name, tt.expectedError, body["error"])
			}
		})
	}
}

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
				[]interface{}{},
			)
			h := newSyncHandler(stub.URL)

			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/sync", h.SyncUser)

			req, _ := http.NewRequest(http.MethodPost, "/sync", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("[%s] expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var body map[string]interface{}
				_ = json.Unmarshal(w.Body.Bytes(), &body)
				if body["error"] != tt.expectedError {
					t.Errorf("[%s] expected error %q, got %q", tt.name, tt.expectedError, body["error"])
				}
			}
		})
	}
}

func TestSyncUser_UserAlreadyExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := supabaseSyncStub(t,
		http.StatusOK, validAuthBody("user@focus.com"),
		[]interface{}{map[string]interface{}{"id": "uuid-001"}},
	)
	h := newSyncHandler(stub.URL)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/sync", h.SyncUser)

	req, _ := http.NewRequest(http.MethodPost, "/sync", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var body map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["synced"] != false {
		t.Errorf("expected synced=false, got %v", body["synced"])
	}
	if body["message"] != "usuario ya existe" {
		t.Errorf("expected message %q, got %q", "usuario ya existe", body["message"])
	}
}

func failuresStub(existsStatus, profileStatus, progressStatus int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/v1/user":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(validAuthBody("user@focus.com"))
		case "/rest/v1/users":
			if r.Method == http.MethodGet {
				w.WriteHeader(existsStatus)
				_ = json.NewEncoder(w).Encode([]interface{}{})
			} else {
				w.WriteHeader(profileStatus)
				_ = json.NewEncoder(w).Encode([]interface{}{})
			}
		case "/rest/v1/user_progress":
			w.WriteHeader(progressStatus)
			if progressStatus != http.StatusCreated {
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "failed to save initial progress"})
			}
		}
	}))
}

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
			expectedError:  "failed to save initial progress",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			srv := failuresStub(tt.existsStatus, tt.profileStatus, tt.progressStatus)
			defer srv.Close()

			h := newSyncHandler(srv.URL)
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/sync", h.SyncUser)

			req, _ := http.NewRequest(http.MethodPost, "/sync", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("[%s] expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}
			var body map[string]interface{}
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			errStr, _ := body["error"].(string)
			if !strings.Contains(errStr, tt.expectedError) {
				t.Errorf("[%s] expected error to contain %q, got %q", tt.name, tt.expectedError, errStr)
			}
		})
	}
}

func TestCreateUserProfileSync_FallbackError(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"something": "else"})
	}))
	defer stub.Close()
	h := newSyncHandler(stub.URL)
	err := h.createUserProfileSync("uuid", "test@test.com", "First", "Last")
	if err == nil {
		t.Error("expected error, got nil")
	} else if !strings.Contains(err.Error(), "failed to create profile (status 400)") {
		t.Errorf("expected fallback error, got %v", err)
	}
}

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
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if fn != tt.expectedFirstName {
				t.Errorf("expected first name %q, got %q", tt.expectedFirstName, fn)
			}
			if ln != tt.expectedLastName {
				t.Errorf("expected last name %q, got %q", tt.expectedLastName, ln)
			}
		})
	}
}

func TestFetchSupabaseUser_Errors(t *testing.T) {
	t.Run("Connection Error", func(t *testing.T) {
		h := newSyncHandler("http://localhost:9999")
		_, err := h.fetchSupabaseUser("token")
		if err == nil {
			t.Error("expected error, got nil")
		} else if !strings.Contains(err.Error(), "error connecting to auth service") {
			t.Errorf("expected connection error, got %v", err)
		}
	})

	t.Run("Decoding Error", func(t *testing.T) {
		stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{invalid-json"))
		}))
		defer stub.Close()
		h := newSyncHandler(stub.URL)
		_, err := h.fetchSupabaseUser("token")
		if err == nil {
			t.Error("expected error, got nil")
		} else if !strings.Contains(err.Error(), "error decoding user data") {
			t.Errorf("expected decoding error, got %v", err)
		}
	})
}

func TestUserExists_Errors(t *testing.T) {
	t.Run("Connection Error", func(t *testing.T) {
		h := newSyncHandler("http://localhost:9999")
		_, err := h.userExists("uuid")
		if err == nil {
			t.Error("expected error, got nil")
		} else if !strings.Contains(err.Error(), "error connecting to database") {
			t.Errorf("expected connection error, got %v", err)
		}
	})

	t.Run("Non-200 Status", func(t *testing.T) {
		stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer stub.Close()
		h := newSyncHandler(stub.URL)
		_, err := h.userExists("uuid")
		if err == nil {
			t.Error("expected error, got nil")
		} else if !strings.Contains(err.Error(), "error verifying user existence") {
			t.Errorf("expected status error, got %v", err)
		}
	})

	t.Run("Decoding Error", func(t *testing.T) {
		stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{invalid-json"))
		}))
		defer stub.Close()
		h := newSyncHandler(stub.URL)
		_, err := h.userExists("uuid")
		if err == nil {
			t.Error("expected error, got nil")
		} else if !strings.Contains(err.Error(), "error decoding database response") {
			t.Errorf("expected decoding error, got %v", err)
		}
	})
}

func TestCreateUserProfileSync_Errors(t *testing.T) {
	t.Run("Network Error", func(t *testing.T) {
		h := newSyncHandler("http://localhost:9999")
		err := h.createUserProfileSync("uuid", "test@test.com", "First", "Last")
		if err == nil {
			t.Error("expected error, got nil")
		} else if !strings.Contains(err.Error(), "network error while creating profile") {
			t.Errorf("expected network error, got %v", err)
		}
	})

	t.Run("Duplicate Error (23505)", func(t *testing.T) {
		stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "23505"})
		}))
		defer stub.Close()
		h := newSyncHandler(stub.URL)
		err := h.createUserProfileSync("uuid", "test@test.com", "First", "Last")
		if err != nil {
			t.Errorf("expected no error for duplicate, got %v", err)
		}
	})

	t.Run("Database Error with Message", func(t *testing.T) {
		stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "custom error"})
		}))
		defer stub.Close()
		h := newSyncHandler(stub.URL)
		err := h.createUserProfileSync("uuid", "test@test.com", "First", "Last")
		if err == nil {
			t.Error("expected error, got nil")
		} else if !strings.Contains(err.Error(), "database error: custom error") {
			t.Errorf("expected custom error message, got %v", err)
		}
	})

	t.Run("Decoding error message fails", func(t *testing.T) {
		stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("{invalid-json"))
		}))
		defer stub.Close()
		h := newSyncHandler(stub.URL)
		err := h.createUserProfileSync("uuid", "test@test.com", "First", "Last")
		if err == nil {
			t.Error("expected error, got nil")
		} else if !strings.Contains(err.Error(), "failed to create profile (status 400)") {
			t.Errorf("expected fallback error, got %v", err)
		}
	})
}

func TestSyncUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := supabaseSyncStub(t,
		http.StatusOK, validAuthBody("ada@focus.com"),
		[]interface{}{},
	)
	h := newSyncHandler(stub.URL)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/sync", h.SyncUser)

	req, _ := http.NewRequest(http.MethodPost, "/sync", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var body map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &body)

	if body["synced"] != true {
		t.Errorf("expected synced=true, got %v", body["synced"])
	}
	if body["id"] != "uuid-001" {
		t.Errorf("expected id %q, got %q", "uuid-001", body["id"])
	}
	if body["email"] != "ada@focus.com" {
		t.Errorf("expected email %q, got %q", "ada@focus.com", body["email"])
	}
	if body["first_name"] != "Ada" {
		t.Errorf("expected first_name %q, got %q", "Ada", body["first_name"])
	}
	if body["last_name"] != "Lovelace" {
		t.Errorf("expected last_name %q, got %q", "Lovelace", body["last_name"])
	}
}
