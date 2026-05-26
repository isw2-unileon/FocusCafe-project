package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
)

// mockGroupService is a test double for the GroupServiceInterface.
type mockGroupService struct {
	createGroupFunc  func(ctx context.Context, name string, leaderID uuid.UUID) (*domain.Group, error)
	joinGroupFunc    func(ctx context.Context, inviteCode string, userID uuid.UUID) (*domain.Group, error)
	getAllGroupsFunc func(ctx context.Context) ([]domain.GroupDetail, error)
	deleteGroupFunc  func(ctx context.Context, groupID int64) error
	leaveGroupFunc   func(ctx context.Context, userID uuid.UUID) error
	getUserGroupFunc func(ctx context.Context, userID uuid.UUID) (*domain.Group, error)
}

func (m *mockGroupService) CreateGroup(ctx context.Context, name string, leaderID uuid.UUID) (*domain.Group, error) {
	if m.createGroupFunc != nil {
		return m.createGroupFunc(ctx, name, leaderID)
	}
	return nil, errors.New("createGroup not mocked")
}

func (m *mockGroupService) JoinGroup(ctx context.Context, inviteCode string, userID uuid.UUID) (*domain.Group, error) {
	if m.joinGroupFunc != nil {
		return m.joinGroupFunc(ctx, inviteCode, userID)
	}
	return nil, errors.New("joinGroup not mocked")
}

func (m *mockGroupService) GetAllGroups(ctx context.Context) ([]domain.GroupDetail, error) {
	if m.getAllGroupsFunc != nil {
		return m.getAllGroupsFunc(ctx)
	}
	return nil, errors.New("getAllGroups not mocked")
}

func (m *mockGroupService) DeleteGroup(ctx context.Context, groupID int64) error {
	if m.deleteGroupFunc != nil {
		return m.deleteGroupFunc(ctx, groupID)
	}
	return errors.New("deleteGroup not mocked")
}

func (m *mockGroupService) LeaveGroup(ctx context.Context, userID uuid.UUID) error {
	if m.leaveGroupFunc != nil {
		return m.leaveGroupFunc(ctx, userID)
	}
	return errors.New("leaveGroup not mocked")
}

func (m *mockGroupService) GetUserGroup(ctx context.Context, userID uuid.UUID) (*domain.Group, error) {
	if m.getUserGroupFunc != nil {
		return m.getUserGroupFunc(ctx, userID)
	}
	return nil, errors.New("getUserGroup not mocked")
}

// helper to inject JWT claims into gin context
func injectUserClaims(c *gin.Context, userID uuid.UUID) {
	claims := &auth.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID.String(),
		},
	}
	c.Set("user", claims)
}

// ============================================
// TestHandler_CreateGroup
// ============================================

