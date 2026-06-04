package repository_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/repository"

	"github.com/glebarez/sqlite"

	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Create a unique in-memory SQLite database per test to avoid cross-test contamination
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// AutoMigrate to create the necessary tables
	err = db.AutoMigrate(
		&models.User{},
		&models.UserProgress{},
		&models.Group{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func TestUserRepository_CreateUser(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		ID:        uuid.New(),
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Email:     "john@example.com",
		Role:      "user",
	}

	err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	var savedUser models.User
	err = db.First(&savedUser, "id = ?", user.ID).Error
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if savedUser.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, savedUser.Email)
	}
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	email := "test@example.com"
	user := &models.User{
		ID:        uuid.New(),
		FirstName: "Test",
		LastName:  "User",
		Username:  "testuser",
		Email:     email,
		Role:      "user",
	}
	db.Create(user)

	t.Run("Success", func(t *testing.T) {
		foundUser, err := repo.GetUserByEmail(ctx, email)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if foundUser == nil {
			t.Fatalf("Expected foundUser to be non-nil")
		}
		if foundUser.ID != user.ID {
			t.Errorf("Expected ID %v, got %v", user.ID, foundUser.ID)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		foundUser, err := repo.GetUserByEmail(ctx, "nonexistent@example.com")
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
		if foundUser != nil {
			t.Errorf("Expected foundUser to be nil, got %v", foundUser)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Errorf("Expected gorm.ErrRecordNotFound, got %v", err)
		}
	})
}

func TestUserRepository_GetUserProfile(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	groupID := int64(1)

	group := &models.Group{
		ID:         groupID,
		Name:       "Test Group",
		InviteCode: "ABCDEF",
		LeaderID:   userID,
	}
	db.Create(group)

	user := &models.User{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Username:  "johndoe",
		Email:     "john@example.com",
		Role:      "user",
		GroupID:   &groupID,
	}
	db.Create(user)

	progress := &models.UserProgress{
		UserID: userID,
		Energy: 100,
		XP:     500,
		Level:  5,
	}
	db.Create(progress)

	testUserProfileSuccess(ctx, t, repo, userID)
	testUserProfileNoProgressOrGroup(ctx, t, db, repo)
	testUserProfileNotFound(ctx, t, repo)
}

func testUserProfileSuccess(ctx context.Context, t *testing.T, repo *repository.UserRepository, userID uuid.UUID) {
	t.Run("Success", func(t *testing.T) {
		profile, err := repo.GetUserProfile(ctx, userID)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if profile == nil {
			t.Fatalf("Expected profile to be non-nil")
		}
		if profile.FirstName != "John" {
			t.Errorf("Expected FirstName 'John', got %s", profile.FirstName)
		}
		if profile.Energy != 100 {
			t.Errorf("Expected Energy 100, got %d", profile.Energy)
		}
		if profile.XP != 500 {
			t.Errorf("Expected XP 500, got %d", profile.XP)
		}
		if profile.Level != 5 {
			t.Errorf("Expected Level 5, got %d", profile.Level)
		}
		if profile.Group == nil {
			t.Fatalf("Expected Group to be non-nil")
		}
		if profile.Group.Name != "Test Group" {
			t.Errorf("Expected Group Name 'Test Group', got %s", profile.Group.Name)
		}
	})
}

func testUserProfileNoProgressOrGroup(ctx context.Context, t *testing.T, db *gorm.DB, repo *repository.UserRepository) {
	t.Run("NoProgressOrGroup", func(t *testing.T) {
		userID2 := uuid.New()
		user2 := &models.User{
			ID:        userID2,
			FirstName: "Jane",
			LastName:  "Doe",
			Username:  "janedoe",
			Email:     "jane@example.com",
		}
		db.Create(user2)

		profile, err := repo.GetUserProfile(ctx, userID2)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if profile == nil {
			t.Fatalf("Expected profile to be non-nil")
		}
		if profile.Energy != 0 {
			t.Errorf("Expected Energy 0, got %d", profile.Energy)
		}
		if profile.Group != nil {
			t.Errorf("Expected Group to be nil, got %v", profile.Group)
		}
	})
}

