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

func setupOrdersTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.CafeOrder{},
		&models.UserOrder{},
		&models.UserProgress{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func seedCafeOrders(t *testing.T, db *gorm.DB) {
	cafes := []models.CafeOrder{
		{ID: 1, Name: "Espresso", EnergyCost: 10, RewardXP: 5},
		{ID: 2, Name: "Latte", EnergyCost: 20, RewardXP: 15},
		{ID: 3, Name: "Mocha", EnergyCost: 15, RewardXP: 10},
	}
	for _, c := range cafes {
		if err := db.Create(&c).Error; err != nil {
			t.Fatalf("failed to seed cafes: %v", err)
		}
	}
}

func setupUserOrdersScenario(db *gorm.DB, u1, u2, u3, u4 uuid.UUID) {
	// User 1 and 2
	db.Create(&models.UserOrder{UserID: u1, CafeOrderID: 1, Status: "pending"})
	db.Create(&models.UserOrder{UserID: u1, CafeOrderID: 2, Status: "pending"})
	db.Create(&models.UserOrder{UserID: u1, CafeOrderID: 1, Status: "completed"})
	db.Create(&models.UserOrder{UserID: u2, CafeOrderID: 2, Status: "pending"})

	// Users for automatic generation
	db.Create(&models.UserProgress{UserID: u3, Level: 1})
	db.Create(&models.UserProgress{UserID: u4, Level: 1})
	db.Create(&models.UserOrder{UserID: u4, CafeOrderID: 1, Status: "completed"})
}

func TestUserOrdersRepository_GetUserOrders(t *testing.T) {
	db := setupOrdersTestDB(t)
	repo := repository.NewUserOrdersRepository(db)
	seedCafeOrders(t, db)

	u1, u2, u3, u4 := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	setupUserOrdersScenario(db, u1, u2, u3, u4)

	tests := []struct {
		name          string
		userID        uuid.UUID
		wantCount     int
		checkFirst    bool
		expectedFirst string
		wantErr       bool
	}{
		{
			name:          "User 1: multiple pending orders",
			userID:        u1,
			wantCount:     2,
			checkFirst:    true,
			expectedFirst: "Espresso",
			wantErr:       false,
		},
		{
			name:          "User 2: single pending order",
			userID:        u2,
			wantCount:     1,
			checkFirst:    true,
			expectedFirst: "Latte",
			wantErr:       false,
		},
		{
			name:       "User Orders empty: auto-generates orders",
			userID:     u3,
			wantCount:  3,
			checkFirst: false,
			wantErr:    false,
		},
		{
			name:       "User with only completed orders: should generate new ones",
			userID:     u4,
			wantCount:  3,
			checkFirst: false,
			wantErr:    false,
		},
		{
			name:       "Non-existent user in progress table: should create progress table and autogenerate orders",
			userID:     uuid.New(),
			wantCount:  3,
			checkFirst: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetUserOrders(context.Background(), tt.userID)
			validateResults(t, got, err, tt.wantCount, tt.expectedFirst, tt.wantErr)
		})
	}
}

func TestUserOrdersRepository_CompleteUserOrder(t *testing.T) {
	db := setupOrdersTestDB(t)
	repo := repository.NewUserOrdersRepository(db)
	seedCafeOrders(t, db)

	u1 := uuid.New()
	// Setup initial progress for u1
	db.Create(&models.UserProgress{UserID: u1, Level: 1, XP: 0, Energy: 50})

	// Create a pending order for u1
	order := models.UserOrder{UserID: u1, CafeOrderID: 1, Status: "pending"} // Espresso: Energy 10, XP 5
	db.Create(&order)

	// Create an expensive order
	expensiveOrder := models.UserOrder{UserID: u1, CafeOrderID: 2, Status: "pending"} // Latte: Energy 20, XP 15
	db.Create(&expensiveOrder)

	tests := []struct {
		name          string
		userID        uuid.UUID
		orderID       uint
		initialXP     int
		initialEnergy int
		initialLevel  int
		wantErr       bool
		expectedErr   string
		checkLevelUp  bool
	}{
		{
			name:          "Success: Complete order",
			userID:        u1,
			orderID:       uint(order.ID),
			initialXP:     0,
			initialEnergy: 50,
			initialLevel:  1,
			wantErr:       false,
		},
		{
			name:          "Error: Insufficient energy",
			userID:        u1,
			orderID:       uint(expensiveOrder.ID),
			initialXP:     0,
			initialEnergy: 5, // Less than 20
			initialLevel:  1,
			wantErr:       true,
			expectedErr:   "insufficient energy",
		},
		{
			name:          "Success: Level up",
			userID:        u1,
			orderID:       uint(order.ID),
			initialXP:     95, // 95 + 5 = 100 (Threshold for level 1 is 1 * 100)
			initialEnergy: 50,
			initialLevel:  1,
			wantErr:       false,
			checkLevelUp:  true,
		},
		{
			name:          "Error: Order not found",
			userID:        u1,
			orderID:       999,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset progress for each test case
			if tt.userID == u1 {
				db.Model(&models.UserProgress{}).Where("user_id = ?", u1).Updates(map[string]interface{}{
					"XP":     tt.initialXP,
					"energy": tt.initialEnergy,
					"Level":  tt.initialLevel,
				})
				// Reset order status if it was completed by a previous test
				if tt.orderID != 999 {
					db.Model(&models.UserOrder{}).Where("id = ?", tt.orderID).Update("status", "pending")
				}
			}

			err := repo.CompleteUserOrder(context.Background(), tt.userID, tt.orderID)

			if (err != nil) != tt.wantErr {
				t.Errorf("CompleteUserOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErr != "" && err.Error() != tt.expectedErr {
				t.Errorf("CompleteUserOrder() error = %v, want %v", err, tt.expectedErr)
			}

			if !tt.wantErr {
				// Verify DB updates
				var updatedOrder models.UserOrder
				db.First(&updatedOrder, tt.orderID)
				if updatedOrder.Status != "completed" {
					t.Errorf("Expected status completed, got %v", updatedOrder.Status)
				}

				var progress models.UserProgress
				db.Where("user_id = ?", tt.userID).First(&progress)

				var cafe models.CafeOrder
				db.First(&cafe, updatedOrder.CafeOrderID)

				expectedXP := tt.initialXP + int(cafe.RewardXP)
				expectedEnergy := tt.initialEnergy - int(cafe.EnergyCost)

				if progress.XP != expectedXP {
					t.Errorf("Expected XP %v, got %v", expectedXP, progress.XP)
				}
				if progress.Energy != expectedEnergy {
					t.Errorf("Expected energy %v, got %v", expectedEnergy, progress.Energy)
				}

				if tt.checkLevelUp && progress.Level != (tt.initialLevel+1) {
					t.Errorf("Expected level up to %v, got %v", tt.initialLevel+1, progress.Level)
				}
			}
		})
	}
}

func validateResults(t *testing.T, got []domain.UserOrder, err error, wantCount int, first string, wantErr bool) {
	if (err != nil) != wantErr {
		t.Errorf("error = %v, wantErr %v", err, wantErr)
		return
	}

	if len(got) != wantCount {
		t.Errorf("got %v orders, want %v", len(got), wantCount)
	}

	if first != "" && len(got) > 0 {
		if got[0].CafeOrder.Name != first {
			t.Errorf("expected first order %v, got %v", first, got[0].CafeOrder.Name)
		}
		if got[0].Status != "pending" {
			t.Errorf("expected status pending, got %v", got[0].Status)
		}
	}
}
