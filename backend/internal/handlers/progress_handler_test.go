package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestUpdateProgressHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. Setup SQL Mock
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open sqlmock: %s", err)
	}
	defer sqlDB.Close()

	gormDB, _ := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})

	userID := uuid.New()

	t.Run("Success: Earn 60 energy with 3 correct answers", func(t *testing.T) {
		// 1. setup request
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user", &auth.UserClaims{
			RegisteredClaims: jwt.RegisteredClaims{Subject: userID.String()},
		})

		input := handlers.ProgressUpdateRequest{SessionID: 1, Score: 3}
		body, _ := json.Marshal(input)
		c.Request = httptest.NewRequest("POST", "/api/user/progress", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		// 2. Setup Mocks based on the new applyProgressUpdate logic
		// First: Check StudySession (Out of transaction)
		mock.ExpectQuery(`SELECT \* FROM "study_sessions"`).
			WithArgs(uint64(1), userID, 1). // ID, UserID, LIMIT 1
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// --- TRANSACTION STARTS HERE ---
		mock.ExpectBegin()

		// Second: Get UserProgress (Inside transaction)
		// IMPORTANT: First() adds LIMIT 1, so we expect 2 arguments (userID and 1)
		mock.ExpectQuery(`SELECT \* FROM "user_progress"`).
			WithArgs(userID, 1).
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "energy"}).AddRow(userID, 10))

		// Third: Update Energy
		mock.ExpectExec(`UPDATE "user_progress"`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()
		// --- TRANSACTION ENDS HERE ---
		// 3. Execute
		handlers.UpdateProgressHandler(gormDB)(c)

		// 4. Assert
		if w.Code != http.StatusOK {
			t.Errorf("Status = %v, want %v. Body: %s", w.Code, http.StatusOK, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), `"energy_earned":60`) {
			t.Errorf("Body = %v, want to contain energy_earned:60", w.Body.String())
		}
	})
}
