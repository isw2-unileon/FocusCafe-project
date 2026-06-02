package database

import (
	"testing"

	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSupabaseSchemaValidation(t *testing.T) {
	// Use an in-memory SQLite database to validate schema without requiring a live PostgreSQL instance.
	var err error
	DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// AutoMigrate all models to create tables
	err = DB.AutoMigrate(
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
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	validations := []struct {
		model   interface{}
		columns []string
	}{
		{&models.User{}, []string{"id", "first_name", "last_name", "email", "username", "role"}},
		{&models.UserProgress{}, []string{"user_id", "energy", "level"}},
		{&models.StudyMaterial{}, []string{"id", "user_id", "title", "subject_name", "file_path"}},
		{&models.StudySession{}, []string{"id", "user_id", "material_id", "duration_minutes", "start_time", "status"}},
		{&models.CafeOrder{}, []string{"id", "name", "category", "energy_cost", "reward_xp"}},
		{&models.UserOrder{}, []string{"id", "user_id", "cafe_order_id", "status"}},
		{&models.Quiz{}, []string{"id", "session_id", "generated_at"}},
		{&models.Question{}, []string{"id", "quiz_id", "question_text", "correct_answer"}},
	}

	migrator := DB.Migrator()

	for _, v := range validations {
		// Get table name safely
		stmt := &gorm.Statement{DB: DB}
		_ = stmt.Parse(v.model)
		tableName := stmt.Schema.Table

		t.Run("Table_"+tableName, func(t *testing.T) {
			if !migrator.HasTable(v.model) {
				t.Errorf("ERROR: Table '%s' does not exist", tableName)
				return
			}

			for _, col := range v.columns {
				if !migrator.HasColumn(v.model, col) {
					t.Errorf("ERROR: Column '%s' does not exist in '%s'", col, tableName)
				}
			}
		})
	}
}
