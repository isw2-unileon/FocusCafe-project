package services_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
)

// ============================================
// TestAIService_GenerateQuiz
// ============================================

func TestAIService_GenerateQuiz(t *testing.T) {
	// Build a mock Gemini API server that returns a valid response
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		key := r.URL.Query().Get("key")
		if key != "test-api-key" {
			t.Errorf("Expected key=test-api-key, got %s", key)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resp := map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": []map[string]string{
							{"text": "```json\n{\"quiz_name\":\"Test Quiz\",\"questions\":[{\"question_text\":\"What is 2+2?\",\"option_a\":\"3\",\"option_b\":\"4\",\"option_c\":\"5\",\"option_d\":\"6\",\"correct_answer\":\"B\",\"explanation\":\"Basic math.\"}]}\n```"},
						},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// Mock server returning non-200
	mockErrorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"Invalid API key"}`))
	}))
	defer mockErrorServer.Close()

	// Mock server returning empty candidates
	mockEmptyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"candidates":[]}`))
	}))
	defer mockEmptyServer.Close()

	// Test 1: Successful response parsing
	t.Run("Success: Parses cleaned quiz JSON", func(t *testing.T) {
		rt := &geminiRoundTripper{serverURL: mockServer.URL}
		client := &http.Client{Transport: rt}

		aiSvc := services.NewAIService("test-api-key")
		aiSvc.SetHTTPClient(client)

		result, err := aiSvc.GenerateQuiz("some text")
		if err != nil {
			t.Fatalf("GenerateQuiz() unexpected error: %v", err)
		}

		if !strings.Contains(result, "Test Quiz") {
			t.Errorf("GenerateQuiz() result = %v, want containing 'Test Quiz'", result)
		}
		if !strings.Contains(result, "question_text") {
			t.Errorf("GenerateQuiz() result = %v, want containing 'question_text'", result)
		}
	})

	// Test 2: API returns error status
	t.Run("Error: API returns 401", func(t *testing.T) {
		rt := &geminiRoundTripper{serverURL: mockErrorServer.URL}
		client := &http.Client{Transport: rt}

		aiSvc := services.NewAIService("bad-key")
		aiSvc.SetHTTPClient(client)

		_, err := aiSvc.GenerateQuiz("some text")
		if err == nil {
			t.Fatal("GenerateQuiz() expected error, got nil")
		}
	})

	// Test 3: Empty candidates
	t.Run("Error: Empty candidates", func(t *testing.T) {
		rt := &geminiRoundTripper{serverURL: mockEmptyServer.URL}
		client := &http.Client{Transport: rt}

		aiSvc := services.NewAIService("test-api-key")
		aiSvc.SetHTTPClient(client)

		_, err := aiSvc.GenerateQuiz("some text")
		if err == nil {
			t.Fatal("GenerateQuiz() expected error for empty candidates, got nil")
		}
	})
}

// geminiRoundTripper intercepts requests destined for the real Gemini API and redirects them to our mock server.
type geminiRoundTripper struct {
	serverURL string
}

func (rt *geminiRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(rt.serverURL, "http://")
	return http.DefaultTransport.RoundTrip(req)
}
