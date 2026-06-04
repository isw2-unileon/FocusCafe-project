package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/repository"

	"github.com/glebarez/sqlite"

	"gorm.io/gorm"
)

func setupStudyTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.UserProgress{},
		&models.StudyMaterial{},
		&models.StudySession{},
		&models.Quiz{},
		&models.Question{},
	)
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func TestStudyRepository_CreateMaterial(t *testing.T) {
	db := setupStudyTestDB(t)
	repo := repository.NewStudyRepository(db)
	userID := uuid.New()
	material := &models.StudyMaterial{
		UserID:      userID,
		Title:       "Test Material",
		SubjectName: "Test Subject",
		FilePath:    "/path/to/file",
		Content:     "Some content",
	}

	err := repo.CreateMaterial(context.Background(), material)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if material.ID == 0 {
		t.Errorf("Expected non-zero material ID")
	}

	var saved models.StudyMaterial
	if err := db.First(&saved, material.ID).Error; err != nil {
		t.Errorf("Could not find saved material: %v", err)
	}
	if saved.Title != material.Title {
		t.Errorf("Expected title %s, got %s", material.Title, saved.Title)
	}
}

func TestStudyRepository_CreateSession(t *testing.T) {
	db := setupStudyTestDB(t)
	repo := repository.NewStudyRepository(db)
	userID := uuid.New()
	session := &models.StudySession{
		UserID:          userID,
		MaterialID:      1,
		DurationMinutes: 25,
		StartTime:       time.Now(),
		Status:          "STUDYING",
	}

	err := repo.CreateSession(context.Background(), session)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if session.ID == 0 {
		t.Errorf("Expected non-zero session ID")
	}
}

func TestStudyRepository_GetSessionWithMaterial(t *testing.T) {
	db := setupStudyTestDB(t)
	repo := repository.NewStudyRepository(db)
	userID := uuid.New()

	material := models.StudyMaterial{UserID: userID, Title: "Associated Material"}
	db.Create(&material)

	session := models.StudySession{UserID: userID, MaterialID: material.ID, Status: "STUDYING"}
	db.Create(&session)

	t.Run("Success", func(t *testing.T) {
		got, err := repo.GetSessionWithMaterial(context.Background(), session.ID)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if got == nil {
			t.Fatalf("Expected session to be non-nil")
		}
		if got.ID != session.ID {
			t.Errorf("Expected ID %d, got %d", session.ID, got.ID)
		}
		if got.Material.Title != "Associated Material" {
			t.Errorf("Expected preloaded material title 'Associated Material', got %s", got.Material.Title)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		got, err := repo.GetSessionWithMaterial(context.Background(), 9999)
		if err == nil {
			t.Errorf("Expected error for non-existent session, got nil")
		}
		if got != nil {
			t.Errorf("Expected nil session, got %v", got)
		}
	})
}

func TestStudyRepository_SaveFullQuiz(t *testing.T) {
	db := setupStudyTestDB(t)
	repo := repository.NewStudyRepository(db)
	sessionID := uint64(1)
	quizName := "Test Quiz"
	questions := []models.Question{
		{QuestionText: "Q1", OptionA: "A", OptionB: "B", OptionC: "C", OptionD: "D", CorrectAnswer: "A"},
		{QuestionText: "Q2", OptionA: "A", OptionB: "B", OptionC: "C", OptionD: "D", CorrectAnswer: "B"},
	}

	t.Run("Success", func(t *testing.T) {
		err := repo.SaveFullQuiz(context.Background(), sessionID, quizName, questions)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		var savedQuiz models.Quiz
		if err := db.Where("session_id = ?", sessionID).First(&savedQuiz).Error; err != nil {
			t.Errorf("Failed to find saved quiz: %v", err)
		}

		var count int64
		db.Model(&models.Question{}).Where("quiz_id = ?", savedQuiz.ID).Count(&count)
		if count != 2 {
			t.Errorf("Expected 2 questions, got %d", count)
		}
	})
}

func TestStudyRepository_UpdateUserProgress(t *testing.T) {
	db := setupStudyTestDB(t)
	repo := repository.NewStudyRepository(db)
	userID := uuid.New()
	sessionID := uint64(1)

	// Setup: Session must exist
	db.Create(&models.StudySession{ID: sessionID, UserID: userID})

	t.Run("Success - Create Progress", func(t *testing.T) {
		newEnergy, err := repo.UpdateUserProgress(context.Background(), userID, sessionID, 50)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if newEnergy != 50 {
			t.Errorf("Expected energy 50, got %d", newEnergy)
		}

		var p models.UserProgress
		db.Where("user_id = ?", userID).First(&p)
		if p.Energy != 50 {
			t.Errorf("Expected DB energy 50, got %d", p.Energy)
		}
	})

	t.Run("Success - Update Existing Progress", func(t *testing.T) {
		newEnergy, err := repo.UpdateUserProgress(context.Background(), userID, sessionID, 100)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if newEnergy != 150 { // 50 + 100
			t.Errorf("Expected energy 150, got %d", newEnergy)
		}
	})

	t.Run("Error - Session Not Found", func(t *testing.T) {
		_, err := repo.UpdateUserProgress(context.Background(), userID, 999, 10)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
		if err.Error() != "study session not found" {
			t.Errorf("Expected 'study session not found', got %v", err)
		}
	})

	t.Run("Error - Session Belongs to Other User", func(t *testing.T) {
		otherUser := uuid.New()
		_, err := repo.UpdateUserProgress(context.Background(), otherUser, sessionID, 10)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}
