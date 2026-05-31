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
)

func TestAdmin_Integration(t *testing.T) {
	_, db, h := setupTestApp()
	adminID := uuid.New()
	userID := uuid.New()

	// Seed Users
	db.Create(&models.User{ID: adminID, Role: "admin", Username: "admin", Email: "admin@test.com", FirstName: "Admin", LastName: "User"})
	db.Create(&models.User{ID: userID, Role: "user", Username: "user1", Email: "user1@test.com", FirstName: "User", LastName: "One"})

	t.Run("GetAllUsers - Access as Admin", func(t *testing.T) {
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
	})

	t.Run("GetAllUsers - Forbidden for normal user", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/admin/users", nil)
		
		// Authenticate as Normal User
		setupRouterWithAuth(h, userID).ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("GetUserByEmail - Success", func(t *testing.T) {
		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/admin/users/search?email=%s", "user1@test.com")
		req, _ := http.NewRequest("GET", url, nil)
		
		setupRouterWithAuth(h, adminID).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var user map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &user)
		if user["email"] != "user1@test.com" {
			t.Errorf("expected email user1@test.com, got %v", user["email"])
		}
	})

	t.Run("AdminDeleteGroup - Success", func(t *testing.T) {
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
	})

	t.Run("AdminCreateUser - Success", func(t *testing.T) {
		// 1. Setup Mock Supabase Server that also writes to our local test DB
		mockSupabaseID := uuid.New()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			
			if strings.HasSuffix(r.URL.Path, "/signup") {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `{"user": {"id": "%s"}}`, mockSupabaseID.String())
				return
			}
			
			if strings.HasSuffix(r.URL.Path, "/users") && r.Method == http.MethodPost {
				// Parse the body to write to local DB
				var body map[string]string
				json.NewDecoder(r.Body).Decode(&body)
				
				db.Create(&models.User{
					ID:        mockSupabaseID,
					FirstName: body["first_name"],
					LastName:  body["last_name"],
					Username:  body["username"],
					Email:     body["email"],
					Role:      body["role"],
				})
				w.WriteHeader(http.StatusCreated)
				fmt.Fprint(w, `[]`)
				return
			}

			if strings.HasSuffix(r.URL.Path, "/user_progress") && r.Method == http.MethodPost {
				var body map[string]interface{}
				json.NewDecoder(r.Body).Decode(&body)
				db.Create(&models.UserProgress{
					UserID: mockSupabaseID,
					Energy: int(body["energy"].(float64)),
					Level:  int(body["level"].(float64)),
					XP:     int(body["xp"].(float64)),
				})
				w.WriteHeader(http.StatusCreated)
				fmt.Fprint(w, `[]`)
				return
			}

			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		// Temporarily point handler to mock server
		oldURL := h.SupabaseURL
		h.SupabaseURL = srv.URL
		defer func() { h.SupabaseURL = oldURL }()

		// 2. Prepare Request
		reqBody, _ := json.Marshal(map[string]string{
			"first_name": "Admin",
			"last_name":  "Created",
			"email":      "admin_created@test.com",
			"password":   "password123",
			"role":       "user",
		})
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/admin/users", bytes.NewBuffer(reqBody))
		
		setupRouterWithAuth(h, adminID).ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d. Body: %s", w.Code, w.Body.String())
		}

		// 3. Verify DB
		var createdUser models.User
		err := db.Where("email = ?", "admin_created@test.com").First(&createdUser).Error
		if err != nil {
			t.Errorf("expected user to be created in DB, got error: %v", err)
		}
		if createdUser.ID != mockSupabaseID {
			t.Errorf("expected ID %s, got %s", mockSupabaseID, createdUser.ID)
		}
	})
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
