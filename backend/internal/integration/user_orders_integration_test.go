package integration

import (
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

// --- PERSONAL USER ORDERS TEST SUITE ---

func TestUserOrders_Integration(t *testing.T) {
	db, h := setupTestApp()
	seedCafeCatalog(db)

	t.Run("GetOrders - Auto Generates 3 Orders If Empty", func(t *testing.T) {
		userID := uuid.New()
		seedUserProgressMock(db, userID, 50, 0)
		router := setupOrdersRouter(h, userID)

		testGetOrdersAutoGeneration(t, router)
	})

	t.Run("CompleteOrder - Updates State Properties Mutably", func(t *testing.T) {
		userID := uuid.New()
		seedUserProgressMock(db, userID, 50, 0)
		router := setupOrdersRouter(h, userID)

		testCompleteOrderSuccess(t, router, db, userID)
	})

	t.Run("CompleteOrder - Fails On Insufficient Energy Pool", func(t *testing.T) {
		userID := uuid.New()
		seedUserProgressMock(db, userID, 0, 0)
		router := setupOrdersRouter(h, userID)

		testCompleteOrderInsufficientEnergy(t, router, db, userID)
	})
}

// --- GROUP ORDERS TEST SUITE ---

func TestGroupOrders_Integration(t *testing.T) {
	db, h := setupTestApp()
	seedCafeCatalog(db)
	seedGroupCafeCatalog(db)

	t.Run("GetOrders - Merges Personal And Shared Group Records", func(t *testing.T) {
		u1, _, _ := seedGroupIntegrationContext(db)
		router := setupOrdersRouter(h, u1)

		testGroupGetOrdersAggregation(t, router)
	})

	t.Run("Complete Group Order - Distributes XP Formula Equivalently", func(t *testing.T) {
		u1, u2, groupID := seedGroupIntegrationContext(db)
		router := setupOrdersRouter(h, u2)

		testGroupCompleteOrderXpSharing(t, router, db, groupID, u1, u2)
	})
}

// --- ROUTER & CONFIGURATION GENERATORS ---

func setupOrdersRouter(h *handlers.Handler, userID uuid.UUID) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	api.Use(mockAuthMiddleware(userID))
	api.GET("/orders", h.GetUserOrders)
	api.POST("/orders/:id/complete", h.CompleteUserOrder)

	return r
}

// --- DATABASE SEED HELPERS ---

func seedCafeCatalog(db *gorm.DB) {
	db.Create(&models.CafeOrder{ID: 1, Name: "Espresso", EnergyCost: 10, RewardXP: 5, RequiredLevel: 1})
	db.Create(&models.CafeOrder{ID: 2, Name: "Latte", EnergyCost: 15, RewardXP: 10, RequiredLevel: 1})
	db.Create(&models.CafeOrder{ID: 3, Name: "Cappuccino", EnergyCost: 20, RewardXP: 15, RequiredLevel: 1})
}

func seedGroupCafeCatalog(db *gorm.DB) {
	db.Create(&models.CafeOrder{ID: 10, Name: "Group Pizza", EnergyCost: 20, RewardXP: 10, RequiredLevel: 1})
	db.Create(&models.CafeOrder{ID: 11, Name: "Group Burger", EnergyCost: 25, RewardXP: 15, RequiredLevel: 1})
	db.Create(&models.CafeOrder{ID: 12, Name: "Group Pasta", EnergyCost: 30, RewardXP: 20, RequiredLevel: 1})
}

func seedUserProgressMock(db *gorm.DB, id uuid.UUID, energy int, xp int) {
	db.Create(&models.UserProgress{UserID: id, Level: 1, Energy: energy, XP: xp})
}

func seedGroupIntegrationContext(db *gorm.DB) (uuid.UUID, uuid.UUID, int64) {
	u1, u2 := uuid.New(), uuid.New()
	groupID := int64(uuid.New().ID())

	group := models.Group{
		ID:         groupID,
		Name:       "Integration Group",
		InviteCode: "INT" + u1.String()[:4],
		LeaderID:   u1,
	}
	db.Create(&group)

	user1 := models.User{
		ID:        u1,
		GroupID:   &groupID,
		Username:  "u_" + u1.String()[:6],
		Email:     u1.String()[:6] + "@test.com",
		FirstName: "Leader",
		LastName:  "Test",
	}
	db.Create(&user1)

	user2 := models.User{
		ID:        u2,
		GroupID:   &groupID,
		Username:  "u_" + u2.String()[:6],
		Email:     u2.String()[:6] + "@test.com",
		FirstName: "Member",
		LastName:  "Test",
	}
	db.Create(&user2)

	seedUserProgressMock(db, u1, 50, 0)
	seedUserProgressMock(db, u2, 50, 0)

	return u1, u2, groupID
}

