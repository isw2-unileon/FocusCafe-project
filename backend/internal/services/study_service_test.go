package services_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
)

// mockStudyRepository is a test double for the StudyRepository.
type mockStudyRepository struct {
	createMaterialFunc         func(ctx context.Context, material *models.StudyMaterial) error
	createSessionFunc          func(ctx context.Context, session *models.StudySession) error
	getSessionWithMaterialFunc func(ctx context.Context, sessionID uint64) (*models.StudySession, error)
	saveFullQuizFunc           func(ctx context.Context, sessionID uint64, quizName string, questions []models.Question) error
	updateUserProgressFunc     func(ctx context.Context, userID uuid.UUID, sessionID uint64, energy int) (int, error)
}

func (m *mockStudyRepository) CreateMaterial(ctx context.Context, material *models.StudyMaterial) error {
	if m.createMaterialFunc != nil {
		return m.createMaterialFunc(ctx, material)
	}
	return errors.New("createMaterial not mocked")
}

func (m *mockStudyRepository) CreateSession(ctx context.Context, session *models.StudySession) error {
	if m.createSessionFunc != nil {
		return m.createSessionFunc(ctx, session)
	}
	return errors.New("createSession not mocked")
}

func (m *mockStudyRepository) GetSessionWithMaterial(ctx context.Context, sessionID uint64) (*models.StudySession, error) {
	if m.getSessionWithMaterialFunc != nil {
		return m.getSessionWithMaterialFunc(ctx, sessionID)
	}
	return nil, errors.New("getSessionWithMaterial not mocked")
}

func (m *mockStudyRepository) SaveFullQuiz(ctx context.Context, sessionID uint64, quizName string, questions []models.Question) error {
	if m.saveFullQuizFunc != nil {
		return m.saveFullQuizFunc(ctx, sessionID, quizName, questions)
	}
	return errors.New("saveFullQuiz not mocked")
}

func (m *mockStudyRepository) UpdateUserProgress(ctx context.Context, userID uuid.UUID, sessionID uint64, energy int) (int, error) {
	if m.updateUserProgressFunc != nil {
		return m.updateUserProgressFunc(ctx, userID, sessionID, energy)
	}
	return 0, errors.New("updateUserProgress not mocked")
}

// ============================================
// TestStudyService_StartStudySession
// ============================================

func TestStudyService_StartStudySession(t *testing.T) {
	testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name             string
		mockMaterialFunc func(ctx context.Context, material *models.StudyMaterial) error
		mockSessionFunc  func(ctx context.Context, session *models.StudySession) error
		wantErr          bool
		wantMaterialID   uint64
		wantStatus       string
	}{
		{
			name: "Success: Creates material and session",
			mockMaterialFunc: func(ctx context.Context, material *models.StudyMaterial) error {
				material.ID = 42 // Simulate DB auto-increment
				return nil
			},
			mockSessionFunc: func(ctx context.Context, session *models.StudySession) error {
				session.ID = 7 // Simulate DB auto-increment
				return nil
			},
			wantErr:        false,
			wantMaterialID: 42,
			wantStatus:     "STUDYING",
		},
		{
			name: "Error: CreateMaterial fails",
			mockMaterialFunc: func(ctx context.Context, material *models.StudyMaterial) error {
				return errors.New("database connection lost")
			},
			mockSessionFunc: nil,
			wantErr:         true,
		},
		{
			name: "Error: CreateSession fails",
			mockMaterialFunc: func(ctx context.Context, material *models.StudyMaterial) error {
				material.ID = 1
				return nil
			},
			mockSessionFunc: func(ctx context.Context, session *models.StudySession) error {
				return errors.New("session insert failed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockStudyRepository{
				createMaterialFunc: tt.mockMaterialFunc,
				createSessionFunc:  tt.mockSessionFunc,
			}
			s := services.NewStudyService(mRepo)

			session, materialID, err := s.StartStudySession(context.Background(), testUUID, "notes.pdf", "Math", "/uploads/notes.pdf", "content")

			if (err != nil) != tt.wantErr {
				t.Errorf("StartStudySession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if materialID != tt.wantMaterialID {
				t.Errorf("StartStudySession() materialID = %d, want %d", materialID, tt.wantMaterialID)
			}
			if session.Status != tt.wantStatus {
				t.Errorf("StartStudySession() status = %v, want %v", session.Status, tt.wantStatus)
			}
			if session.UserID != testUUID {
				t.Errorf("StartStudySession() userID = %v, want %v", session.UserID, testUUID)
			}
		})
	}
}

