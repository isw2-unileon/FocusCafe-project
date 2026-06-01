package repository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/repository"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
	"gorm.io/driver/sqlite"
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
		&models.StudyMaterial{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

// seedTestData handles the creation of initial database records for testing.
// This keeps the main test function clean and reusable.
func seedTestData(t *testing.T, db *gorm.DB, id uuid.UUID) {
	t.Helper() // Marks this function as a test helper

	user := models.User{
		ID:    id,
		Email: "test@focus.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create seed user: %v", err)
	}

	progress := models.UserProgress{
		UserID: id,
		Energy: 500,
		XP:     100,
	}
	if err := db.Create(&progress).Error; err != nil {
		t.Fatalf("Failed to create seed progress: %v", err)
	}
}

// TestUserRepository_GetUserProfile is now clean and easy to read,
// satisfying the gocognit linter requirements.
func TestUserRepository_GetUserProfile(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	testID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	// Populate the in-memory database with test data
	seedTestData(t, db, testID)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "Success: User profile found",
			id:      testID,
			wantErr: false,
		},
		{
			name:    "Error: User ID does not exist",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Logic is delegated to a specialized helper
			runProfileTest(t, repo, tt.id, tt.wantErr)
		})
	}
}

// runProfileTest executes the actual repository call and performs assertions.
func runProfileTest(t *testing.T, repo services.UserRepository, id uuid.UUID, wantErr bool) {
	t.Helper()

	got, err := repo.GetUserProfile(context.Background(), id)

	// Check if error matches expected outcome
	if (err != nil) != wantErr {
		t.Errorf("GetUserProfile() error = %v, wantErr %v", err, wantErr)
		return
	}

	// Additional assertions for success cases
	if !wantErr && got == nil {
		t.Error("GetUserProfile() returned nil profile unexpectedly")
	}
}

// TestUserRepository_GetLeaderboard verifies that only non-admin users are returned,
// ordered by XP descending.
func TestUserRepository_GetLeaderboard(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	adminID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")

	// Create a regular user with progress
	user := models.User{
		ID:        userID,
		Email:     "user@focus.com",
		FirstName: "Alice",
		Username:  "alice",
		Role:      "user",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	userProgress := models.UserProgress{
		UserID: userID,
		Energy: 300,
		XP:     1500,
		Level:  5,
	}
	if err := db.Create(&userProgress).Error; err != nil {
		t.Fatalf("Failed to create user progress: %v", err)
	}

	// Create an admin with higher XP
	admin := models.User{
		ID:        adminID,
		Email:     "admin@focus.com",
		FirstName: "Admin",
		Username:  "admin",
		Role:      "admin",
	}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}
	adminProgress := models.UserProgress{
		UserID: adminID,
		Energy: 500,
		XP:     5000,
		Level:  10,
	}
	if err := db.Create(&adminProgress).Error; err != nil {
		t.Fatalf("Failed to create admin progress: %v", err)
	}

	// Call the leaderboard
	leaderboard, err := repo.GetLeaderboard(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetLeaderboard() unexpected error: %v", err)
	}

	// Should return exactly 1 user (the regular user, not the admin)
	if len(leaderboard) != 1 {
		t.Errorf("GetLeaderboard() returned %d users, want 1", len(leaderboard))
	}

	if len(leaderboard) > 0 {
		if leaderboard[0].ID != userID {
			t.Errorf("GetLeaderboard() first user ID = %v, want %v", leaderboard[0].ID, userID)
		}
		if leaderboard[0].FirstName != "Alice" {
			t.Errorf("GetLeaderboard() first user name = %v, want Alice", leaderboard[0].FirstName)
		}
	}
}

// TestUserRepository_GetUserRank verifies the 1-based rank calculation.
func TestUserRepository_GetUserRank(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	aliceID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440010")
	bobID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440011")
	charlieID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440012")

	createUserWithProgress := func(id uuid.UUID, name string, role string, xp int) {
		u := models.User{
			ID:        id,
			Email:     name + "@focus.com",
			FirstName: name,
			Username:  name,
			Role:      role,
		}
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("Failed to create user %s: %v", name, err)
		}
		p := models.UserProgress{UserID: id, Energy: 100, XP: xp, Level: 1}
		if err := db.Create(&p).Error; err != nil {
			t.Fatalf("Failed to create progress %s: %v", name, err)
		}
	}

	// Alice: highest XP (rank 1)
	createUserWithProgress(aliceID, "Alice", "user", 3000)
	// Bob: middle XP (rank 2)
	createUserWithProgress(bobID, "Bob", "user", 2000)
	// Charlie: lowest XP (rank 3)
	createUserWithProgress(charlieID, "Charlie", "user", 1000)

	// Test Alice rank (should be 1)
	rank, err := repo.GetUserRank(context.Background(), aliceID)
	if err != nil {
		t.Fatalf("GetUserRank(alice) error: %v", err)
	}
	if rank != 1 {
		t.Errorf("GetUserRank(alice) = %d, want 1", rank)
	}

	// Test Bob rank (should be 2)
	rank, err = repo.GetUserRank(context.Background(), bobID)
	if err != nil {
		t.Fatalf("GetUserRank(bob) error: %v", err)
	}
	if rank != 2 {
		t.Errorf("GetUserRank(bob) = %d, want 2", rank)
	}

	// Test Charlie rank (should be 3)
	rank, err = repo.GetUserRank(context.Background(), charlieID)
	if err != nil {
		t.Fatalf("GetUserRank(charlie) error: %v", err)
	}
	if rank != 3 {
		t.Errorf("GetUserRank(charlie) = %d, want 3", rank)
	}
}

// TestUserRepository_GetUserRank_NoProgress verifies error when user has no progress record.
func TestUserRepository_GetUserRank_NoProgress(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440020")
	user := models.User{
		ID:        userID,
		Email:     "noprogress@focus.com",
		FirstName: "NoProgress",
		Username:  "noprogress",
		Role:      "user",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err := repo.GetUserRank(context.Background(), userID)
	if err == nil {
		t.Error("GetUserRank() expected error for user without progress, got nil")
	}
}
