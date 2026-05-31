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

	err = db.AutoMigrate(&models.User{}, &models.CafeOrder{}, &models.UserOrder{}, &models.UserProgress{}, &models.Group{})
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

	u1, u2, u3, u4, u5, u6 := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	// Scenario Setup
	db.Create(&models.UserOrder{UserID: u1, CafeOrderID: 1, Status: "pending"})
	db.Create(&models.UserOrder{UserID: u1, CafeOrderID: 2, Status: "pending"})
	db.Create(&models.UserOrder{UserID: u1, CafeOrderID: 1, Status: "completed"})
	db.Create(&models.UserOrder{UserID: u2, CafeOrderID: 2, Status: "pending"})
	db.Create(&models.UserProgress{UserID: u3, Level: 1})
	db.Create(&models.UserProgress{UserID: u4, Level: 1})
	db.Create(&models.UserOrder{UserID: u4, CafeOrderID: 1, Status: "completed"})

	// User in DB but without group
	db.Create(&models.User{ID: u6, GroupID: nil, FirstName: "Solo", LastName: "User", Username: "solouser", Email: "solo@test.com"})
	db.Create(&models.UserOrder{UserID: u6, CafeOrderID: 1, Status: "pending"})

	// Group Scenario Setup
	groupID := int64(1)
	db.Create(&models.Group{ID: groupID, Name: "Test Group", InviteCode: "TEST", LeaderID: u5})
	db.Create(&models.User{ID: u5, GroupID: &groupID, FirstName: "Group", LastName: "User", Username: "groupuser", Email: "group@test.com"})
	db.Create(&models.UserProgress{UserID: u5, Level: 1})

	tests := []GetOrdersTC{
		{"User 1: multiple pending orders", u1, 2, true, "Espresso", false},
		{"User 2: single pending order", u2, 1, true, "Latte", false},
		{"User 3: empty list -> auto-generate", u3, 3, false, "", false},
		{"User 4: only completed -> generate new ones", u4, 3, false, "", false},
		{"Non-existent user -> create progress and generate", uuid.New(), 3, false, "", false},
		{"User 5: group member -> generate personal and group orders", u5, 6, false, "", false},

		{"User 6: exists in DB but has no group", u6, 1, true, "Espresso", false},
		{"Database Error: Cancelled context returns error", u1, 0, false, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Timeout is forced only when test explicitly searches for an error in DB
			if tt.wantErr && tt.name == "Database Error: Cancelled context returns error" {
				var cancel context.CancelFunc
				// GORM cancels queries
				ctx, cancel = context.WithTimeout(context.Background(), 1)
				defer cancel()
			}

			got, err := repo.GetUserOrders(ctx, tt.userID)
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

	t.Run("Error: Personal order - no progress", func(t *testing.T) {
		uNoprogress := uuid.New()
		oNoProgress := models.UserOrder{UserID: uNoprogress, CafeOrderID: 1, Status: "pending"}
		db.Create(&oNoProgress)

		err := repo.CompleteUserOrder(context.Background(), uNoprogress, uint(oNoProgress.ID))
		if err == nil {
			t.Errorf("expected error when no progress exists, got nil")
		}
	})
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

func TestUserOrdersRepository_CompleteGroupOrder(t *testing.T) {
	db := setupOrdersTestDB(t)
	repo := repository.NewUserOrdersRepository(db)
	seedCafeOrders(db)

	u1, u2 := uuid.New(), uuid.New()
	groupID := int64(2)
	db.Create(&models.Group{ID: groupID, Name: "Group 2", InviteCode: "TEST2", LeaderID: u1})
	db.Create(&models.User{ID: u1, GroupID: &groupID, FirstName: "L", LastName: "D", Username: "leader", Email: "l@t.com"})
	db.Create(&models.User{ID: u2, GroupID: &groupID, FirstName: "M", LastName: "B", Username: "member", Email: "m@t.com"})
	db.Create(&models.UserProgress{UserID: u1, Level: 1, XP: 0, Energy: 50})
	db.Create(&models.UserProgress{UserID: u2, Level: 1, XP: 0, Energy: 50})

	o1 := models.UserOrder{UserID: u1, CafeOrderID: 1, Status: "pending", GroupID: &groupID}
	db.Create(&o1)

	t.Run("Success: Complete group order", func(t *testing.T) {
		err := repo.CompleteUserOrder(context.Background(), u2, uint(o1.ID))
		if err != nil {
			t.Fatalf("failed to complete group order: %v", err)
		}

		var p1, p2 models.UserProgress
		db.Where("user_id = ?", u1).First(&p1)
		db.Where("user_id = ?", u2).First(&p2)

		// RewardXP for CafeOrder 1 is 5. 5 / 2 members = 2 XP each. Remainder 1 goes to completer (u2).
		if p1.XP != 2 {
			t.Errorf("expected leader to have 2 XP, got %d", p1.XP)
		}
		if p2.XP != 3 {
			t.Errorf("expected completer to have 3 XP, got %d", p2.XP)
		}
		if p2.Energy != 40 {
			t.Errorf("expected completer to have 40 energy, got %d", p2.Energy)
		}

		// Verify group orders regeneration
		var remainingGroupOrders int64
		db.Model(&models.UserOrder{}).Where("group_id = ? AND status = ?", groupID, "pending").Count(&remainingGroupOrders)
		if remainingGroupOrders != 3 {
			t.Errorf("expected 3 regenerated group orders, got %d", remainingGroupOrders)
		}
	})

	t.Run("Error: Order already completed", func(t *testing.T) {
		err := repo.CompleteUserOrder(context.Background(), u2, uint(o1.ID))
		if err == nil || err.Error() != "order already completed" {
			t.Errorf("expected 'order already completed' error, got %v", err)
		}
	})

	t.Run("Error: Group has no members", func(t *testing.T) {
		g3 := int64(3)
		db.Create(&models.Group{ID: g3, Name: "No members", InviteCode: "NONE", LeaderID: u1})
		oNoMembers := models.UserOrder{UserID: u1, CafeOrderID: 1, Status: "pending", GroupID: &g3}
		db.Create(&oNoMembers)

		err := repo.CompleteUserOrder(context.Background(), u1, uint(oNoMembers.ID))
		if err == nil || err.Error() != "group has no members" {
			t.Errorf("expected 'group has no members' error, got %v", err)
		}
	})

	t.Run("Error: Group completer has no progress", func(t *testing.T) {
		oNew := models.UserOrder{UserID: u1, CafeOrderID: 1, Status: "pending", GroupID: &groupID}
		db.Create(&oNew)
		uNoProg := uuid.New()

		err := repo.CompleteUserOrder(context.Background(), uNoProg, uint(oNew.ID))
		if err == nil {
			t.Errorf("expected error when completer has no progress")
		}
	})

	t.Run("Error: Group member has no progress", func(t *testing.T) {
		u3 := uuid.New()
		db.Create(&models.User{ID: u3, GroupID: &groupID, FirstName: "M2", LastName: "B2", Username: "member2", Email: "m2@t.com"})
		// u3 has no progress entry

		oNew2 := models.UserOrder{UserID: u1, CafeOrderID: 1, Status: "pending", GroupID: &groupID}
		db.Create(&oNew2)

		err := repo.CompleteUserOrder(context.Background(), u1, uint(oNew2.ID))
		if err == nil {
			t.Errorf("expected error when a group member has no progress")
		}
	})
}
