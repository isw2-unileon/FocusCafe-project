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
		&models.Group{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

// seedTestData handles the creation of initial database records for testing.
func seedTestData(t *testing.T, db *gorm.DB, id uuid.UUID, email string, progress *models.UserProgress, groupID *int64) {
	t.Helper() // Marks this function as a test helper

	user := models.User{
		ID:        id,
		Email:     email,
		FirstName: "Test",
		LastName:  "User",
		Username:  "testuser",
		GroupID:   groupID,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create seed user: %v", err)
	}

	if progress != nil {
		progress.UserID = id
		if err := db.Create(progress).Error; err != nil {
			t.Fatalf("Failed to create seed progress: %v", err)
		}
	}
}

func TestUserRepository_GetUserProfile(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	// Populate the in-memory database with test data
	uNormal := uuid.New()
	uWithGroup := uuid.New()
	uWoProgress := uuid.New()

	groupID := int64(99)
	db.Create(&models.Group{ID: groupID, Name: "Cafesss", InviteCode: "CAFE12"})

	seedTestData(t, db, uNormal, "normal@focus.com", &models.UserProgress{Energy: 300, XP: 150, Level: 4}, nil)
	seedTestData(t, db, uWithGroup, "group@focus.com", &models.UserProgress{Energy: 100, XP: 10, Level: 1}, &groupID)
	seedTestData(t, db, uWoProgress, "worogress@focus.com", nil, nil)

	tests := []struct {
		name         string
		id           uuid.UUID
		wantErr      bool
		expectGroup  bool
		expectEnergy int
	}{
		{
			name:         "Success: User profile found",
			id:           uNormal,
			wantErr:      false,
			expectGroup:  false,
			expectEnergy: 300,
		},
		{
			name:         "Error: User ID does not exist",
			id:           uuid.New(),
			wantErr:      true,
			expectGroup:  false,
			expectEnergy: 0,
		},
		{
			name:         "Success: User with Group",
			id:           uWithGroup,
			wantErr:      false,
			expectGroup:  true,
			expectEnergy: 100,
		},
		{
			name:         "Success: User Without Progress",
			id:           uWoProgress,
			wantErr:      false,
			expectGroup:  false,
			expectEnergy: 0,
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

func TestUserRepository_UpdateUserProfile(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	id := uuid.New()

	seedTestData(t, db, id, "update@test.com", nil, nil)

	t.Run("Success: Update fields", func(t *testing.T) {
		err := repo.UpdateUserProfile(context.Background(), id, "NEw", "Surname")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var updated models.User
		db.First(&updated, id)
		if updated.FirstName != "NEw" || updated.LastName != "Surname" {
			t.Errorf("profile updates were not saved correctly: %v", updated)
		}
	})
}

func TestUserRepository_GetAllUsers(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	seedTestData(t, db, uuid.New(), "all1@test.com", nil, nil)
	seedTestData(t, db, uuid.New(), "all2@test.com", nil, nil)

	t.Run("Success: Retrieve all users", func(t *testing.T) {
		users, err := repo.GetAllUsers(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(users) < 2 {
			t.Errorf("expected at least 2 users, got %d", len(users))
		}
	})
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	targetEmail := "find_me@focus.com"
	id := uuid.New()

	seedTestData(t, db, id, targetEmail, nil, nil)

	var existingUser models.User
	db.Preload("Progress").Preload("Group").First(&existingUser, id)

	tests := []struct {
		name    string // description of this test case
		email   string
		want    *models.User
		wantErr bool
	}{
		{"Success: Existing email", targetEmail, &existingUser, false},
		{"Error: Non-existing email", "fake@focus.com", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetUserByEmail(context.Background(), tt.email)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetUserByEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got.Email != tt.email {
				t.Errorf("expected email %s, got %s", tt.email, got.Email)
			}
		})
	}
}

func TestUserRepository_CreateAndDeleteUser(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	id := uuid.New()

	newUser := &models.User{
		ID:        id,
		FirstName: "Juan",
		Email:     "juan@focus.com",
	}

	t.Run("Success: Create user", func(t *testing.T) {
		err := repo.CreateUser(context.Background(), newUser)
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		var check models.User
		if err := db.First(&check, id).Error; err != nil {
			t.Error("user was not written to database")
		}
	})

	t.Run("Success: Permanent delete (Unscoped)", func(t *testing.T) {
		err := repo.DeleteUser(context.Background(), id)
		if err != nil {
			t.Fatalf("failed to delete user: %v", err)
		}

		var check models.User
		err = db.Unscoped().First(&check, id).Error
		if err == nil {
			t.Error("expected user to be permanently removed from database")
		}
	})
}
