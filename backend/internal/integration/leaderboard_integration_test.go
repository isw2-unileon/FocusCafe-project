package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"gorm.io/gorm"
)

func TestLeaderboard_Integration(t *testing.T) {
	db, h := setupTestApp()

	userA := uuid.New()
	userB := uuid.New()

	seedUserWithXP(t, db, userA, "UserA", 100)
	seedUserWithXP(t, db, userB, "UserB", 200)

	t.Run("GET /api/leaderboard - Sorted List", func(t *testing.T) {
		runSortedListTest(t, h, userA)
	})

	t.Run("GET /api/leaderboard/me - User A Rank", func(t *testing.T) {
		runUserARankTest(t, h, userA)
	})

	t.Run("GET /api/leaderboard/me - User B Rank", func(t *testing.T) {
		runUserBRankTest(t, h, userB)
	})
}

// --- SUBTEST EXECUTORS---

func runSortedListTest(t *testing.T, h *handlers.Handler, userA uuid.UUID) {
	r := setupLeaderboardRouter(h, userA)
	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/api/leaderboard", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var result []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) < 2 {
		t.Fatalf("expected at least 2 users in leaderboard, got %d", len(result))
	}

	if result[0]["first_name"] != "UserB" {
		t.Errorf("expected first user to be UserB, got %v", result[0]["first_name"])
	}
}

func runUserARankTest(t *testing.T, h *handlers.Handler, userA uuid.UUID) {
	r := setupLeaderboardRouter(h, userA)
	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/api/leaderboard/me", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	rankVal, ok := result["rank"].(float64)
	if !ok || int(rankVal) != 2 {
		t.Errorf("expected rank 2 for UserA, got %v", result["rank"])
	}
}

func runUserBRankTest(t *testing.T, h *handlers.Handler, userB uuid.UUID) {
	r := setupLeaderboardRouter(h, userB)
	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/api/leaderboard/me", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	rankVal, ok := result["rank"].(float64)
	if !ok || int(rankVal) != 1 {
		t.Errorf("expected rank 1 for UserB, got %v", result["rank"])
	}
}

// --- INFRASTRUCTURE & SEEDERS ---

func setupLeaderboardRouter(h *handlers.Handler, userID uuid.UUID) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		protected := api.Group("/")
		protected.Use(mockAuthMiddleware(userID))
		protected.GET("/leaderboard", h.GetLeaderboard)
		protected.GET("/leaderboard/me", h.GetUserLeaderboardRank)
	}
	return r
}

func seedUserWithXP(t *testing.T, db *gorm.DB, id uuid.UUID, name string, xp int) {
	user := models.User{
		ID:        id,
		FirstName: name,
		LastName:  "Tester",
		Email:     fmt.Sprintf("%s@test.com", name),
		Username:  name,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	progress := models.UserProgress{
		UserID: id,
		Level:  1,
		Energy: 100,
		XP:     xp,
	}
	if err := db.Create(&progress).Error; err != nil {
		t.Fatalf("failed to seed progress: %v", err)
	}
}
