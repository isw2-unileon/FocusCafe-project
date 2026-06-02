package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
)

// mockStudyRepository implements the StudyRepository interface for testing purposes.
type mockStudyRepository struct {
	createMaterialErr error
	createSessionErr  error
	getSessionErr     error
	saveFullQuizErr   error
	updateProgressErr error

	capturedMaterial  *models.StudyMaterial
	capturedSession   *models.StudySession
	capturedQuizName  string
	capturedQuestions []models.Question
	capturedEnergy    int

	mockSessionResult *models.StudySession
	mockProgressValue int
}

func (m *mockStudyRepository) CreateMaterial(ctx context.Context, material *models.StudyMaterial) error {
	if m.createMaterialErr != nil {
		return m.createMaterialErr
	}
	material.ID = 42 // Mocked database ID assignment
	m.capturedMaterial = material
	return nil
}

func (m *mockStudyRepository) CreateSession(ctx context.Context, session *models.StudySession) error {
	if m.createSessionErr != nil {
		return m.createSessionErr
	}
	session.ID = 100 // Mocked database ID assignment
	m.capturedSession = session
	return nil
}

func (m *mockStudyRepository) GetSessionWithMaterial(ctx context.Context, sessionID uint64) (*models.StudySession, error) {
	if m.getSessionErr != nil {
		return nil, m.getSessionErr
	}
	if m.mockSessionResult != nil {
		return m.mockSessionResult, nil
	}
	return &models.StudySession{ID: sessionID}, nil
}

func (m *mockStudyRepository) SaveFullQuiz(ctx context.Context, sessionID uint64, quizName string, questions []models.Question) error {
	if m.saveFullQuizErr != nil {
		return m.saveFullQuizErr
	}
	m.capturedQuizName = quizName
	m.capturedQuestions = questions
	return nil
}

func (m *mockStudyRepository) UpdateUserProgress(ctx context.Context, userID uuid.UUID, sessionID uint64, energy int) (int, error) {
	if m.updateProgressErr != nil {
		return 0, m.updateProgressErr
	}
	m.capturedEnergy = energy
	return m.mockProgressValue, nil
}

func TestStartStudySession_Success(t *testing.T) {
	repo := &mockStudyRepository{}
	service := NewStudyService(repo)

	userID := uuid.New()
	fileName := "operating_systems.pdf"
	subjectName := "Computer Science"
	filePath := "/uploads/operating_systems.pdf"
	content := "This is the parsed content of the operating systems book."

	ctx := context.Background()
	session, materialID, err := service.StartStudySession(ctx, userID, fileName, subjectName, filePath, content)
	if err != nil {
		t.Fatalf("Unexpected error starting session: %v", err)
	}

	if materialID != 42 {
		t.Errorf("Expected material ID to be 42, got %d", materialID)
	}

	if session.ID != 100 {
		t.Errorf("Expected session ID to be 100, got %d", session.ID)
	}

	if session.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, session.UserID)
	}

	if repo.capturedMaterial == nil || repo.capturedMaterial.Title != fileName {
		t.Errorf("Material properties were not set or persisted correctly")
	}

	if repo.capturedSession == nil || repo.capturedSession.Status != "STUDYING" {
		t.Errorf("Session state properties were not initialized correctly")
	}
}

