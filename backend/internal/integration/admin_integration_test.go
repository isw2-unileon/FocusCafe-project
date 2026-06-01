package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"gorm.io/gorm"
)

func TestAdmin_Integration(t *testing.T) {
	db, h := setupTestApp()

	db = db.Debug()

	t.Run("GetAllUsers - Access as Admin", func(t *testing.T) {
		adminID := uuid.New()
		seedAdminUserMock(db, adminID, "admin_get", "admin_get@test.com")
		seedNormalUserMock(db, uuid.New(), "regular_user1", "regular@test1.com")

		testGetAllUsersAsAdmin(t, h, adminID)
	})

	t.Run("GetAllUsers - Forbidden for normal user", func(t *testing.T) {
		userID := uuid.New()
		seedNormalUserMock(db, userID, "regular_user2", "regular@test.com")

		testGetAllUsersForbidden(t, h, userID)
	})

	t.Run("GetUserByEmail - Success", func(t *testing.T) {
		adminID := uuid.New()
		targetUserID := uuid.New()
		seedAdminUserMock(db, adminID, "admin_search", "admin_search@test.com")
		seedNormalUserMock(db, targetUserID, "search_target", "user1@test.com")

		testGetUserByEmailSuccess(t, h, adminID)
	})

	t.Run("AdminDeleteGroup - Success", func(t *testing.T) {
		adminID := uuid.New()
		userID := uuid.New()
		seedAdminUserMock(db, adminID, "admin_del", "admin_del@test.com")
		seedNormalUserMock(db, userID, "leader_del", "leader_del@test.com")

		testAdminDeleteGroupSuccess(t, db, h, adminID, userID)
	})

	t.Run("AdminCreateUser - Success", func(t *testing.T) {
		adminID := uuid.New()
		seedAdminUserMock(db, adminID, "admin_creator", "creator@test.com")

		testAdminCreateUserSuccess(t, db, h, adminID)
	})
}

func testGetAllUsersAsAdmin(t *testing.T, h *handlers.Handler, adminID uuid.UUID) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/admin/users", nil)

	// Authenticate as Admin
	setupRouterWithAuth(h, adminID).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var users []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Should see at least 2 users
	if len(users) < 2 {
		t.Errorf("expected at least 2 users, got %d", len(users))
	}
}

func testGetAllUsersForbidden(t *testing.T, h *handlers.Handler, userID uuid.UUID) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/admin/users", nil)

	// Authenticate as Normal User
	setupRouterWithAuth(h, userID).ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func testGetUserByEmailSuccess(t *testing.T, h *handlers.Handler, adminID uuid.UUID) {
	w := httptest.NewRecorder()
	url := fmt.Sprintf("/api/admin/users/search?email=%s", "user1@test.com")
	req, _ := http.NewRequest("GET", url, nil)

	setupRouterWithAuth(h, adminID).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var user map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &user); err != nil {
		t.Fatalf("failed to unmarshal user search result: %v", err)
	}
	if user["email"] != "user1@test.com" {
		t.Errorf("expected email user1@test.com, got %v", user["email"])
	}
}

func testAdminDeleteGroupSuccess(t *testing.T, db *gorm.DB, h *handlers.Handler, adminID, userID uuid.UUID) {
	groupID := int64(500)
	db.Create(&models.Group{ID: groupID, Name: "Admin Delete Me", InviteCode: "DEL", LeaderID: userID})

	w := httptest.NewRecorder()
	url := fmt.Sprintf("/api/admin/groups/%d", groupID)
	req, _ := http.NewRequest("DELETE", url, nil)

	setupRouterWithAuth(h, adminID).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify DB
	var count int64
	db.Model(&models.Group{}).Where("id = ?", groupID).Count(&count)
	if count != 0 {
		t.Errorf("expected group to be deleted")
	}
}