func testUserProfileNotFound(ctx context.Context, t *testing.T, repo *repository.UserRepository) {
	t.Run("NotFound", func(t *testing.T) {
		profile, err := repo.GetUserProfile(ctx, uuid.New())
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
		if profile != nil {
			t.Errorf("Expected profile to be nil, got %v", profile)
		}
	})
}

func TestUserRepository_UpdateUserProfile(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	user := &models.User{
		ID:        userID,
		FirstName: "Old",
		LastName:  "Name",
		Email:     "old@example.com",
	}
	db.Create(user)

	err := repo.UpdateUserProfile(ctx, userID, "New", "Name")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	var updatedUser models.User
	db.First(&updatedUser, userID)
	if updatedUser.FirstName != "New" {
		t.Errorf("Expected FirstName 'New', got %s", updatedUser.FirstName)
	}
	if updatedUser.LastName != "Name" {
		t.Errorf("Expected LastName 'Name', got %s", updatedUser.LastName)
	}
}

func TestUserRepository_GetAllUsers(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	db.Create(&models.User{ID: uuid.New(), Email: "user1@example.com"})
	db.Create(&models.User{ID: uuid.New(), Email: "user2@example.com"})

	users, err := repo.GetAllUsers(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

func TestUserRepository_DeleteUser(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	userID := uuid.New()
	db.Create(&models.User{ID: userID, Email: "delete@example.com"})

	err := repo.DeleteUser(ctx, userID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	var count int64
	db.Unscoped().Model(&models.User{}).Where("id = ?", userID).Count(&count)
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}
}

func TestUserRepository_GetLeaderboard(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	u1 := uuid.New()
	u2 := uuid.New()
	adminID := uuid.New()

	db.Create(&models.User{ID: u1, Role: "user", Email: "u1@e.com"})
	db.Create(&models.UserProgress{UserID: u1, XP: 100})

	db.Create(&models.User{ID: u2, Role: "user", Email: "u2@e.com"})
	db.Create(&models.UserProgress{UserID: u2, XP: 200})

	db.Create(&models.User{ID: adminID, Role: "admin", Email: "admin@e.com"})
	db.Create(&models.UserProgress{UserID: adminID, XP: 300})

	leaderboard, err := repo.GetLeaderboard(ctx, 10)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(leaderboard) != 2 {
		t.Errorf("Expected 2 users in leaderboard, got %d", len(leaderboard))
	}
	if leaderboard[0].ID != u2 {
		t.Errorf("Expected top user to be %v, got %v", u2, leaderboard[0].ID)
	}
	if leaderboard[1].ID != u1 {
		t.Errorf("Expected second user to be %v, got %v", u1, leaderboard[1].ID)
	}
}

func TestUserRepository_GetUserRank(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	u1 := uuid.New()
	u2 := uuid.New()
	u3 := uuid.New()

	db.Create(&models.User{ID: u1, Role: "user", Email: "u1@e.com"})
	db.Create(&models.UserProgress{UserID: u1, XP: 100})

	db.Create(&models.User{ID: u2, Role: "user", Email: "u2@e.com"})
	db.Create(&models.UserProgress{UserID: u2, XP: 200})

	db.Create(&models.User{ID: u3, Role: "user", Email: "u3@e.com"})
	db.Create(&models.UserProgress{UserID: u3, XP: 150})

	rank1, err := repo.GetUserRank(ctx, u2) // Should be 1
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rank1 != 1 {
		t.Errorf("Expected rank 1, got %d", rank1)
	}

	rank2, err := repo.GetUserRank(ctx, u3) // Should be 2
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rank2 != 2 {
		t.Errorf("Expected rank 2, got %d", rank2)
	}

	rank3, err := repo.GetUserRank(ctx, u1) // Should be 3
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rank3 != 3 {
		t.Errorf("Expected rank 3, got %d", rank3)
	}

	t.Run("NotFound", func(t *testing.T) {
		rank, err := repo.GetUserRank(ctx, uuid.New())
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
		if rank != 0 {
			t.Errorf("Expected rank 0, got %d", rank)
		}
	})
}
