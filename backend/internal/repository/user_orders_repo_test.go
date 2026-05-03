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
