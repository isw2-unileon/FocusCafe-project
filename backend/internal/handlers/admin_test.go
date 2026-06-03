package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
)

// adminMockUserService is a test double for the UserServiceInterface.
// It mirrors the mock from user_test.go but is kept here to avoid cross-file dependencies.
type adminMockUserService struct {
	getUserProfileFunc func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
	getAllUsersFunc    func(ctx context.Context) ([]domain.UserProfile, error)
	getUserByEmailFunc func(ctx context.Context, email string) (*domain.UserProfile, error)
	deleteUserFunc     func(ctx context.Context, id uuid.UUID) error
}

func (m *adminMockUserService) GetUserProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	if m.getUserProfileFunc != nil {
		return m.getUserProfileFunc(ctx, id)
	}
	return nil, errors.New("getUserProfile not mocked")
}

func (m *adminMockUserService) UpdateUserProfile(_ context.Context, _ uuid.UUID, _, _ string) error {
	return nil
}

func (m *adminMockUserService) GetAllUsers(ctx context.Context) ([]domain.UserProfile, error) {
	if m.getAllUsersFunc != nil {
		return m.getAllUsersFunc(ctx)
	}
	return nil, nil
}

func (m *adminMockUserService) GetUserByEmail(ctx context.Context, email string) (*domain.UserProfile, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *adminMockUserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if m.deleteUserFunc != nil {
		return m.deleteUserFunc(ctx, id)
	}
	return nil
}

func (m *adminMockUserService) GetLeaderboard(_ context.Context, _ int) ([]domain.UserProfile, error) {
	return nil, nil
}

func (m *adminMockUserService) GetUserLeaderboard(_ context.Context, _ uuid.UUID) (int, *domain.UserProfile, error) {
	return 0, nil, nil
}

// ============================================
// TestHandler_GetAllUsers
// ============================================

func TestHandler_GetAllUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mockBehavior   func(ctx context.Context) ([]domain.UserProfile, error)
		wantStatusCode int
		expectedBody   string
	}{
		{
			name: "Success: Returns list of users",
			mockBehavior: func(_ context.Context) ([]domain.UserProfile, error) {
				return []domain.UserProfile{
					{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), FirstName: "Alice", Email: "alice@test.com", Role: "user"},
					{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), FirstName: "Bob", Email: "bob@test.com", Role: "admin"},
				}, nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `"first_name":"Alice"`,
		},
		{
			name: "Success: Returns empty list",
			mockBehavior: func(_ context.Context) ([]domain.UserProfile, error) {
				return []domain.UserProfile{}, nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `[]`,
		},
		{
			name: "Error: Service failure returns 500",
			mockBehavior: func(_ context.Context) ([]domain.UserProfile, error) {
				return nil, errors.New("database connection lost")
			},
			wantStatusCode: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to fetch users"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mService := &adminMockUserService{getAllUsersFunc: tt.mockBehavior}
			h := &handlers.Handler{UserService: mService}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/api/admin/users", nil)

			h.GetAllUsers(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler.GetAllUsers() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			gotBody := w.Body.String()
			if !strings.Contains(gotBody, tt.expectedBody) {
				t.Errorf("Handler.GetAllUsers() body = %v, want to contain %v", gotBody, tt.expectedBody)
			}
		})
	}
}

// ============================================
// TestHandler_GetUserByEmail
// ============================================

