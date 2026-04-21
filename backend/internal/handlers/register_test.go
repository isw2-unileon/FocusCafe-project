package handlers_test

import (
	"bytes"
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

// supabaseMultiStub starts an httptest.Server that routes requests to
// /auth/v1/signup and /rest/v1/users to their respective handlers.
func supabaseMultiStub(
	t *testing.T,
	authStatus int, authBody interface{},
	profileStatus int, profileBody interface{},
) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/auth/v1/signup":
			w.WriteHeader(authStatus)
			_ = json.NewEncoder(w).Encode(authBody)
		case "/rest/v1/users":
			w.WriteHeader(profileStatus)
			_ = json.NewEncoder(w).Encode(profileBody)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// registerBody builds a JSON body for a RegisterRequest.
func registerBody(t *testing.T, fields map[string]string) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("registerBody: %v", err)
	}
	return bytes.NewBuffer(b)
}

// validRegisterFields returns a complete, valid RegisterRequest payload.
func validRegisterFields() map[string]string {
	return map[string]string{
		"first_name":       "Ada",
		"last_name":        "Lovelace",
		"email":            "ada@focus.com",
		"password":         "secret123",
		"confirm_password": "secret123",
	}
}

// successAuthBody is a typical Supabase Auth /signup success response.
var successAuthBody = map[string]interface{}{
	"user": map[string]interface{}{
		"id":    "uuid-ada-001",
		"email": "ada@focus.com",
	},
}

// successProfileBody is a typical Supabase REST /users 201 response.
var successProfileBody = []interface{}{
	map[string]interface{}{"id": "uuid-ada-001"},
}

// ─────────────────────────────────────────────
// Register – validation tests (no Supabase call needed)
// ─────────────────────────────────────────────

func TestRegister_Validation_TableDriven(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Malformed JSON body",
			body:           "not-json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "cuerpo de solicitud inválido",
		},
		{
			name: "Missing first_name",
			body: map[string]string{
				"first_name": "", "last_name": "Lovelace",
				"email": "ada@focus.com", "password": "secret123", "confirm_password": "secret123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "nombre y apellido son obligatorios",
		},
		{
			name: "Missing last_name",
			body: map[string]string{
				"first_name": "Ada", "last_name": "",
				"email": "ada@focus.com", "password": "secret123", "confirm_password": "secret123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "nombre y apellido son obligatorios",
		},
		{
			name: "Missing email",
			body: map[string]string{
				"first_name": "Ada", "last_name": "Lovelace",
				"email": "", "password": "secret123", "confirm_password": "secret123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "email es obligatorio",
		},
		{
			name: "Password too short (less than 6 chars)",
			body: map[string]string{
				"first_name": "Ada", "last_name": "Lovelace",
				"email": "ada@focus.com", "password": "abc", "confirm_password": "abc",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "la contraseña debe tener al menos 6 caracteres",
		},
		{
			name: "Passwords do not match",
			body: map[string]string{
				"first_name": "Ada", "last_name": "Lovelace",
				"email": "ada@focus.com", "password": "secret123", "confirm_password": "different",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "las contraseñas no coinciden",
		},
		{
			name: "Fields with only whitespace trimmed to empty",
			body: map[string]string{
				"first_name": "   ", "last_name": "   ",
				"email": "ada@focus.com", "password": "secret123", "confirm_password": "secret123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "nombre y apellido son obligatorios",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Stub that should never be reached for validation errors
			stub := supabaseMultiStub(t,
				http.StatusOK, successAuthBody,
				http.StatusCreated, successProfileBody,
			)
			h := newHandler(stub.URL)

			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/register", h.Register)

			var buf *bytes.Buffer
			switch v := tt.body.(type) {
			case string:
				buf = bytes.NewBufferString(v)
			default:
				b, _ := json.Marshal(v)
				buf = bytes.NewBuffer(b)
			}

			req, _ := http.NewRequest(http.MethodPost, "/register", buf)
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("[%s] expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}

			var body map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("[%s] could not parse response body: %v", tt.name, err)
			}
			assert.Equal(t, tt.expectedError, body["error"], "[%s] unexpected error message", tt.name)
		})
	}
}

// ─────────────────────────────────────────────
// Register – Supabase Auth interaction tests
// ─────────────────────────────────────────────

func TestRegister_AuthUser_TableDriven(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authStatus     int
		authBody       interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Supabase Auth returns non-200 with msg",
			authStatus:     http.StatusUnprocessableEntity,
			authBody:       map[string]interface{}{"msg": "User already registered"},
			expectedStatus: http.StatusConflict,
			expectedError:  "User already registered",
		},
		{
			name:           "Supabase Auth returns non-200 without msg",
			authStatus:     http.StatusInternalServerError,
			authBody:       map[string]interface{}{},
			expectedStatus: http.StatusConflict,
			expectedError:  "error al crear el usuario",
		},
		{
			name:           "Supabase Auth returns 200 but missing user field",
			authStatus:     http.StatusOK,
			authBody:       map[string]interface{}{"something": "unexpected"},
			expectedStatus: http.StatusConflict,
			expectedError:  "respuesta inesperada de Supabase Auth",
		},
		{
			name:           "Supabase Auth returns 200 but user has no id",
			authStatus:     http.StatusOK,
			authBody:       map[string]interface{}{"user": map[string]interface{}{"email": "ada@focus.com"}},
			expectedStatus: http.StatusConflict,
			expectedError:  "no se pudo obtener el ID del usuario",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			stub := supabaseMultiStub(t,
				tt.authStatus, tt.authBody,
				http.StatusCreated, successProfileBody, // never reached
			)
			h := newHandler(stub.URL)

			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/register", h.Register)

			req, _ := http.NewRequest(http.MethodPost, "/register",
				registerBody(t, validRegisterFields()))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("[%s] expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}

			var body map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("[%s] could not parse response body: %v", tt.name, err)
			}
			assert.Equal(t, tt.expectedError, body["error"], "[%s] unexpected error message", tt.name)
		})
	}
}

