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
		&models.StudyMaterial{},
		&models.StudySession{},
		&models.CafeOrder{},
		&models.UserOrder{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// 3. Test: Create a User
	userID := uuid.New()
	user := models.User{
		ID:     userID,
		Name:   "Test User",
		Email:  "test@example.com",
		Level:  1,
		Energy: 100,
	}

	if err := db.Create(&user).Error; err != nil {
		t.Errorf("failed to create user: %v", err)
	}

	// 4. Test: Create a StudyMaterial for that user
	material := models.StudyMaterial{
		UserID:      userID,
		Title:       "Go Testing Guide",
		SubjectName: "Computer Science",
		FilePath:    "/tmp/test.pdf",
	}

	if err := db.Create(&material).Error; err != nil {
		t.Errorf("failed to create material: %v", err)
	}

	// 5. Verify relationship
	var foundMaterial models.StudyMaterial
	err = db.Preload("User").First(&foundMaterial, material.ID).Error
	if err != nil {
		t.Fatalf("failed to find material: %v", err)
	}

	if foundMaterial.User.Name != "Test User" {
		t.Errorf("expected related user name 'Test User', got '%s'", foundMaterial.User.Name)
	}
}