func TestHandler_GetUserByEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name           string
		queryEmail     string
		mockBehavior   func(ctx context.Context, email string) (*domain.UserProfile, error)
		wantStatusCode int
		expectedBody   string
	}{
		{
			name:       "Success: User found by email",
			queryEmail: "alice@test.com",
			mockBehavior: func(_ context.Context, _ string) (*domain.UserProfile, error) {
				return &domain.UserProfile{
					ID:        testUUID,
					FirstName: "Alice",
					LastName:  "Smith",
					Email:     "alice@test.com",
					Role:      "admin",
				}, nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `"email":"alice@test.com"`,
		},
		{
			name:       "Error: Email query parameter is empty",
			queryEmail: "",
			mockBehavior: func(_ context.Context, _ string) (*domain.UserProfile, error) {
				return nil, nil // should not be called
			},
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `{"error":"email query parameter is required"}`,
		},
		{
			name:       "Error: User not found",
			queryEmail: "missing@test.com",
			mockBehavior: func(_ context.Context, _ string) (*domain.UserProfile, error) {
				return nil, errors.New("record not found")
			},
			wantStatusCode: http.StatusNotFound,
			expectedBody:   `{"error":"user not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mService := &adminMockUserService{getUserByEmailFunc: tt.mockBehavior}
			h := &handlers.Handler{UserService: mService}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/api/admin/users/search?email="+tt.queryEmail, nil)

			h.GetUserByEmail(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler.GetUserByEmail() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			gotBody := w.Body.String()
			if !strings.Contains(gotBody, tt.expectedBody) {
				t.Errorf("Handler.GetUserByEmail() body = %v, want to contain %v", gotBody, tt.expectedBody)
			}
		})
	}
}

// ============================================
// TestAdminOnly_Middleware
// ============================================

func TestAdminOnly_Middleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	adminUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

	tests := []struct {
		name           string
		setClaims      bool
		claimsID       string
		role           string
		mockBehavior   func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error)
		wantStatusCode int
		expectedBody   string
	}{
		{
			name:      "Success: Admin user passes middleware",
			setClaims: true,
			claimsID:  adminUUID.String(),
			role:      "admin",
			mockBehavior: func(_ context.Context, _ uuid.UUID) (*domain.UserProfile, error) {
				return &domain.UserProfile{ID: adminUUID, Role: "admin"}, nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `{"message":"passed"}`,
		},
		{
			name:      "Forbidden: Non-admin user is rejected",
			setClaims: true,
			claimsID:  userUUID.String(),
			role:      "user",
			mockBehavior: func(_ context.Context, _ uuid.UUID) (*domain.UserProfile, error) {
				return &domain.UserProfile{ID: userUUID, Role: "user"}, nil
			},
			wantStatusCode: http.StatusForbidden,
			expectedBody:   `{"error":"admin access required"}`,
		},
		{
			name:           "Forbidden: No claims in context",
			setClaims:      false,
			wantStatusCode: http.StatusForbidden,
			expectedBody:   `{"error":"forbidden"}`,
		},
		{
			name:      "Forbidden: User not found in database",
			setClaims: true,
			claimsID:  userUUID.String(),
			role:      "user",
			mockBehavior: func(_ context.Context, _ uuid.UUID) (*domain.UserProfile, error) {
				return nil, errors.New("record not found")
			},
			wantStatusCode: http.StatusForbidden,
			expectedBody:   `{"error":"user not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mService := &adminMockUserService{getUserProfileFunc: tt.mockBehavior}
			h := &handlers.Handler{UserService: mService}

			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)

			// fakeAuth middleware simulates the real Auth middleware by injecting claims into the context
			fakeAuth := func(ctx *gin.Context) {
				if tt.setClaims {
					claims := &auth.UserClaims{
						RegisteredClaims: jwt.RegisteredClaims{
							Subject: tt.claimsID,
						},
					}
					ctx.Set("user", claims)
				}
				ctx.Next()
			}

			// Register: fakeAuth → AdminOnly → protected handler
			r.GET("/admin-only", fakeAuth, h.AdminOnly(), func(ctx *gin.Context) {
				ctx.JSON(http.StatusOK, gin.H{"message": "passed"})
			})

			req := httptest.NewRequest("GET", "/admin-only", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("AdminOnly middleware status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			gotBody := w.Body.String()
			if !strings.Contains(gotBody, tt.expectedBody) {
				t.Errorf("AdminOnly middleware body = %v, want to contain %v", gotBody, tt.expectedBody)
			}
		})
	}
}

// ============================================
// TestHandler_GetAllGroups (Admin)
// ============================================

func TestHandler_GetAllGroups(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		mockBehavior   func(ctx context.Context) ([]domain.GroupDetail, error)
		wantStatusCode int
		expectedBody   string
	}{
		{
			name: "Success: Returns list of groups with members",
			mockBehavior: func(_ context.Context) ([]domain.GroupDetail, error) {
				return []domain.GroupDetail{
					{
						ID:         1,
						Name:       "The A-Team",
						InviteCode: "AB12CD",
						LeaderID:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
						Members: []domain.GroupMember{
							{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), FirstName: "Alice", Email: "alice@test.com", Level: 5},
							{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), FirstName: "Bob", Email: "bob@test.com", Level: 3},
						},
					},
				}, nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `"name":"The A-Team"`,
		},
		{
			name: "Success: Returns empty list",
			mockBehavior: func(_ context.Context) ([]domain.GroupDetail, error) {
				return []domain.GroupDetail{}, nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `[]`,
		},
		{
			name: "Error: Service failure returns 500",
			mockBehavior: func(_ context.Context) ([]domain.GroupDetail, error) {
				return nil, errors.New("database connection lost")
			},
			wantStatusCode: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to fetch groups"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mService := &mockGroupService{getAllGroupsFunc: tt.mockBehavior}
			h := &handlers.Handler{GroupService: mService}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/api/admin/groups", nil)

			h.GetAllGroups(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler.GetAllGroups() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			gotBody := w.Body.String()
			if !strings.Contains(gotBody, tt.expectedBody) {
				t.Errorf("Handler.GetAllGroups() body = %v, want to contain %v", gotBody, tt.expectedBody)
			}
		})
	}
}

