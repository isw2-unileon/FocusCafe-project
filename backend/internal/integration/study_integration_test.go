package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
)

func TestStudy_Integration(t *testing.T) {
	r, db, h := setupTestApp()
	userID := uuid.New()

	// Setup Router with real middleware and routes
	api := r.Group("/api")
	api.Use(mockAuthMiddleware(userID))
	api.POST("/user/progress", h.UpdateProgressHandler)

	// Seed Data
	db.Create(&models.UserProgress{UserID: userID, Level: 1, Energy: 100, XP: 0})
	
	// Create a dummy study session
	session := models.StudySession{
		UserID: userID,
		Status: "active",
	}
	db.Create(&session)

	t.Run("POST /api/user/progress - Success", func(t *testing.T) {
		// Score = 3 -> 3 * 20 = 60 energy
		reqBody, _ := json.Marshal(map[string]interface{}{
			"session_id": session.ID,
			"score":      3,
		})
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/progress", bytes.NewBuffer(reqBody))
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
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
	})

	t.Run("POST /api/user/progress - Invalid Session ID", func(t *testing.T) {
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
	})

	t.Run("POST /api/user/progress - Missing Body Fields", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]interface{}{
			"score": 1,
		})
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/user/progress", bytes.NewBuffer(reqBody))
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}
