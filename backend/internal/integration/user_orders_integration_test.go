package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
)

func TestUserOrders_Integration(t *testing.T) {
	r, db, h := setupTestApp()
	userID := uuid.New()

	// 1. Setup Routes for this test
	api := r.Group("/api")
	api.Use(mockAuthMiddleware(userID))
	api.GET("/orders", h.GetUserOrders)
	api.POST("/orders/:id/complete", h.CompleteUserOrder)

	// 2. Seed Data
	db.Create(&models.CafeOrder{ID: 1, Name: "Espresso", EnergyCost: 10, RewardXP: 5, RequiredLevel: 1})
	db.Create(&models.CafeOrder{ID: 2, Name: "Latte", EnergyCost: 15, RewardXP: 10, RequiredLevel: 1})
	db.Create(&models.CafeOrder{ID: 3, Name: "Cappuccino", EnergyCost: 20, RewardXP: 15, RequiredLevel: 1})
	
	db.Create(&models.UserProgress{UserID: userID, Level: 1, Energy: 50, XP: 0})

	t.Run("GetOrders - Should auto-generate 3 orders if empty", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/orders", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		
		var orders []map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &orders); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if len(orders) != 3 {
			t.Errorf("expected 3 orders, got %d", len(orders))
		}
	})

	t.Run("CompleteOrder - Should update DB correctly", func(t *testing.T) {
		var order models.UserOrder
		db.Where("user_id = ? AND status = ? AND group_id IS NULL", userID, "pending").First(&order)

		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/orders/%d/complete", order.ID)
		req, _ := http.NewRequest("POST", url, nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var updatedOrder models.UserOrder
		db.First(&updatedOrder, order.ID)
		if updatedOrder.Status != "completed" {
			t.Errorf("expected status 'completed', got %s", updatedOrder.Status)
		}

		var progress models.UserProgress
		db.Where("user_id = ?", userID).First(&progress)
		if progress.Energy >= 50 {
			t.Errorf("expected energy to decrease from 50, got %d", progress.Energy)
		}
		if progress.XP <= 0 {
			t.Errorf("expected XP to increase from 0, got %d", progress.XP)
		}
	})

	t.Run("CompleteOrder - Insufficient Energy", func(t *testing.T) {
		db.Model(&models.UserProgress{}).Where("user_id = ?", userID).Update("energy", 0)

		var nextOrder models.UserOrder
		db.Where("user_id = ? AND status = ? AND group_id IS NULL", userID, "pending").First(&nextOrder)

		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/orders/%d/complete", nextOrder.ID)
		req, _ := http.NewRequest("POST", url, nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Not enough energy") {
			t.Errorf("expected body to contain 'Not enough energy', got %s", w.Body.String())
		}
	})
}

func TestGroupOrders_Integration(t *testing.T) {
	r, db, h := setupTestApp()
	u1, u2 := uuid.New(), uuid.New()
	groupID := int64(100)

	db.Create(&models.CafeOrder{ID: 10, Name: "Group Pizza", EnergyCost: 20, RewardXP: 10, RequiredLevel: 1})
	db.Create(&models.CafeOrder{ID: 11, Name: "Group Burger", EnergyCost: 25, RewardXP: 15, RequiredLevel: 1})
	db.Create(&models.CafeOrder{ID: 12, Name: "Group Pasta", EnergyCost: 30, RewardXP: 20, RequiredLevel: 1})

	db.Create(&models.Group{ID: groupID, Name: "Integration Group", InviteCode: "INTG", LeaderID: u1})
	db.Create(&models.User{ID: u1, GroupID: &groupID, Username: "u1", FirstName: "User", LastName: "1", Email: "u1@test.com"})
	db.Create(&models.User{ID: u2, GroupID: &groupID, Username: "u2", FirstName: "User", LastName: "2", Email: "u2@test.com"})
	db.Create(&models.UserProgress{UserID: u1, Level: 1, Energy: 50, XP: 0})
	db.Create(&models.UserProgress{UserID: u2, Level: 1, Energy: 50, XP: 0})

	t.Run("GetOrders - Should include group orders", func(t *testing.T) {
		api := r.Group("/api/group-test")
		api.Use(mockAuthMiddleware(u1))
		api.GET("/orders", h.GetUserOrders)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/group-test/orders", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		
		var orders []map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &orders)
		if len(orders) != 6 {
			t.Errorf("expected 6 orders (3 personal + 3 group), got %d", len(orders))
		}
	})

	t.Run("Complete Group Order - XP shared", func(t *testing.T) {
		var groupOrder models.UserOrder
		db.Preload("CafeOrder").Where("group_id = ?", groupID).First(&groupOrder)

		api := r.Group("/api/group-complete")
		api.Use(mockAuthMiddleware(u2)) // u2 completes it
		api.POST("/orders/:id/complete", h.CompleteUserOrder)

		w := httptest.NewRecorder()
		url := fmt.Sprintf("/api/group-complete/orders/%d/complete", groupOrder.ID)
		req, _ := http.NewRequest("POST", url, nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var p1, p2 models.UserProgress
		db.Where("user_id = ?", u1).First(&p1)
		db.Where("user_id = ?", u2).First(&p2)

		totalXP := int(groupOrder.CafeOrder.RewardXP)
		xpPerMember := totalXP / 2
		remainder := totalXP % 2

		if p1.XP != xpPerMember {
			t.Errorf("expected leader XP %d, got %d", xpPerMember, p1.XP)
		}
		if p2.XP != (xpPerMember + remainder) {
			t.Errorf("expected completer XP %d, got %d", xpPerMember+remainder, p2.XP)
		}
		expectedEnergy := 50 - int(groupOrder.CafeOrder.EnergyCost)
		if p2.Energy != expectedEnergy {
			t.Errorf("expected completer energy %d, got %d", expectedEnergy, p2.Energy)
		}
	})
}
