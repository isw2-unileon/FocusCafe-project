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

func TestUser_Integration(t *testing.T) {
	_, db, h := setupTestApp()
	userID := uuid.New()

	// Seed Data
	db.Create(&models.User{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Email:     "john@example.com",
	})
	db.Create(&models.UserProgress{
		UserID: userID,
		Level:  1,
		Energy: 50,
		XP:     0,
	})

	t.Run("Get User Profile", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/users/me", nil)
		
		c, _ := setupSubtestContext(w, req, userID)
		h.GetUserProfile(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var profile map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &profile); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if profile["first_name"] != "John" {
			t.Errorf("expected first_name 'John', got %v", profile["first_name"])
		}
		if profile["email"] != "john@example.com" {
			t.Errorf("expected email 'john@example.com', got %v", profile["email"])
		}

		// Verify progress fields (flattened)
		if profile["level"].(float64) != 1 {
			t.Errorf("expected level 1, got %v", profile["level"])
		}
		if profile["energy"].(float64) != 50 {
			t.Errorf("expected energy 50, got %v", profile["energy"])
		}
	})

	t.Run("Update User Profile", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]string{
			"first_name": "Jane",
			"last_name":  "Smith",
		})
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/users/me", bytes.NewBuffer(reqBody))
		
		c, _ := setupSubtestContext(w, req, userID)
		h.UpdateUserProfile(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		// Verify DB changes
		var user models.User
		db.First(&user, userID)
		if user.FirstName != "Jane" || user.LastName != "Smith" {
			t.Errorf("expected updated name Jane Smith, got %s %s", user.FirstName, user.LastName)
		}

		// Verify response contains updated data
		var profile map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &profile)
		if profile["first_name"] != "Jane" {
			t.Errorf("expected response first_name 'Jane', got %v", profile["first_name"])
		}
	})

	t.Run("Update User Profile - Validation Error", func(t *testing.T) {
		reqBody, _ := json.Marshal(map[string]string{
			"first_name": "",
			"last_name":  "   ",
		})
		
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("PUT", "/api/users/me", bytes.NewBuffer(reqBody))
		
		c, _ := setupSubtestContext(w, req, userID)
		h.UpdateUserProfile(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}