// ============================================
// TestStudyService_GetSessionWithMaterial
// ============================================

func TestStudyService_GetSessionWithMaterial(t *testing.T) {
	expectedSession := &models.StudySession{
		ID:         1,
		UserID:     uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		MaterialID: 42,
		Status:     "STUDYING",
	}

	tests := []struct {
		name         string
		mockBehavior func(ctx context.Context, sessionID uint64) (*models.StudySession, error)
		wantErr      bool
		want         *models.StudySession
	}{
		{
			name: "Success: Returns session",
			mockBehavior: func(ctx context.Context, sessionID uint64) (*models.StudySession, error) {
				return expectedSession, nil
			},
			wantErr: false,
			want:    expectedSession,
		},
		{
			name: "Error: Session not found",
			mockBehavior: func(ctx context.Context, sessionID uint64) (*models.StudySession, error) {
				return nil, errors.New("record not found")
			},
			wantErr: true,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockStudyRepository{getSessionWithMaterialFunc: tt.mockBehavior}
			s := services.NewStudyService(mRepo)

			got, err := s.GetSessionWithMaterial(context.Background(), 1)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSessionWithMaterial() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSessionWithMaterial() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================
// TestStudyService_SaveQuiz
// ============================================

func TestStudyService_SaveQuiz(t *testing.T) {
	validQuizJSON := `{"quiz_name":"Math Quiz","questions":[{"question_text":"2+2?","option_a":"3","option_b":"4","option_c":"5","option_d":"6","correct_answer":"B","explanation":"Basic addition."}]}`

	tests := []struct {
		name         string
		quizJSON     string
		mockBehavior func(ctx context.Context, sessionID uint64, quizName string, questions []models.Question) error
		wantErr      bool
		expectedErr  string
	}{
		{
			name:     "Success: Parses and saves quiz",
			quizJSON: validQuizJSON,
			mockBehavior: func(ctx context.Context, sessionID uint64, quizName string, questions []models.Question) error {
				if quizName != "Math Quiz" {
					t.Errorf("SaveQuiz() quizName = %v, want Math Quiz", quizName)
				}
				if len(questions) != 1 {
					t.Errorf("SaveQuiz() questions len = %d, want 1", len(questions))
				}
				if questions[0].CorrectAnswer != "B" {
					t.Errorf("SaveQuiz() correctAnswer = %v, want B", questions[0].CorrectAnswer)
				}
				return nil
			},
			wantErr: false,
		},
		{
			name:     "Error: Invalid JSON",
			quizJSON: `{invalid json`,
			mockBehavior: nil,
			wantErr:      true,
			expectedErr:  "invalid character",
		},
		{
			name:     "Error: Repository fails",
			quizJSON: validQuizJSON,
			mockBehavior: func(ctx context.Context, sessionID uint64, quizName string, questions []models.Question) error {
				return errors.New("database error")
			},
			wantErr:     true,
			expectedErr: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockStudyRepository{saveFullQuizFunc: tt.mockBehavior}
			s := services.NewStudyService(mRepo)

			err := s.SaveQuiz(context.Background(), 1, tt.quizJSON)

			if (err != nil) != tt.wantErr {
				t.Errorf("SaveQuiz() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErr != "" && err != nil {
				if !contains(err.Error(), tt.expectedErr) {
					t.Errorf("SaveQuiz() error = %v, want containing %v", err, tt.expectedErr)
				}
			}
		})
	}
}

// ============================================
// TestStudyService_UpdateUserProgress
// ============================================

func TestStudyService_UpdateUserProgress(t *testing.T) {
	testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name         string
		energy       int
		mockBehavior func(ctx context.Context, userID uuid.UUID, sessionID uint64, energy int) (int, error)
		want         int
		wantErr      bool
	}{
		{
			name:   "Success: Returns new energy total",
			energy: 60,
			mockBehavior: func(ctx context.Context, userID uuid.UUID, sessionID uint64, energy int) (int, error) {
				return 560, nil
			},
			want:    560,
			wantErr: false,
		},
		{
			name:   "Error: Repository fails",
			energy: 60,
			mockBehavior: func(ctx context.Context, userID uuid.UUID, sessionID uint64, energy int) (int, error) {
				return 0, errors.New("progress update failed")
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockStudyRepository{updateUserProgressFunc: tt.mockBehavior}
			s := services.NewStudyService(mRepo)

			got, err := s.UpdateUserProgress(context.Background(), testUUID, 1, tt.energy)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateUserProgress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("UpdateUserProgress() = %d, want %d", got, tt.want)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && len(substr) > 0 && findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
