package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// --- Helper Structs for Clean Testing ---
type GetOrdersTC struct {
	name          string
	userID        uuid.UUID
	wantCount     int
	checkFirst    bool
	expectedFirst string
	wantErr       bool
}

type CompleteOrderTC struct {
	name          string
	userID        uuid.UUID
	orderID       uint
	initialXP     int
	initialEnergy int
	initialLevel  int
	wantErr       bool
	expectedErr   string
	checkLevelUp  bool
}

// --- Database Setup (In-Memory) ---
func setupOrdersTestDB(t *testing.T) *gorm.DB {
	// Initializes a private in-memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.CafeOrder{}, &models.UserOrder{}, &models.UserProgress{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func seedCafeOrders(db *gorm.DB) {
	cafes := []models.CafeOrder{
		{ID: 1, Name: "Espresso", EnergyCost: 10, RewardXP: 5},
		{ID: 2, Name: "Latte", EnergyCost: 20, RewardXP: 15},
		{ID: 3, Name: "Mocha", EnergyCost: 15, RewardXP: 10},
	}
	for _, c := range cafes {
		db.Create(&c)
	}
}

// --- TEST 1: GET USER ORDERS ---
func TestUserOrdersRepository_GetUserOrders(t *testing.T) {
	db := setupOrdersTestDB(t)
	repo := repository.NewUserOrdersRepository(db)
	seedCafeOrders(db)

	u1, u2, u3, u4 := uuid.New(), uuid.New(), uuid.New(), uuid.New()

	// Scenario Setup
	db.Create(&models.UserOrder{UserID: u1, CafeOrderID: 1, Status: "pending"})
	db.Create(&models.UserOrder{UserID: u1, CafeOrderID: 2, Status: "pending"})
	db.Create(&models.UserOrder{UserID: u1, CafeOrderID: 1, Status: "completed"})
	db.Create(&models.UserOrder{UserID: u2, CafeOrderID: 2, Status: "pending"})
	db.Create(&models.UserProgress{UserID: u3, Level: 1})
	db.Create(&models.UserProgress{UserID: u4, Level: 1})
	db.Create(&models.UserOrder{UserID: u4, CafeOrderID: 1, Status: "completed"})

	tests := []GetOrdersTC{
		{"User 1: multiple pending orders", u1, 2, true, "Espresso", false},
		{"User 2: single pending order", u2, 1, true, "Latte", false},
		{"User 3: empty list -> auto-generate", u3, 3, false, "", false},
		{"User 4: only completed -> generate new ones", u4, 3, false, "", false},
		{"Non-existent user -> create progress and generate", uuid.New(), 3, false, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetUserOrders(context.Background(), tt.userID)
			validateGetResults(t, got, err, tt)
		})
	}
}

func validateGetResults(t *testing.T, got []domain.UserOrder, err error, tt GetOrdersTC) {
	if (err != nil) != tt.wantErr {
		t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
		return
	}
	if len(got) != tt.wantCount {
		t.Errorf("got %d orders, want %d", len(got), tt.wantCount)
	}
	if tt.checkFirst && len(got) > 0 {
		if got[0].CafeOrder.Name != tt.expectedFirst {
			t.Errorf("expected first order %s, got %s", tt.expectedFirst, got[0].CafeOrder.Name)
		}
	}
}

// --- TEST 2: COMPLETE ORDER ---
func TestUserOrdersRepository_CompleteUserOrder(t *testing.T) {
	db := setupOrdersTestDB(t)
	repo := repository.NewUserOrdersRepository(db)
	seedCafeOrders(db)
	u1 := uuid.New()

	db.Create(&models.UserProgress{UserID: u1, Level: 1, XP: 0, Energy: 50})
	o1 := models.UserOrder{UserID: u1, CafeOrderID: 1, Status: "pending"}
	o2 := models.UserOrder{UserID: u1, CafeOrderID: 2, Status: "pending"}
	db.Create(&o1)
	db.Create(&o2)

	tests := []CompleteOrderTC{
		{"Success: Standard completion", u1, uint(o1.ID), 0, 50, 1, false, "", false},
		{"Error: Insufficient energy", u1, uint(o2.ID), 0, 5, 1, true, "insufficient energy", false},
		{"Success: Trigger level up", u1, uint(o1.ID), 95, 50, 1, false, "", true},
		{"Error: Invalid order ID", u1, 999, 0, 50, 1, true, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetUserStats(db, tt)
			err := repo.CompleteUserOrder(context.Background(), tt.userID, tt.orderID)
			validateCompletion(t, db, tt, err)
		})
	}
}

func resetUserStats(db *gorm.DB, tt CompleteOrderTC) {
	db.Model(&models.UserProgress{}).Where("user_id = ?", tt.userID).Updates(map[string]interface{}{
		"xp": tt.initialXP, "energy": tt.initialEnergy, "level": tt.initialLevel,
	})
	if tt.orderID != 999 {
		db.Model(&models.UserOrder{}).Where("id = ?", tt.orderID).Update("status", "pending")
	}
}

func validateCompletion(t *testing.T, db *gorm.DB, tt CompleteOrderTC, err error) {
	if (err != nil) != tt.wantErr {
		t.Fatalf("CompleteUserOrder() error = %v, wantErr %v", err, tt.wantErr)
	}
	if tt.wantErr {
		if tt.expectedErr != "" && err.Error() != tt.expectedErr {
			t.Errorf("expected error %s, got %v", tt.expectedErr, err)
		}
		return
	}

	var p models.UserProgress
	var o models.UserOrder
	db.First(&o, tt.orderID)
	db.Where("user_id = ?", tt.userID).First(&p)

	if o.Status != "completed" {
		t.Error("database status should be 'completed'")
	}
	if tt.checkLevelUp && p.Level <= tt.initialLevel {
		t.Error("level up check failed")
	}
}
