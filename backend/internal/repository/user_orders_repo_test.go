package repository_test

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite" // Cambiado al driver que no requiere CGO ni GCC
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/domain"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/repository"
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

	t.Run("Success: Group orders generated based on max level", func(t *testing.T) {
		g8 := int64(8)
		uA, uB := uuid.New(), uuid.New()
		db.Create(&models.Group{ID: g8, Name: "G8", InviteCode: "T8", LeaderID: uA})
		db.Create(&models.User{ID: uA, GroupID: &g8, Email: "ua@g8.com", Username: "uag8", FirstName: "f", LastName: "l"})
		db.Create(&models.User{ID: uB, GroupID: &g8, Email: "ub@g8.com", Username: "ubg8", FirstName: "f", LastName: "l"})
		db.Create(&models.UserProgress{UserID: uA, Level: 1, Energy: 100})
		db.Create(&models.UserProgress{UserID: uB, Level: 10, Energy: 100}) // High level member

		// High level cafe
		db.Create(&models.CafeOrder{ID: 20, Name: "Level 10 Coffee", RequiredLevel: 10, EnergyCost: 10, RewardXP: 10})

		// This should trigger addGroupOrders
		orders, err := repo.GetUserOrders(context.Background(), uA)
		if err != nil {
			t.Fatalf("failed to get orders: %v", err)
		}

		// Since uB is level 10, group orders *could* include the Level 10 Coffee.
		// We verify at least that group orders were generated (count > 0).
		hasGroupOrder := false
		for _, o := range orders {
			if o.GroupID != nil && *o.GroupID == g8 {
				hasGroupOrder = true
				break
			}
		}
		if !hasGroupOrder {
			t.Errorf("expected group orders to be generated")
		}
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// MODIFIED HERE: To prevent the pure Go driver from completing the query
			// before the nanosecond expires, we cancel the context manually and
			// immediately before execution.
			if tt.wantErr && tt.name == "Database Error: Cancelled context returns error" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel() // Cancel immediately
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

	testCompleteGroupOrderSuccess(t, db, repo, u1, u2, groupID, uint(o1.ID))
	testCompleteGroupOrderErrors(t, db, repo, u1, u2, groupID, uint(o1.ID))
	testCompleteGroupOrderXPDistribution(t, db, repo)
	testCompleteGroupOrderComplexScenarios(t, db, repo)
}

func testCompleteGroupOrderSuccess(t *testing.T, db *gorm.DB, repo *repository.UserOrdersRepository, u1, u2 uuid.UUID, groupID int64, orderID uint) {
	t.Run("Success: Complete group order", func(t *testing.T) {
		err := repo.CompleteUserOrder(context.Background(), u2, orderID)
		if err != nil {
			t.Fatalf("failed to complete group order: %v", err)
		}

		var p1, p2 models.UserProgress
		db.Where("user_id = ?", u1).First(&p1)
		db.Where("user_id = ?", u2).First(&p2)

		if p1.XP != 2 || p2.XP != 3 || p2.Energy != 40 {
			t.Errorf("XP or Energy distribution mismatch: p1.XP=%d, p2.XP=%d, p2.Energy=%d", p1.XP, p2.XP, p2.Energy)
		}

		var remainingGroupOrders int64
		db.Model(&models.UserOrder{}).Where("group_id = ? AND status = ?", groupID, "pending").Count(&remainingGroupOrders)
		if remainingGroupOrders != 3 {
			t.Errorf("expected 3 regenerated group orders, got %d", remainingGroupOrders)
		}
	})
}

func testCompleteGroupOrderErrors(t *testing.T, db *gorm.DB, repo *repository.UserOrdersRepository, u1, u2 uuid.UUID, groupID int64, orderID uint) {
	t.Run("Error: Order already completed", func(t *testing.T) {
		err := repo.CompleteUserOrder(context.Background(), u2, orderID)
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
		db.Create(&models.User{ID: u3, GroupID: &groupID, FirstName: "M2", LastName: "B2", Username: "member2", Email: "m2@test.com"})

		oNew2 := models.UserOrder{UserID: u1, CafeOrderID: 1, Status: "pending", GroupID: &groupID}
		db.Create(&oNew2)

		err := repo.CompleteUserOrder(context.Background(), u1, uint(oNew2.ID))
		if err == nil {
			t.Errorf("expected error when a group member has no progress")
		}
	})
}

func testCompleteGroupOrderXPDistribution(t *testing.T, db *gorm.DB, repo *repository.UserOrdersRepository) {
	t.Run("Success: XP distribution with 3 members", func(t *testing.T) {
		g4 := int64(4)
		u1, u2, u3 := uuid.New(), uuid.New(), uuid.New()
		db.Create(&models.Group{ID: g4, Name: "Group 4", InviteCode: "TEST4", LeaderID: u1})
		db.Create(&models.User{ID: u1, GroupID: &g4, Email: "u1@g4.com", Username: "u1g4", FirstName: "f", LastName: "l"})
		db.Create(&models.User{ID: u2, GroupID: &g4, Email: "u2@g4.com", Username: "u2g4", FirstName: "f", LastName: "l"})
		db.Create(&models.User{ID: u3, GroupID: &g4, Email: "u3@g4.com", Username: "u3g4", FirstName: "f", LastName: "l"})
		db.Create(&models.UserProgress{UserID: u1, Level: 1, XP: 0, Energy: 100})
		db.Create(&models.UserProgress{UserID: u2, Level: 1, XP: 0, Energy: 100})
		db.Create(&models.UserProgress{UserID: u3, Level: 1, XP: 0, Energy: 100})

		o := models.UserOrder{UserID: u1, CafeOrderID: 2, Status: "pending", GroupID: &g4}
		db.Create(&o)

		err := repo.CompleteUserOrder(context.Background(), u1, uint(o.ID))
		if err != nil {
			t.Fatalf("failed to complete group order: %v", err)
		}

		var p1, p2, p3 models.UserProgress
		db.Where("user_id = ?", u1).First(&p1)
		db.Where("user_id = ?", u2).First(&p2)
		db.Where("user_id = ?", u3).First(&p3)

		if p1.XP != 5 || p2.XP != 5 || p3.XP != 5 {
			t.Errorf("expected 5 XP each, got %d, %d, %d", p1.XP, p2.XP, p3.XP)
		}
	})
}

func testCompleteGroupOrderComplexScenarios(t *testing.T, db *gorm.DB, repo *repository.UserOrdersRepository) {
	t.Run("Error: Group completer insufficient energy", func(t *testing.T) {
		g6 := int64(6)
		uInsuff := uuid.New()
		db.Create(&models.Group{ID: g6, Name: "G6", InviteCode: "T6", LeaderID: uInsuff})
		db.Create(&models.User{ID: uInsuff, GroupID: &g6, Email: "uinsuff@g6.com", Username: "uinsuff", FirstName: "f", LastName: "l"})
		db.Create(&models.UserProgress{UserID: uInsuff, Energy: 5, Level: 1})

		o := models.UserOrder{UserID: uInsuff, CafeOrderID: 2, Status: "pending", GroupID: &g6}
		db.Create(&o)

		err := repo.CompleteUserOrder(context.Background(), uInsuff, uint(o.ID))
		if err == nil || err.Error() != "insufficient energy" {
			t.Errorf("expected 'insufficient energy' error, got %v", err)
		}
	})

	t.Run("Success: Group member levels up from shared XP", func(t *testing.T) {
		g7 := int64(7)
		uLvl1, uLvl2 := uuid.New(), uuid.New()
		db.Create(&models.Group{ID: g7, Name: "G7", InviteCode: "T7", LeaderID: uLvl1})
		db.Create(&models.User{ID: uLvl1, GroupID: &g7, Email: "ulvl1@g7.com", Username: "ulvl1", FirstName: "f", LastName: "l"})
		db.Create(&models.User{ID: uLvl2, GroupID: &g7, Email: "ulvl2@g7.com", Username: "ulvl2", FirstName: "f", LastName: "l"})
		db.Create(&models.UserProgress{UserID: uLvl1, XP: 0, Level: 1, Energy: 100})
		db.Create(&models.UserProgress{UserID: uLvl2, XP: 98, Level: 1, Energy: 100})

		o := models.UserOrder{UserID: uLvl1, CafeOrderID: 1, Status: "pending", GroupID: &g7}
		db.Create(&o)

		err := repo.CompleteUserOrder(context.Background(), uLvl1, uint(o.ID))
		if err != nil {
			t.Fatalf("failed to complete group order: %v", err)
		}

		var p2 models.UserProgress
		db.Where("user_id = ?", uLvl2).First(&p2)
		if p2.Level != 2 {
			t.Errorf("expected member to level up to 2, got %d", p2.Level)
		}
	})
}