func TestStartStudySession_CreateMaterialFailure(t *testing.T) {
	expectedErr := errors.New("database connection failed on material insert")
	repo := &mockStudyRepository{
		createMaterialErr: expectedErr,
	}
	service := NewStudyService(repo)

	ctx := context.Background()
	_, _, err := service.StartStudySession(ctx, uuid.New(), "file.pdf", "Subject", "path", "content")

	if err == nil {
		t.Fatal("Expected an error during material insertion, but got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestStartStudySession_CreateSessionFailure(t *testing.T) {
	expectedErr := errors.New("database connection failed on session insert")
	repo := &mockStudyRepository{
		createSessionErr: expectedErr,
	}
	service := NewStudyService(repo)

	ctx := context.Background()
	_, _, err := service.StartStudySession(ctx, uuid.New(), "file.pdf", "Subject", "path", "content")

	if err == nil {
		t.Fatal("Expected an error during session insertion, but got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestGetSessionWithMaterial_Success(t *testing.T) {
	expectedSession := &models.StudySession{
		ID: 55,
		Material: models.StudyMaterial{
			Title: "algorithms.pdf",
		},
	}
	repo := &mockStudyRepository{
		mockSessionResult: expectedSession,
	}
	service := NewStudyService(repo)

	ctx := context.Background()
	result, err := service.GetSessionWithMaterial(ctx, 55)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.ID != 55 {
		t.Errorf("Expected session ID 55, got %d", result.ID)
	}

	if result.Material.Title != "algorithms.pdf" {
		t.Errorf("Expected preloaded material title 'algorithms.pdf', got %s", result.Material.Title)
	}
}

func TestGetSessionWithMaterial_Failure(t *testing.T) {
	expectedErr := errors.New("session not found")
	repo := &mockStudyRepository{
		getSessionErr: expectedErr,
	}
	service := NewStudyService(repo)

	ctx := context.Background()
	_, err := service.GetSessionWithMaterial(ctx, 999)

	if err == nil {
		t.Fatal("Expected error when fetching missing session, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestSaveQuiz_Success(t *testing.T) {
	repo := &mockStudyRepository{}
	service := NewStudyService(repo)

	quizJSON := `{
		"quiz_name": "Go Basics Quiz",
		"questions": [
			{
				"question_text": "What is a goroutine?",
				"option_a": "A thread",
				"option_b": "A lightweight thread managed by Go runtime",
				"option_c": "A compiler directive",
				"option_d": "A package",
				"correct_answer": "B",
				"explanation": "Goroutines are multiplexed onto operating system threads."
			}
		]
	}`

	ctx := context.Background()
	err := service.SaveQuiz(ctx, 100, quizJSON)
	if err != nil {
		t.Fatalf("Unexpected error saving quiz: %v", err)
	}

	if repo.capturedQuizName != "Go Basics Quiz" {
		t.Errorf("Expected quiz name 'Go Basics Quiz', got %s", repo.capturedQuizName)
	}

	if len(repo.capturedQuestions) != 1 {
		t.Fatalf("Expected 1 question parsed, got %d", len(repo.capturedQuestions))
	}

	q := repo.capturedQuestions[0]
	if q.QuestionText != "What is a goroutine?" {
		t.Errorf("Unexpected question text parsed: %s", q.QuestionText)
	}

	if q.CorrectAnswer != "B" {
		t.Errorf("Expected correct option B, got %s", q.CorrectAnswer)
	}
}

func TestSaveQuiz_InvalidJSON(t *testing.T) {
	repo := &mockStudyRepository{}
	service := NewStudyService(repo)

	// Sending malformed JSON payload
	invalidJSON := `{"quiz_name": "Broken Quiz", "questions": [`

	ctx := context.Background()
	err := service.SaveQuiz(ctx, 100, invalidJSON)

	if err == nil {
		t.Fatal("Expected unmarshaling parsing error, got nil")
	}
}

func TestSaveQuiz_RepoFailure(t *testing.T) {
	expectedErr := errors.New("failed to write quiz questions transaction")
	repo := &mockStudyRepository{
		saveFullQuizErr: expectedErr,
	}
	service := NewStudyService(repo)

	validJSON := `{"quiz_name": "Go Basics Quiz", "questions": []}`

	ctx := context.Background()
	err := service.SaveQuiz(ctx, 100, validJSON)

	if err == nil {
		t.Fatal("Expected database storage failure propagation, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestUpdateUserProgress_Success(t *testing.T) {
	repo := &mockStudyRepository{
		mockProgressValue: 250,
	}
	service := NewStudyService(repo)

	ctx := context.Background()
	energyGain, err := service.UpdateUserProgress(ctx, uuid.New(), 100, 10)
	if err != nil {
		t.Fatalf("Unexpected progress update error: %v", err)
	}

	if energyGain != 250 {
		t.Errorf("Expected resulting reward value 250, got %d", energyGain)
	}

	if repo.capturedEnergy != 10 {
		t.Errorf("Expected parameter passing level to match 10, got %d", repo.capturedEnergy)
	}
}

func TestUpdateUserProgress_Failure(t *testing.T) {
	expectedErr := errors.New("transaction rollbacked on progress limit validation")
	repo := &mockStudyRepository{
		updateProgressErr: expectedErr,
	}
	service := NewStudyService(repo)

	ctx := context.Background()
	_, err := service.UpdateUserProgress(ctx, uuid.New(), 100, 10)

	if err == nil {
		t.Fatal("Expected transaction error to bubble up, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}
