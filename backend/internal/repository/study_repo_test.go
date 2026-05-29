package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupMockDB initializes a mock database connection and returns the gorm DB instance along with the sqlmock for setting expectations.
func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open sqlmock: %s", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open gorm: %s", err)
	}

	return gormDB, mock
}

// TestStudyRepository_CreateMaterial tests the CreateMaterial method of the StudyRepository to ensure it correctly inserts a new study material into the database and returns the generated ID.
func TestStudyRepository_CreateMaterial(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := repository.NewStudyRepository(db)
	userID := uuid.New()
	material := &models.StudyMaterial{
		UserID:      userID,
		Title:       "Test Material",
		SubjectName: "Test Subject",
		FilePath:    "/path/to/file",
		Content:     "Some content",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "study_materials" .* RETURNING "id"`).
		WithArgs(userID, material.Title, material.SubjectName, material.FilePath, sqlmock.AnyArg(), material.Content).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.CreateMaterial(context.Background(), material)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), material.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestStudyRepository_CreateSession tests the CreateSession method of the StudyRepository to ensure it correctly inserts a new study session into the database and returns the generated ID.
func TestStudyRepository_CreateSession(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := repository.NewStudyRepository(db)
	userID := uuid.New()
	session := &models.StudySession{
		UserID:          userID,
		MaterialID:      1,
		DurationMinutes: 25,
		StartTime:       time.Now(),
		Status:          "STUDYING",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "study_sessions" .* RETURNING "id"`).
		WithArgs(userID, session.MaterialID, session.DurationMinutes, sqlmock.AnyArg(), nil, session.Status).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.CreateSession(context.Background(), session)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), session.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestStudyRepository_GetSessionWithMaterial tests the GetSessionWithMaterial method of the StudyRepository to ensure it correctly retrieves a study session along with its associated material.
func TestStudyRepository_GetSessionWithMaterial(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := repository.NewStudyRepository(db)
	sessionID := uint64(1)
	userID := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "study_sessions" WHERE id = \$1 .* LIMIT \$2`).
		WithArgs(sessionID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "material_id"}).AddRow(sessionID, userID, uint64(10)))

	mock.ExpectQuery(`SELECT \* FROM "study_materials" WHERE "study_materials"\."id" = \$1`).
		WithArgs(uint64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).AddRow(10, "Test Material"))

	session, err := repo.GetSessionWithMaterial(context.Background(), sessionID)
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, sessionID, session.ID)
	assert.Equal(t, "Test Material", session.Material.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestStudyRepository_SaveFullQuiz tests the SaveFullQuiz method of the StudyRepository to ensure it correctly inserts a new quiz and its questions into the database.
func TestStudyRepository_SaveFullQuiz(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := repository.NewStudyRepository(db)
	sessionID := uint64(1)
	quizName := "Test Quiz"
	questions := []models.Question{
		{QuestionText: "Q1", OptionA: "A", OptionB: "B", OptionC: "C", OptionD: "D", CorrectAnswer: "A"},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "quizzes" .* RETURNING "id","generated_at"`).
		WithArgs(sessionID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "generated_at"}).AddRow(1, time.Now()))

	mock.ExpectQuery(`INSERT INTO "questions" .* RETURNING "id"`).
		WithArgs(uint64(1), questions[0].QuestionText, questions[0].OptionA, questions[0].OptionB, questions[0].OptionC, questions[0].OptionD, questions[0].CorrectAnswer, "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectCommit()

	err := repo.SaveFullQuiz(context.Background(), sessionID, quizName, questions)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestStudyRepository_UpdateUserProgress tests the UpdateUserProgress method of the StudyRepository to ensure it correctly updates a user's progress in the database.
func TestStudyRepository_UpdateUserProgress(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := repository.NewStudyRepository(db)
	userID := uuid.New()
	sessionID := uint64(1)
	energy := 100

	// 1. Check session
	mock.ExpectQuery(`SELECT \* FROM "study_sessions" WHERE id = \$1 AND user_id = \$2 .* LIMIT \$3`).
		WithArgs(sessionID, userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(sessionID))

	mock.ExpectBegin()
	// 2. Get Progress
	mock.ExpectQuery(`SELECT \* FROM "user_progress" WHERE user_id = \$1 .* LIMIT \$2`).
		WithArgs(userID, 1).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "energy", "level", "xp"}).AddRow(userID, 50, 1, 10))

	// 3. Save Progress
	mock.ExpectExec(`UPDATE "user_progress" SET "energy"=\$1,"level"=\$2,"xp"=\$3 WHERE "user_id" = \$4`).
		WithArgs(150, 1, 10, userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	newEnergy, err := repo.UpdateUserProgress(context.Background(), userID, sessionID, energy)
	assert.NoError(t, err)
	assert.Equal(t, 150, newEnergy)
	assert.NoError(t, mock.ExpectationsWereMet())
}