// ============================================
// TestHandler_AdminDeleteGroup
// ============================================

func TestHandler_AdminDeleteGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		groupIDParam   string
		mockBehavior   func(ctx context.Context, groupID int64) error
		wantStatusCode int
		expectedBody   string
	}{
		{
			name:         "Success: Deletes group and returns 200",
			groupIDParam: "1",
			mockBehavior: func(_ context.Context, _ int64) error {
				return nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `{"message":"group deleted successfully"}`,
		},
		{
			name:           "Error: Invalid group ID returns 400",
			groupIDParam:   "abc",
			mockBehavior:   nil,
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid group id format"}`,
		},
		{
			name:         "Error: Service failure returns 500",
			groupIDParam: "99",
			mockBehavior: func(_ context.Context, _ int64) error {
				return errors.New("group not found")
			},
			wantStatusCode: http.StatusInternalServerError,
			expectedBody:   `{"error":"group not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mService := &mockGroupService{deleteGroupFunc: tt.mockBehavior}
			h := &handlers.Handler{GroupService: mService}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.groupIDParam}}
			c.Request = httptest.NewRequest("DELETE", "/api/admin/groups/"+tt.groupIDParam, nil)

			h.AdminDeleteGroup(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler.AdminDeleteGroup() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			gotBody := w.Body.String()
			if !strings.Contains(gotBody, tt.expectedBody) {
				t.Errorf("Handler.AdminDeleteGroup() body = %v, want to contain %v", gotBody, tt.expectedBody)
			}
		})
	}
}

// ============================================
// TestHandler_AdminCreateUser
// ============================================

func TestHandler_AdminCreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    handlers.AdminCreateUserRequest
		wantStatusCode int
		expectedBody   string
	}{
		{
			name: "Success: Creates user",
			requestBody: handlers.AdminCreateUserRequest{
				FirstName: "John", LastName: "Doe", Email: "john@test.com",
				Password: "password123", ConfirmPassword: "password123", Role: "user",
			},
			wantStatusCode: http.StatusCreated,
			expectedBody:   `"email":"john@test.com"`,
		},
		{
			name: "Error: Missing names",
			requestBody: handlers.AdminCreateUserRequest{
				Email: "john@test.com", Password: "password123", ConfirmPassword: "password123",
			},
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `{"error":"first name and last name are required"}`,
		},
		{
			name: "Error: Missing email",
			requestBody: handlers.AdminCreateUserRequest{
				FirstName: "John", LastName: "Doe",
				Password: "password123", ConfirmPassword: "password123",
			},
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `{"error":"email is required"}`,
		},
		{
			name: "Error: Password too short",
			requestBody: handlers.AdminCreateUserRequest{
				FirstName: "John", LastName: "Doe", Email: "john@test.com",
				Password: "123", ConfirmPassword: "123",
			},
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `{"error":"password must be at least 6 characters long"}`,
		},
		{
			name: "Error: Passwords do not match",
			requestBody: handlers.AdminCreateUserRequest{
				FirstName: "John", LastName: "Doe", Email: "john@test.com",
				Password: "password123", ConfirmPassword: "wrongpassword",
			},
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `{"error":"passwords do not match"}`,
		},
		{
			name: "Error: Supabase Auth failure",
			requestBody: handlers.AdminCreateUserRequest{
				FirstName: "John", LastName: "Doe", Email: "error-auth@test.com",
				Password: "password123", ConfirmPassword: "password123",
			},
			wantStatusCode: http.StatusConflict,
			expectedBody:   `auth error`,
		},
		{
			name: "Error: Supabase Profile failure",
			requestBody: handlers.AdminCreateUserRequest{
				FirstName: "John", LastName: "Doe", Email: "error-profile@test.com",
				Password: "password123", ConfirmPassword: "password123",
			},
			wantStatusCode: http.StatusInternalServerError,
			expectedBody:   `profile error`,
		},
		{
			name: "Error: Supabase Progress failure",
			requestBody: handlers.AdminCreateUserRequest{
				FirstName: "John", LastName: "Doe", Email: "error-progress@test.com",
				Password: "password123", ConfirmPassword: "password123",
			},
			wantStatusCode: http.StatusInternalServerError,
			expectedBody:   `progress error`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			localServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				path := r.URL.Path
				switch {
				case strings.Contains(path, "/auth/v1/signup"):
					if tt.requestBody.Email == "error-auth@test.com" {
						w.WriteHeader(http.StatusBadRequest)
						_ = json.NewEncoder(w).Encode(map[string]any{"message": "auth error"})
						return
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"user": map[string]any{"id": "550e8400-e29b-41d4-a716-446655440000"},
					})
				case strings.Contains(path, "/rest/v1/users"):
					if tt.requestBody.Email == "error-profile@test.com" {
						w.WriteHeader(http.StatusInternalServerError)
						_ = json.NewEncoder(w).Encode(map[string]any{"message": "profile error"})
						return
					}
					w.WriteHeader(http.StatusCreated)
				case strings.Contains(path, "/rest/v1/user_progress"):
					if tt.requestBody.Email == "error-progress@test.com" {
						w.WriteHeader(http.StatusInternalServerError)
						_ = json.NewEncoder(w).Encode(map[string]any{"message": "progress error"})
						return
					}
					w.WriteHeader(http.StatusCreated)
				}
			}))
			defer localServer.Close()

			h := &handlers.Handler{SupabaseURL: localServer.URL, SupabaseKey: "test-key"}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			body, _ := json.Marshal(tt.requestBody)
			c.Request = httptest.NewRequest("POST", "/api/admin/users", bytes.NewBuffer(body))

			h.AdminCreateUser(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("%s: status = %v, want %v", tt.name, w.Code, tt.wantStatusCode)
			}
			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("%s: body = %v, want %v", tt.name, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

// ============================================
// TestHandler_DeleteUser
// ============================================

func TestHandler_DeleteUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userIDParam    string
		supabaseResp   int
		roleKey        string
		mockBehavior   func(ctx context.Context, id uuid.UUID) error
		wantStatusCode int
		expectedBody   string
	}{
		{
			name:         "Success: Deletes user",
			userIDParam:  "550e8400-e29b-41d4-a716-446655440000",
			supabaseResp: http.StatusOK,
			roleKey:      "test-role-key",
			mockBehavior: func(_ context.Context, _ uuid.UUID) error {
				return nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `{"message":"user deleted successfully"}`,
		},
		{
			name:           "Error: Invalid UUID",
			userIDParam:    "invalid-uuid",
			roleKey:        "test-role-key",
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid user id format"}`,
		},
		{
			name:           "Error: Supabase Auth failure",
			userIDParam:    "550e8400-e29b-41d4-a716-446655440000",
			supabaseResp:   http.StatusInternalServerError,
			roleKey:        "test-role-key",
			wantStatusCode: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to delete user from authentication provider`,
		},
		{
			name:         "Error: Local DB failure",
			userIDParam:  "550e8400-e29b-41d4-a716-446655440000",
			supabaseResp: http.StatusOK,
			roleKey:      "test-role-key",
			mockBehavior: func(_ context.Context, _ uuid.UUID) error {
				return errors.New("db error")
			},
			wantStatusCode: http.StatusInternalServerError,
			expectedBody:   `{"error":"failed to delete user from database"}`,
		},
		{
			name:           "Error: Role key missing",
			userIDParam:    "550e8400-e29b-41d4-a716-446655440000",
			roleKey:        "",
			wantStatusCode: http.StatusInternalServerError,
			expectedBody:   `SUPABASE_SERVICE_ROLE_KEY not configured`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.supabaseResp)
			}))
			defer server.Close()

			mService := &adminMockUserService{deleteUserFunc: tt.mockBehavior}
			h := &handlers.Handler{
				UserService:            mService,
				SupabaseURL:            server.URL,
				SupabaseServiceRoleKey: tt.roleKey,
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: tt.userIDParam}}
			c.Request = httptest.NewRequest("DELETE", "/api/admin/users/"+tt.userIDParam, nil)

			h.DeleteUser(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("%s: Handler.DeleteUser() status = %v, want %v", tt.name, w.Code, tt.wantStatusCode)
			}
			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("%s: Handler.DeleteUser() body = %v, want to contain %v", tt.name, w.Body.String(), tt.expectedBody)
			}
		})
	}
}