// ─────────────────────────────────────────────
// Register – Profile creation tests
// ─────────────────────────────────────────────

func TestRegister_UserProfile_TableDriven(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		profileStatus  int
		profileBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Profile insertion fails",
			profileStatus:  http.StatusInternalServerError,
			profileBody:    map[string]interface{}{"message": "db error"},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "error al guardar el perfil",
		},
		{
			name:           "Profile insertion returns unexpected status",
			profileStatus:  http.StatusConflict,
			profileBody:    map[string]interface{}{"message": "duplicate key"},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "error al guardar el perfil",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			stub := supabaseMultiStub(t,
				http.StatusOK, successAuthBody,
				tt.profileStatus, tt.profileBody,
			)
			h := newHandler(stub.URL)

			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.POST("/register", h.Register)

			req, _ := http.NewRequest(http.MethodPost, "/register",
				registerBody(t, validRegisterFields()))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("[%s] expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
			}

			var body map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("[%s] could not parse response body: %v", tt.name, err)
			}
			assert.Equal(t, tt.expectedError, body["error"], "[%s] unexpected error message", tt.name)
		})
	}
}

// ─────────────────────────────────────────────
// Register – Full success path
// ─────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := supabaseMultiStub(t,
		http.StatusOK, successAuthBody,
		http.StatusCreated, successProfileBody,
	)
	h := newHandler(stub.URL)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)
	r.POST("/register", h.Register)

	req, _ := http.NewRequest(http.MethodPost, "/register",
		registerBody(t, validRegisterFields()))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d — body: %s", w.Code, w.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("could not parse response body: %v", err)
	}

	assert.Equal(t, "uuid-ada-001", body["id"])
	assert.Equal(t, "ada@focus.com", body["email"])
	assert.Equal(t, "Ada", body["first_name"])
	assert.Equal(t, "Lovelace", body["last_name"])
}
