package database

import (
	"testing"

	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestORMModels(t *testing.T) {
	// 1. Setup in-memory SQLite for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// 2. Run AutoMigrate for all models
	err = db.AutoMigrate(
		&models.User{},
		&models.UserProgress{},
		&models.StudyMaterial{},
		&models.StudySession{},
		&models.CafeOrder{},
		&models.UserOrder{},
		&models.Quiz{},
		&models.Question{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// 3. Test: Create a User
	userID := uuid.New()
	user := models.User{
		ID:        userID,
		FirstName: "Test",
		LastName:  "User",
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      "user",
	}

	if err := db.Create(&user).Error; err != nil {
		t.Errorf("failed to create user: %v", err)
	}

	// 4. Test: Create UserProgress for that user
	progress := models.UserProgress{
		UserID: userID,
		Energy: 100,
		Level:  1,
	}
	if err := db.Create(&progress).Error; err != nil {
		t.Errorf("failed to create user progress: %v", err)
	}

	// 5. Test: Create a StudyMaterial for that user
	material := models.StudyMaterial{
		UserID:      userID,
		Title:       "Go Testing Guide",
		SubjectName: "Computer Science",
		FilePath:    "/tmp/test.pdf",
	}

	if err := db.Create(&material).Error; err != nil {
		t.Errorf("failed to create material: %v", err)
	}

	// 6. Verify relationship Material -> User
	var foundMaterial models.StudyMaterial
	err = db.Preload("User").First(&foundMaterial, material.ID).Error
	if err != nil {
		t.Fatalf("failed to find material: %v", err)
	}

	if foundMaterial.User.FirstName != "Test" {
		t.Errorf("expected related user first name 'Test', got '%s'", foundMaterial.User.FirstName)
	}

	// 7. Verify relationship User -> Progress
	var foundUser models.User
	err = db.Preload("Progress").First(&foundUser, "id = ?", userID).Error
	if err != nil {
		t.Fatalf("failed to find user with progress: %v", err)
	}

	if foundUser.Progress == nil || foundUser.Progress.Energy != 100 {
		t.Errorf("expected user progress energy 100, got %+v", foundUser.Progress)
	}
}
