package handlers_test

import (
	"context"
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
}

func (m *adminMockUserService) GetUserProfile(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	if m.getUserProfileFunc != nil {
		return m.getUserProfileFunc(ctx, id)
	}
	return nil, errors.New("getUserProfile not mocked")
}

func (m *adminMockUserService) UpdateUserProfile(ctx context.Context, id uuid.UUID, firstName, lastName string) error {
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
	return nil
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
			mockBehavior: func(ctx context.Context) ([]domain.UserProfile, error) {
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
			mockBehavior: func(ctx context.Context) ([]domain.UserProfile, error) {
				return []domain.UserProfile{}, nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `[]`,
		},
		{
			name: "Error: Service failure returns 500",
			mockBehavior: func(ctx context.Context) ([]domain.UserProfile, error) {
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
			mockBehavior: func(ctx context.Context, email string) (*domain.UserProfile, error) {
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
			mockBehavior: func(ctx context.Context, email string) (*domain.UserProfile, error) {
				return nil, nil // should not be called
			},
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `{"error":"email query parameter is required"}`,
		},
		{
			name:       "Error: User not found",
			queryEmail: "missing@test.com",
			mockBehavior: func(ctx context.Context, email string) (*domain.UserProfile, error) {
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
			mockBehavior: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
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
			mockBehavior: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
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
			mockBehavior: func(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
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
			mockBehavior: func(ctx context.Context) ([]domain.GroupDetail, error) {
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
			mockBehavior: func(ctx context.Context) ([]domain.GroupDetail, error) {
				return []domain.GroupDetail{}, nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `[]`,
		},
		{
			name: "Error: Service failure returns 500",
			mockBehavior: func(ctx context.Context) ([]domain.GroupDetail, error) {
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
			mockBehavior: func(ctx context.Context, groupID int64) error {
				return nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `{"message":"group deleted successfully"}`,
		},
		{
			name:         "Error: Invalid group ID returns 400",
			groupIDParam: "abc",
			mockBehavior: nil,
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid group id format"}`,
		},
		{
			name:         "Error: Service failure returns 500",
			groupIDParam: "99",
			mockBehavior: func(ctx context.Context, groupID int64) error {
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