// --- PERSONAL SUBTEST LOGIC CALLS ---

func testGetOrdersAutoGeneration(t *testing.T, r *gin.Engine) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/orders", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var orders []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &orders); err != nil {
		t.Fatalf("failed to unmarshal response payload: %v", err)
	}
	if len(orders) != 3 {
		t.Errorf("expected 3 auto-generated orders, got %d", len(orders))
	}
}

func testCompleteOrderSuccess(t *testing.T, r *gin.Engine, db *gorm.DB, userID uuid.UUID) {
	wGet := httptest.NewRecorder()
	reqGet, _ := http.NewRequest("GET", "/api/orders", nil)
	r.ServeHTTP(wGet, reqGet)

	var order models.UserOrder
	if err := db.Where("user_id = ? AND status = ? AND group_id IS NULL", userID, "pending").First(&order).Error; err != nil {
		t.Fatalf("prerequisite user order state row assignment missing: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/orders/%d/complete", order.ID), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var updatedOrder models.UserOrder
	db.First(&updatedOrder, order.ID)
	if updatedOrder.Status != "completed" {
		t.Errorf("expected status flag target to point to 'completed', got %s", updatedOrder.Status)
	}

	var progress models.UserProgress
	db.Where("user_id = ?", userID).First(&progress)
	if progress.Energy >= 50 {
		t.Errorf("expected user energy pool limits to drop below initial value (50), got %d", progress.Energy)
	}
	if progress.XP <= 0 {
		t.Errorf("expected system experience metrics to grow above initial value (0), got %d", progress.XP)
	}
}

func testCompleteOrderInsufficientEnergy(t *testing.T, r *gin.Engine, db *gorm.DB, userID uuid.UUID) {
	wGet := httptest.NewRecorder()
	reqGet, _ := http.NewRequest("GET", "/api/orders", nil)
	r.ServeHTTP(wGet, reqGet)

	var nextOrder models.UserOrder
	if err := db.Where("user_id = ? AND status = ? AND group_id IS NULL", userID, "pending").First(&nextOrder).Error; err != nil {
		t.Fatalf("failed to parse testing target references: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/orders/%d/complete", nextOrder.ID), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected validation exception status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Not enough energy") {
		t.Errorf("expected body runtime error message string matches missing capacity, got %s", w.Body.String())
	}
}

// --- GROUP SUBTEST LOGIC CALLS ---

func testGroupGetOrdersAggregation(t *testing.T, r *gin.Engine) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/orders", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var orders []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &orders); err != nil {
		t.Fatalf("failed to unmarshal JSON order elements: %v", err)
	}
	if len(orders) != 6 {
		t.Errorf("expected aggregated return schema of 6 total indices (3 Personal + 3 Group items), got %d", len(orders))
	}
}

func testGroupCompleteOrderXpSharing(t *testing.T, r *gin.Engine, db *gorm.DB, groupID int64, u1, u2 uuid.UUID) {
	wGet := httptest.NewRecorder()
	reqGet, _ := http.NewRequest("GET", "/api/orders", nil)
	r.ServeHTTP(wGet, reqGet)

	var groupOrder models.UserOrder
	if err := db.Preload("CafeOrder").Where("group_id = ?", groupID).First(&groupOrder).Error; err != nil {
		t.Fatalf("failed to query instantiated group order bindings: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/orders/%d/complete", groupOrder.ID), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var p1, p2 models.UserProgress
	db.Where("user_id = ?", u1).First(&p1)
	db.Where("user_id = ?", u2).First(&p2)

	totalXP := int(groupOrder.CafeOrder.RewardXP)
	xpPerMember := totalXP / 2
	remainder := totalXP % 2

	if p1.XP != xpPerMember {
		t.Errorf("expected team leader shared partition size %d, got %d", xpPerMember, p1.XP)
	}
	if p2.XP != (xpPerMember + remainder) {
		t.Errorf("expected operational completer partition size %d, got %d", xpPerMember+remainder, p2.XP)
	}

	expectedEnergy := 50 - int(groupOrder.CafeOrder.EnergyCost)
	if p2.Energy != expectedEnergy {
		t.Errorf("expected energy balance deduction to map to %d, got %d", expectedEnergy, p2.Energy)
	}
}