func TestHandler_CreateGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name            string
		body            string
		userIDInContext uuid.UUID
		mockBehavior    func(ctx context.Context, name string, leaderID uuid.UUID) (*domain.Group, error)
		wantStatusCode  int
		expectedBody    string
	}{
		{
			name:            "Success: Creates group and returns 201",
			body:            `{"name":"The A-Team"}`,
			userIDInContext: testUUID,
			mockBehavior: func(ctx context.Context, name string, leaderID uuid.UUID) (*domain.Group, error) {
				return &domain.Group{
					ID:         1,
					Name:       name,
					InviteCode: "AB12CD",
					LeaderID:   leaderID,
					CreatedAt:  time.Now(),
				}, nil
			},
			wantStatusCode: http.StatusCreated,
			expectedBody:   `"invite_code":"AB12CD"`,
		},
		{
			name:            "Error: Missing name returns 400",
			body:            `{"name":""}`,
			userIDInContext: testUUID,
			mockBehavior:    nil,
			wantStatusCode:  http.StatusBadRequest,
			expectedBody:    `{"error":"group name is required"}`,
		},
		{
			name:            "Error: User already in group returns 400",
			body:            `{"name":"The B-Team"}`,
			userIDInContext: testUUID,
			mockBehavior: func(ctx context.Context, name string, leaderID uuid.UUID) (*domain.Group, error) {
				return nil, errors.New("user is already in a group")
			},
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `{"error":"user is already in a group"}`,
		},
		{
			name:            "Error: No user in context returns 401",
			body:            `{"name":"The C-Team"}`,
			userIDInContext: uuid.Nil,
			mockBehavior:    nil,
			wantStatusCode:  http.StatusUnauthorized,
			expectedBody:    `{"error":"unauthorized"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mService := &mockGroupService{createGroupFunc: tt.mockBehavior}
			h := &handlers.Handler{GroupService: mService}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.userIDInContext != uuid.Nil {
				injectUserClaims(c, tt.userIDInContext)
			}

			c.Request = httptest.NewRequest("POST", "/api/groups", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			h.CreateGroup(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler.CreateGroup() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Handler.CreateGroup() body = %v, want to contain %v", w.Body.String(), tt.expectedBody)
			}
		})
	}
}

// ============================================
// TestHandler_JoinGroup
// ============================================

func TestHandler_JoinGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

	tests := []struct {
		name            string
		body            string
		userIDInContext uuid.UUID
		mockBehavior    func(ctx context.Context, inviteCode string, userID uuid.UUID) (*domain.Group, error)
		wantStatusCode  int
		expectedBody    string
	}{
		{
			name:            "Success: Joins group and returns 200",
			body:            `{"invite_code":"AB12CD"}`,
			userIDInContext: testUUID,
			mockBehavior: func(ctx context.Context, inviteCode string, userID uuid.UUID) (*domain.Group, error) {
				return &domain.Group{
					ID:         1,
					Name:       "The A-Team",
					InviteCode: inviteCode,
					LeaderID:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
				}, nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `"name":"The A-Team"`,
		},
		{
			name:            "Error: Invalid invite code returns 400",
			body:            `{"invite_code":"ZZZZZZ"}`,
			userIDInContext: testUUID,
			mockBehavior: func(ctx context.Context, inviteCode string, userID uuid.UUID) (*domain.Group, error) {
				return nil, errors.New("invalid invite code: ZZZZZZ")
			},
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid invite code: ZZZZZZ"}`,
		},
		{
			name:            "Error: User already in group returns 400",
			body:            `{"invite_code":"AB12CD"}`,
			userIDInContext: testUUID,
			mockBehavior: func(ctx context.Context, inviteCode string, userID uuid.UUID) (*domain.Group, error) {
				return nil, errors.New("user is already in a group")
			},
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `{"error":"user is already in a group"}`,
		},
		{
			name:            "Error: No user in context returns 401",
			body:            `{"invite_code":"AB12CD"}`,
			userIDInContext: uuid.Nil,
			mockBehavior:    nil,
			wantStatusCode:  http.StatusUnauthorized,
			expectedBody:    `{"error":"unauthorized"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mService := &mockGroupService{joinGroupFunc: tt.mockBehavior}
			h := &handlers.Handler{GroupService: mService}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.userIDInContext != uuid.Nil {
				injectUserClaims(c, tt.userIDInContext)
			}

			c.Request = httptest.NewRequest("POST", "/api/groups/join", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			h.JoinGroup(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler.JoinGroup() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Handler.JoinGroup() body = %v, want to contain %v", w.Body.String(), tt.expectedBody)
			}
		})
	}
}

// ============================================
// TestHandler_LeaveGroup
// ============================================

func TestHandler_LeaveGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")

	tests := []struct {
		name            string
		userIDInContext uuid.UUID
		mockBehavior    func(ctx context.Context, userID uuid.UUID) error
		wantStatusCode  int
		expectedBody    string
	}{
		{
			name:            "Success: Leaves group and returns 200",
			userIDInContext: testUUID,
			mockBehavior: func(ctx context.Context, userID uuid.UUID) error {
				return nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `{"message":"left group successfully"}`,
		},
		{
			name:            "Error: User is leader returns 400",
			userIDInContext: testUUID,
			mockBehavior: func(ctx context.Context, userID uuid.UUID) error {
				return errors.New("group leader cannot leave the group, use delete instead")
			},
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `{"error":"group leader cannot leave the group, use delete instead"}`,
		},
		{
			name:            "Error: User not in any group returns 400",
			userIDInContext: testUUID,
			mockBehavior: func(ctx context.Context, userID uuid.UUID) error {
				return errors.New("user is not in any group")
			},
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `{"error":"user is not in any group"}`,
		},
		{
			name:            "Error: No user in context returns 401",
			userIDInContext: uuid.Nil,
			mockBehavior:    nil,
			wantStatusCode:  http.StatusUnauthorized,
			expectedBody:    `{"error":"unauthorized"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mService := &mockGroupService{leaveGroupFunc: tt.mockBehavior}
			h := &handlers.Handler{GroupService: mService}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.userIDInContext != uuid.Nil {
				injectUserClaims(c, tt.userIDInContext)
			}

			c.Request = httptest.NewRequest("POST", "/api/groups/leave", nil)

			h.LeaveGroup(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler.LeaveGroup() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Handler.LeaveGroup() body = %v, want to contain %v", w.Body.String(), tt.expectedBody)
			}
		})
	}
}

