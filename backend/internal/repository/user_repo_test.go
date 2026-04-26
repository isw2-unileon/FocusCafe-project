package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/repository"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// Create a connection to an in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// AutoMigrate to creat the necessary tables
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