func testAdminCreateUserSuccess(t *testing.T, db *gorm.DB, h *handlers.Handler, adminID uuid.UUID) {
	var checkAdmin models.User
	if err := db.First(&checkAdmin, "id = ?", adminID).Error; err != nil {
		// If the admin does not exist in the context, create it
		db.Create(&models.User{
			ID:        adminID,
			Role:      "admin",
			Username:  "admin_creator_secure",
			Email:     "creator_secure@test.com",
			FirstName: "Admin",
			LastName:  "User",
		})
	}

	// 1. Setup Mock Supabase Server that also writes to a local test DB
	mockExternalID := uuid.New()
	srv := setupMockSupabaseServer(mockExternalID)
	defer srv.Close()

	// Temporarily point handler to mock server
	oldURL := h.SupabaseURL
	h.SupabaseURL = srv.URL
	defer func() { h.SupabaseURL = oldURL }()

	// Requested payload data
	expectedEmail := "admin_created@test.com"
	expectedFirstName := "Admin"
	expectedLastName := "Created"

	// 2. Prepare Request
	reqBody, _ := json.Marshal(map[string]string{
		"first_name": "Admin",
		"last_name":  "Created",
		"email":      "admin_created@test.com",
		"password":   "password123",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/admin/users", bytes.NewBuffer(reqBody))

	setupRouterWithAuth(h, adminID).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var responseMap map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &responseMap); err != nil {
		t.Fatalf("failed to unmarshal handler response: %v", err)
	}
	// 3. Verify DB
	if responseMap["email"] != expectedEmail {
		t.Errorf("expected responded email to be %s, got %v", expectedEmail, responseMap["email"])
	}
	if responseMap["id"] != mockExternalID.String() {
		t.Errorf("expected responded user ID to match Supabase mock uuid %s, got %v", mockExternalID.String(), responseMap["id"])
	}
	if responseMap["first_name"] != expectedFirstName {
		t.Errorf("expected responded first_name to be %s, got %v", expectedFirstName, responseMap["first_name"])
	}
	if responseMap["last_name"] != expectedLastName {
		t.Errorf("expected responded last_name to be %s, got %v", expectedLastName, responseMap["last_name"])
	}
}

func seedAdminUserMock(db *gorm.DB, id uuid.UUID, username, email string) {
	db.Create(&models.User{ID: id, Role: "admin", Username: username, Email: email, FirstName: "Admin", LastName: "User"})
}

func seedNormalUserMock(db *gorm.DB, id uuid.UUID, username, email string) {
	db.Create(&models.User{ID: id, Role: "user", Username: username, Email: email, FirstName: "Normal", LastName: "User"})
}

// Helper for creating the mock server
func setupMockSupabaseServer(mockSupabaseID uuid.UUID) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Simulate Supabase Authentication Route
		if strings.HasSuffix(r.URL.Path, "/signup") || strings.Contains(r.URL.Path, "/auth/v1/signup") {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"user": map[string]interface{}{
					"id":    mockSupabaseID.String(),
					"email": "admin_created_dynamic@test.com",
				},
				"access_token":  "mock-token",
				"refresh_token": "mock-refresh",
			})
			return
		}

		// Simulate Supabase Database API Responses (Success states)
		if (strings.HasSuffix(r.URL.Path, "/users") || strings.HasSuffix(r.URL.Path, "/user_progress")) && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, `[{"id": "%s", "user_id": "%s", "email": "admin_created_dynamic@test.com"}]`, mockSupabaseID.String(), mockSupabaseID.String())
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
}

// Helper to setup a router with a specific authenticated user for a test
func setupRouterWithAuth(h *handlers.Handler, userID uuid.UUID) *gin.Engine {
	r := setupTestRouter()
	api := r.Group("/api/admin")
	api.Use(mockAuthMiddleware(userID))
	api.Use(h.AdminOnly())

	api.GET("/users", h.GetAllUsers)
	api.POST("/users", h.AdminCreateUser)
	api.GET("/users/search", h.GetUserByEmail)
	api.GET("/groups", h.GetAllGroups)
	api.DELETE("/groups/:id", h.AdminDeleteGroup)

	return r
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	return r
}