// ============================================
// TestHandler_DeleteGroup
// ============================================

func TestHandler_DeleteGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	leaderUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")
	memberUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440004")

	tests := []struct {
		name             string
		userIDInContext  uuid.UUID
		mockGetGroupFunc func(ctx context.Context, userID uuid.UUID) (*domain.Group, error)
		mockDeleteFunc   func(ctx context.Context, groupID int64) error
		wantStatusCode   int
		expectedBody     string
	}{
		{
			name:            "Success: Leader deletes group and returns 200",
			userIDInContext: leaderUUID,
			mockGetGroupFunc: func(ctx context.Context, userID uuid.UUID) (*domain.Group, error) {
				return &domain.Group{
					ID:       1,
					LeaderID: leaderUUID,
				}, nil
			},
			mockDeleteFunc: func(ctx context.Context, groupID int64) error {
				return nil
			},
			wantStatusCode: http.StatusOK,
			expectedBody:   `{"message":"group deleted successfully"}`,
		},
		{
			name:            "Forbidden: Non-leader tries to delete returns 403",
			userIDInContext: memberUUID,
			mockGetGroupFunc: func(ctx context.Context, userID uuid.UUID) (*domain.Group, error) {
				return &domain.Group{
					ID:       1,
					LeaderID: leaderUUID,
				}, nil
			},
			mockDeleteFunc: nil,
			wantStatusCode: http.StatusForbidden,
			expectedBody:   `{"error":"only the group leader can delete the group"}`,
		},
		{
			name:            "Error: User not in any group returns 400",
			userIDInContext: memberUUID,
			mockGetGroupFunc: func(ctx context.Context, userID uuid.UUID) (*domain.Group, error) {
				return nil, errors.New("user is not in any group")
			},
			mockDeleteFunc: nil,
			wantStatusCode: http.StatusBadRequest,
			expectedBody:   `{"error":"user is not in any group"}`,
		},
		{
			name:            "Error: No user in context returns 401",
			userIDInContext: uuid.Nil,
			mockGetGroupFunc: nil,
			mockDeleteFunc:   nil,
			wantStatusCode:   http.StatusUnauthorized,
			expectedBody:     `{"error":"unauthorized"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mService := &mockGroupService{
				getUserGroupFunc: tt.mockGetGroupFunc,
				deleteGroupFunc:  tt.mockDeleteFunc,
			}
			h := &handlers.Handler{GroupService: mService}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			if tt.userIDInContext != uuid.Nil {
				injectUserClaims(c, tt.userIDInContext)
			}

			c.Request = httptest.NewRequest("DELETE", "/api/groups", nil)

			h.DeleteGroup(c)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Handler.DeleteGroup() status = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if !strings.Contains(w.Body.String(), tt.expectedBody) {
				t.Errorf("Handler.DeleteGroup() body = %v, want to contain %v", w.Body.String(), tt.expectedBody)
			}
		})
	}
}
