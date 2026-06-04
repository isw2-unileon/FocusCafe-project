package repository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/repository"

	"github.com/glebarez/sqlite"

	"gorm.io/gorm"
)

func setupGroupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.UserProgress{},
		&models.Group{},
		&models.UserOrder{},
		&models.CafeOrder{},
	)
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func TestGroupRepository_CreateGroup(t *testing.T) {
	db := setupGroupTestDB(t)
	repo := repository.NewGroupRepository(db)
	ctx := context.Background()

	leaderID := uuid.New()
	group := &models.Group{
		Name:       "Test Group",
		InviteCode: "INVITE123",
		LeaderID:   leaderID,
	}

	err := repo.CreateGroup(ctx, group)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if group.ID == 0 {
		t.Errorf("Expected non-zero group ID")
	}

	var saved models.Group
	err = db.First(&saved, group.ID).Error
	if err != nil {
		t.Errorf("Expected to find group, got error %v", err)
	}
	if saved.Name != "Test Group" {
		t.Errorf("Expected name 'Test Group', got %s", saved.Name)
	}
}

func TestGroupRepository_GetGroupByInviteCode(t *testing.T) {
	db := setupGroupTestDB(t)
	repo := repository.NewGroupRepository(db)
	ctx := context.Background()

	inviteCode := "CODE123"
	db.Create(&models.Group{
		Name:       "Find Me",
		InviteCode: inviteCode,
		LeaderID:   uuid.New(),
	})

	t.Run("Success", func(t *testing.T) {
		group, err := repo.GetGroupByInviteCode(ctx, inviteCode)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if group == nil {
			t.Fatalf("Expected group to be found")
		}
		if group.InviteCode != inviteCode {
			t.Errorf("Expected invite code %s, got %s", inviteCode, group.InviteCode)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		group, err := repo.GetGroupByInviteCode(ctx, "NONEXISTENT")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if group != nil {
			t.Errorf("Expected nil group, got %v", group)
		}
	})
}

func TestGroupRepository_GetGroupByID(t *testing.T) {
	db := setupGroupTestDB(t)
	repo := repository.NewGroupRepository(db)
	ctx := context.Background()

	group := &models.Group{
		Name:       "ID Group",
		InviteCode: "IDCODE",
		LeaderID:   uuid.New(),
	}
	db.Create(group)

	t.Run("Success", func(t *testing.T) {
		found, err := repo.GetGroupByID(ctx, group.ID)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if found == nil {
			t.Fatalf("Expected group to be found")
		}
		if found.ID != group.ID {
			t.Errorf("Expected ID %d, got %d", group.ID, found.ID)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		found, err := repo.GetGroupByID(ctx, 9999)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if found != nil {
			t.Errorf("Expected nil group, got %v", found)
		}
	})
}

func TestGroupRepository_AddUserToGroup(t *testing.T) {
	db := setupGroupTestDB(t)
	repo := repository.NewGroupRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	db.Create(&models.User{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Email:     "john@example.com",
	})

	groupID := int64(1)
	db.Create(&models.Group{ID: groupID, Name: "Target Group", InviteCode: "TGT", LeaderID: uuid.New()})

	err := repo.AddUserToGroup(ctx, userID, groupID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	var updatedUser models.User
	db.First(&updatedUser, "id = ?", userID)
	if updatedUser.GroupID == nil {
		t.Fatalf("Expected group_id to be set")
	}
	if *updatedUser.GroupID != groupID {
		t.Errorf("Expected group_id %d, got %d", groupID, *updatedUser.GroupID)
	}
}

func TestGroupRepository_GetUserGroupID(t *testing.T) {
	db := setupGroupTestDB(t)
	repo := repository.NewGroupRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	groupID := int64(10)
	db.Create(&models.User{
		ID:        userID,
		FirstName: "Member",
		LastName:  "User",
		Username:  "member",
		Email:     "member@example.com",
		GroupID:   &groupID,
	})

	resID, err := repo.GetUserGroupID(ctx, userID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if resID == nil {
		t.Fatalf("Expected result ID to be non-nil")
	}
	if *resID != groupID {
		t.Errorf("Expected group ID %d, got %d", groupID, *resID)
	}
}

func TestGroupRepository_IsUserInGroup(t *testing.T) {
	db := setupGroupTestDB(t)
	repo := repository.NewGroupRepository(db)
	ctx := context.Background()

	u1, u2 := uuid.New(), uuid.New()
	gID := int64(1)

	db.Create(&models.User{ID: u1, Email: "u1@e.com", FirstName: "a", LastName: "b", Username: "u1", GroupID: &gID})
	db.Create(&models.User{ID: u2, Email: "u2@e.com", FirstName: "a", LastName: "b", Username: "u2", GroupID: nil})

	inGroup, err := repo.IsUserInGroup(ctx, u1)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !inGroup {
		t.Errorf("Expected user to be in group")
	}

	inGroup, err = repo.IsUserInGroup(ctx, u2)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if inGroup {
		t.Errorf("Expected user to NOT be in group")
	}
}

func TestGroupRepository_GetGroupMembers(t *testing.T) {
	db := setupGroupTestDB(t)
	repo := repository.NewGroupRepository(db)
	ctx := context.Background()

	gID := int64(5)
	db.Create(&models.User{ID: uuid.New(), Email: "m1@e.com", FirstName: "a", LastName: "b", Username: "m1", GroupID: &gID})
	db.Create(&models.User{ID: uuid.New(), Email: "m2@e.com", FirstName: "a", LastName: "b", Username: "m2", GroupID: &gID})
	db.Create(&models.User{ID: uuid.New(), Email: "m3@e.com", FirstName: "a", LastName: "b", Username: "m3", GroupID: nil})

	members, err := repo.GetGroupMembers(ctx, gID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(members))
	}
}

func TestGroupRepository_GetAllGroups(t *testing.T) {
	db := setupGroupTestDB(t)
	repo := repository.NewGroupRepository(db)
	ctx := context.Background()

	db.Create(&models.Group{ID: 1, Name: "G1", InviteCode: "C1", LeaderID: uuid.New()})
	db.Create(&models.Group{ID: 2, Name: "G2", InviteCode: "C2", LeaderID: uuid.New()})

	groups, err := repo.GetAllGroups(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}
}

func TestGroupRepository_DeleteGroup(t *testing.T) {
	db := setupGroupTestDB(t)
	repo := repository.NewGroupRepository(db)
	ctx := context.Background()

	gID := int64(1)
	db.Create(&models.Group{ID: gID, Name: "To Delete", InviteCode: "DEL", LeaderID: uuid.New()})

	err := repo.DeleteGroup(ctx, gID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	var count int64
	db.Model(&models.Group{}).Where("id = ?", gID).Count(&count)
	if count != 0 {
		t.Errorf("Expected group to be deleted")
	}
}

func TestGroupRepository_RemoveAllUsersFromGroup(t *testing.T) {
	db := setupGroupTestDB(t)
	repo := repository.NewGroupRepository(db)
	ctx := context.Background()

	gID := int64(1)
	u1 := uuid.New()
	u2 := uuid.New()
	db.Create(&models.User{ID: u1, Email: "u1@e.com", FirstName: "a", LastName: "b", Username: "u1", GroupID: &gID})
	db.Create(&models.User{ID: u2, Email: "u2@e.com", FirstName: "a", LastName: "b", Username: "u2", GroupID: &gID})

	err := repo.RemoveAllUsersFromGroup(ctx, gID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	var users []models.User
	db.Where("group_id = ?", gID).Find(&users)
	if len(users) != 0 {
		t.Errorf("Expected 0 members, got %d", len(users))
	}
}

func TestGroupRepository_IsGroupLeader(t *testing.T) {
	db := setupGroupTestDB(t)
	repo := repository.NewGroupRepository(db)
	ctx := context.Background()

	leaderID := uuid.New()
	otherID := uuid.New()
	gID := int64(1)

	db.Create(&models.Group{ID: gID, Name: "Leader Group", InviteCode: "LG", LeaderID: leaderID})

	isLeader, err := repo.IsGroupLeader(ctx, leaderID, gID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !isLeader {
		t.Errorf("Expected user to be leader")
	}

	isLeader, err = repo.IsGroupLeader(ctx, otherID, gID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if isLeader {
		t.Errorf("Expected user to NOT be leader")
	}
}

func TestGroupRepository_RemoveUserFromGroup(t *testing.T) {
	db := setupGroupTestDB(t)
	repo := repository.NewGroupRepository(db)
	ctx := context.Background()

	uID := uuid.New()
	gID := int64(1)
	db.Create(&models.User{ID: uID, Email: "u@e.com", FirstName: "a", LastName: "b", Username: "u", GroupID: &gID})

	err := repo.RemoveUserFromGroup(ctx, uID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	var user models.User
	db.First(&user, "id = ?", uID)
	if user.GroupID != nil {
		t.Errorf("Expected nil group_id")
	}
}

func TestGroupRepository_DeleteGroupOrders(t *testing.T) {
	db := setupGroupTestDB(t)
	repo := repository.NewGroupRepository(db)
	ctx := context.Background()

	gID := int64(1)
	uID := uuid.New()
	
	// Create CafeOrder first due to FK
	db.Create(&models.CafeOrder{ID: 1, Name: "C", EnergyCost: 1, RewardXP: 1})

	db.Create(&models.UserOrder{ID: 1, UserID: uID, GroupID: &gID, CafeOrderID: 1, Status: "pending"})
	db.Create(&models.UserOrder{ID: 2, UserID: uID, GroupID: nil, CafeOrderID: 1, Status: "pending"})

	err := repo.DeleteGroupOrders(ctx, gID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	var count int64
	db.Model(&models.UserOrder{}).Where("group_id = ?", gID).Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 group orders, got %d", count)
	}

	db.Model(&models.UserOrder{}).Where("group_id IS NULL").Count(&count)
	if count != 1 {
		t.Errorf("Expected 1 personal order to remain, got %d", count)
	}
}
