package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupOrdersTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
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

func TestUserOrdersRepository_GetUserOrders(t *testing.T) {
	db := setupOrdersTestDB(t)
	repo := repository.NewUserOrdersRepository(db)

	// Test data
	user1ID := uuid.New()
	user2ID := uuid.New()
	user3ID := uuid.New()

	cafe1 := models.CafeOrder{ID: 1, Name: "Espresso", EnergyCost: 10, RewardXP: 5}
	cafe2 := models.CafeOrder{ID: 2, Name: "Latte", EnergyCost: 20, RewardXP: 15}
	cafe3 := models.CafeOrder{ID: 3, Name: "Mocha", EnergyCost: 15, RewardXP: 10}
	db.Create(&cafe1)
	db.Create(&cafe2)
	db.Create(&cafe3)

	orders := []models.UserOrder{
		{UserID: user1ID, CafeOrderID: 1, Status: "pending"},
		{UserID: user1ID, CafeOrderID: 2, Status: "pending"},
		{UserID: user1ID, CafeOrderID: 1, Status: "completed"}, // Should be ignored
		{UserID: user2ID, CafeOrderID: 2, Status: "pending"},
	}
	for _, o := range orders {
		db.Create(&o)
	}

	db.Create(&models.UserProgress{UserID: user3ID, Level: 1})

	user4ID := uuid.New()
	db.Create(&models.UserProgress{UserID: user4ID, Level: 1})
	db.Create(&models.UserOrder{UserID: user4ID, CafeOrderID: 1, Status: "completed"})

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
			userID:        user1ID,
			wantCount:     2,
			checkFirst:    true,
			expectedFirst: "Espresso",
			wantErr:       false,
		},
		{
			name:          "User 2: single pending order",
			userID:        user2ID,
			wantCount:     1,
			checkFirst:    true,
			expectedFirst: "Latte",
			wantErr:       false,
		},
		{
			name:       "User Orders empty: auto-generates orders",
			userID:     user3ID,
			wantCount:  3,
			checkFirst: false,
			wantErr:    false,
		},
		{
			name:       "User with only completed orders: should generate new ones",
			userID:     user4ID,
			wantCount:  3,
			checkFirst: false,
			wantErr:    false,
		},
		{
			name:       "Non-existent user in progress table: should return error",
			userID:     uuid.New(),
			wantCount:  0,
			checkFirst: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetUserOrders(context.Background(), tt.userID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserOrders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantCount {
				t.Errorf("GetUserOrders() got %v orders, want %v", len(got), tt.wantCount)
			}

			if tt.checkFirst && len(got) > 0 {
				if got[0].CafeOrder == nil {
					t.Fatal("Expected CafeOrder to be preloaded, got nil")
				}
				if got[0].CafeOrder.Name != tt.expectedFirst {
					t.Errorf("Expected first order to be %v, got %v", tt.expectedFirst, got[0].CafeOrder.Name)
				}
				if got[0].Status != "pending" {
					t.Errorf("Expected status 'pending', got %v", got[0].Status)
				}
			}
		})
	}
}
