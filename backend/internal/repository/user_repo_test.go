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

func TestUserRepository_GetUserProfile(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewUserRepository(db)

	testID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	// Insert user
	user := models.User{
		ID:    testID,
		Email: "test@focus.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Error creating user: %v", err)
	}

	progress := models.UserProgress{
		UserID: testID,
		Energy: 500,
		XP:     100,
	}
	if err := db.Create(&progress).Error; err != nil {
		t.Fatalf("Error creating progress: %v", err)
	}
	tests := []struct {
		name    string
		id      uuid.UUID
		want    *domain.UserProfile
		wantErr bool
	}{
		{
			name: "Success: User profile retrieved successfully",
			id:   uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			want: &domain.UserProfile{
				ID:     testID,
				Energy: 500,
				XP:     100,
				Level:  1,
			},
			wantErr: false,
		},
		{
			name:    "Error: User not found",
			id:      uuid.New(),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := repo.GetUserProfile(context.Background(), tt.id)

			if (gotErr != nil) != tt.wantErr {
				t.Errorf("GetUserProfile() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.ID != tt.want.ID {
					t.Errorf("ID: got %v, want %v", got.ID, tt.want.ID)
				}
				if got.Energy != tt.want.Energy {
					t.Errorf("Energy: got %d, want %d", got.Energy, tt.want.Energy)
				}
				if got.XP != tt.want.XP {
					t.Errorf("XP: got %d, want %d", got.XP, tt.want.XP)
				}
			}
		})
	}
}
