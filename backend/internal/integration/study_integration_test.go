package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"gorm.io/gorm"
)

func TestStudy_Integration(t *testing.T) {
	db, h := setupTestApp()

	t.Run("Update Progress - Success", func(t *testing.T) {
		userID := uuid.New()
		sessionID := seedStudyTestData(db, userID, 100)

		router := setupStudyRouter(h, userID)
		testUpdateProgressSuccess(t, router, db, userID, sessionID)
	})

	t.Run("Update Progress - Invalid Session ID", func(t *testing.T) {
		userID := uuid.New()
		_ = seedStudyTestData(db, userID, 100)

		router := setupStudyRouter(h, userID)
		testUpdateProgressInvalidSession(t, router)
	})

	t.Run("Update Progress - Missing Body Fields", func(t *testing.T) {
		userID := uuid.New()
		_ = seedStudyTestData(db, userID, 100)

		router := setupStudyRouter(h, userID)
		testUpdateProgressMissingFields(t, router)
	})
}

func setupStudyRouter(h *handlers.Handler, userID uuid.UUID) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	api.Use(mockAuthMiddleware(userID)) // Injects the target user authenticated context safely
	api.POST("/user/progress", h.UpdateProgressHandler)

	return r
}

func seedStudyTestData(db *gorm.DB, userID uuid.UUID, initialEnergy int) uint64 {
	db.Create(&models.UserProgress{
		UserID: userID,
		Level:  1,
		Energy: initialEnergy,
		XP:     0,
	})

	session := models.StudySession{
		UserID: userID,
		Status: "active",
	}
	db.Create(&session)

	return session.ID
}

func testUpdateProgressSuccess(t *testing.T, r *gin.Engine, db *gorm.DB, userID uuid.UUID, sessionID uint64) {
	// Score = 3 -> 3 * 20 = 60 energy
	reqBody, _ := json.Marshal(map[string]interface{}{
		"session_id": sessionID,
		"score":      3,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/user/progress", bytes.NewBuffer(reqBody))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal updated progress result: %v", err)
	}
	if response["energy_earned"].(float64) != 60 {
		t.Errorf("expected energy_earned 60, got %v", response["energy_earned"])
	}
	if response["new_total"].(float64) != 160 {
		t.Errorf("expected new_total 160, got %v", response["new_total"])
	}

	// Verify DB
	var progress models.UserProgress
	db.Where("user_id = ?", userID).First(&progress)
	if progress.Energy != 160 {
		t.Errorf("DB: expected energy 160, got %d", progress.Energy)
	}
}

func testUpdateProgressInvalidSession(t *testing.T, r *gin.Engine) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"session_id": 9999,
		"score":      1,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/user/progress", bytes.NewBuffer(reqBody))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func testUpdateProgressMissingFields(t *testing.T, r *gin.Engine) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"score": 1,
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/user/progress", bytes.NewBuffer(reqBody))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
