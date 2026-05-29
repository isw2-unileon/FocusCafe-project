package handlers

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
// /auth/v1/signup, /rest/v1/users and /rest/v1/user_progress.
// Si un body es un string directo (ej: "{bad-json"), se escribe directamente en la respuesta.
func supabaseMultiStub(
	t *testing.T,
	authStatus int, authBody interface{},
	profileStatus int, profileBody interface{},
	progressStatus int, progressBody interface{},
) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var status int
		var body interface{}

		switch r.URL.Path {
		case "/auth/v1/signup":
			status = authStatus
			body = authBody
		case "/rest/v1/users":
			status = profileStatus
			body = profileBody
		case "/rest/v1/user_progress":
			status = progressStatus
			body = progressBody
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(status)
		if strBody, ok := body.(string); ok {
			_, _ = w.Write([]byte(strBody))
		} else {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// newHandler inicializa el Handler original apuntando a la URL del stub
func newHandler(url string) *Handler {
	return &Handler{
		SupabaseURL: url,
		SupabaseKey: "test-key-de-mentira",
	}
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
			expectedError:  "invalid request body",
		},
		{
			name: "Missing first_name",
			body: map[string]string{
				"first_name": "", "last_name": "Lovelace",
				"email": "ada@focus.com", "password": "secret123", "confirm_password": "secret123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "error: first name and surname are required",
		},
		{
			name: "Missing last_name",
			body: map[string]string{
				"first_name": "Ada", "last_name": "",
				"email": "ada@focus.com", "password": "secret123", "confirm_password": "secret123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "error: first name and surname are required",
		},
		{
			name: "Missing email",
			body: map[string]string{
				"first_name": "Ada", "last_name": "Lovelace",
				"email": "", "password": "secret123", "confirm_password": "secret123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "error: email is mandatory",
		},
		{
			name: "Password too short (less than 6 chars)",
			body: map[string]string{
				"first_name": "Ada", "last_name": "Lovelace",
				"email": "ada@focus.com", "password": "abc", "confirm_password": "abc",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "error: the password must be at least 6 characters long",
		},
		{
			name: "Passwords do not match",
			body: map[string]string{
				"first_name": "Ada", "last_name": "Lovelace",
				"email": "ada@focus.com", "password": "secret123", "confirm_password": "different",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "error: the passwords do not match",
		},
		{
			name: "Fields with only whitespace trimmed to empty",
			body: map[string]string{
				"first_name": "   ", "last_name": "   ",
				"email": "ada@focus.com", "password": "secret123", "confirm_password": "secret123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "error: first name and surname are required",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			stub := supabaseMultiStub(t,
				http.StatusOK, successAuthBody,
				http.StatusCreated, successProfileBody,
				http.StatusCreated, []interface{}{},
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
		closeServer    bool // Fuerza un error de conexión de red
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
			expectedError:  "error: error creating the user",
		},
		{
			name:           "Supabase Auth returns 200 but missing user field",
			authStatus:     http.StatusOK,
			authBody:       map[string]interface{}{"something": "unexpected"},
			expectedStatus: http.StatusConflict,
			expectedError:  "error: unexpected response from Supabase Auth",
		},
		{
			name:           "Supabase Auth returns 200 but user has no id",
			authStatus:     http.StatusOK,
			authBody:       map[string]interface{}{"user": map[string]interface{}{"email": "ada@focus.com"}},
			expectedStatus: http.StatusConflict,
			expectedError:  "error: the user ID could not be retrieved",
		},
		{
			name:           "Supabase Auth returns corrupt JSON",
			authStatus:     http.StatusOK,
			authBody:       "{invalid-json-corrupt",
			expectedStatus: http.StatusConflict,
			expectedError:  "invalid character 'i' looking for beginning of object key string",
		},
		{
			name:           "Supabase Auth network connection error",
			authStatus:     http.StatusOK,
			authBody:       successAuthBody,
			closeServer:    true,
			expectedStatus: http.StatusConflict,
			expectedError:  "error: error connecting to Supabase Auth",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			stub := supabaseMultiStub(t,
				tt.authStatus, tt.authBody,
				http.StatusCreated, successProfileBody,
				http.StatusCreated, []interface{}{},
			)

			if tt.closeServer {
				stub.Close() // Cerramos el servidor antes del request para forzar error de red
			}
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
			assert.Contains(t, body["error"].(string), tt.expectedError, "[%s] unexpected error message", tt.name)
		})
	}
}

// ─────────────────────────────────────────────
// Register – Profile and Progress creation tests
// ─────────────────────────────────────────────

func TestRegister_UserProfileAndProgress_TableDriven(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		profileStatus  int
		profileBody    interface{}
		progressStatus int
		progressBody   interface{}
		closeServer    bool
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Profile insertion fails",
			profileStatus:  http.StatusInternalServerError,
			profileBody:    map[string]interface{}{"message": "db error"},
			progressStatus: http.StatusCreated,
			progressBody:   []interface{}{},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "error al guardar el perfil",
		},
		{
			name:           "Profile insertion returns corrupt JSON error",
			profileStatus:  http.StatusInternalServerError,
			profileBody:    "{corrupt-json",
			progressStatus: http.StatusCreated,
			progressBody:   []interface{}{},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "invalid character 'c' looking for beginning of object key string",
		},

		{
			name:           "Progress insertion fails",
			profileStatus:  http.StatusCreated,
			profileBody:    successProfileBody,
			progressStatus: http.StatusInternalServerError,
			progressBody:   map[string]interface{}{"message": "db error"},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "error: Error saving progress",
		},
		{
			name:           "Progress insertion returns corrupt JSON error",
			profileStatus:  http.StatusCreated,
			profileBody:    successProfileBody,
			progressStatus: http.StatusInternalServerError,
			progressBody:   "{corrupt-json",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "invalid character 'c' looking for beginning of object key string",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			stub := supabaseMultiStub(t,
				http.StatusOK, successAuthBody,
				tt.profileStatus, tt.profileBody,
				tt.progressStatus, tt.progressBody,
			)
			h := newHandler(stub.URL)

			if tt.closeServer {
				// Para simular caída de red justo en el perfil, destruimos el stub
				stub.Close()
			}

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
			assert.Contains(t, body["error"].(string), tt.expectedError, "[%s] unexpected error message", tt.name)
		})
	}
}

// ─────────────────────────────────────────────
// Register – Progress network failure specific
// ─────────────────────────────────────────────

func TestRegister_ProgressNetworkError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Stub que simula éxito inicial en Auth. No podemos cerrar el stub completo
	// porque mataría el flujo del perfil. Así que usaremos un truco: una URL rota para progreso modificando el handler.
	// Pero más directo: interceptamos el flujo llamando al método directamente.
	stub := supabaseMultiStub(t,
		http.StatusOK, successAuthBody,
		http.StatusCreated, successProfileBody,
		http.StatusCreated, []interface{}{},
	)
	h := newHandler(stub.URL)

	// Cambiamos la URL a algo inválido justo antes de ejecutar para simular caída de red en progreso
	h.SupabaseURL = "http://localhost:9999" // URL inexistente

	err := h.createUserProgress("uuid-test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al crear el progreso del usuario")
}

// ─────────────────────────────────────────────
// Register – Full success path
// ─────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := supabaseMultiStub(t,
		http.StatusOK, successAuthBody,
		http.StatusCreated, successProfileBody,
		http.StatusCreated, []interface{}{},
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

// ─────────────────────────────────────────────
// Direct Unit Tests (For isolated blocks)
// ─────────────────────────────────────────────

// TestCreateUserProfile_EmptyRole ejecuta directamente el método para cubrir el bloque `if role == ""`
func TestCreateUserProfile_EmptyRole(t *testing.T) {
	stub := supabaseMultiStub(t,
		http.StatusOK, successAuthBody,
		http.StatusCreated, successProfileBody,
		http.StatusCreated, []interface{}{},
	)
	h := newHandler(stub.URL)

	req := RegisterRequest{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     "ada@focus.com",
	}

	// Pasamos rol vacío "" para forzar la asignación por defecto `role = "user"`
	err := h.createUserProfile("uuid-ada-001", req, "")
	assert.NoError(t, err)
}
