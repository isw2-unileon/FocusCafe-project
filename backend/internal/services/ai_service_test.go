package services

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var originalTransport = http.DefaultTransport

type mockTransport struct {
	mockServerURL string
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	target, err := url.Parse(m.mockServerURL)
	if err != nil {
		return nil, err
	}

	// Route the scheme and host directly to our local test server while keeping path and query intact
	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host

	return originalTransport.RoundTrip(req)
}

func TestGenerateQuiz_Success(t *testing.T) {
	backticks := "```"
	mockResponseJSON := `{
		"candidates": [
			{
				"content": {
					"parts": [
						{
							"text": "` + backticks + `json\n{\"quiz_name\": \"Software Engineering Quiz\", \"questions\": []}\n` + backticks + `"
						}
					]
				}
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "generateContent") {
			t.Errorf("Unexpected URL path called: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockResponseJSON))
	}))
	defer server.Close()

	http.DefaultTransport = &mockTransport{mockServerURL: server.URL}
	defer func() { http.DefaultTransport = originalTransport }()

	service := NewAIService("fake-api-key")
	result, err := service.GenerateQuiz("Sample content text.")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, `"quiz_name"`) {
		t.Errorf("Expected output to contain 'quiz_name', got: %s", result)
	}

	if strings.Contains(result, "```") {
		t.Errorf("Output contains markdown blocks: %s", result)
	}
}

func TestGenerateQuiz_GoogleAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": {"message": "API key expired"}}`))
	}))
	defer server.Close()

	http.DefaultTransport = &mockTransport{mockServerURL: server.URL}
	defer func() { http.DefaultTransport = originalTransport }()

	service := NewAIService("expired-key")
	_, err := service.GenerateQuiz("Sample content text.")

	if err == nil {
		t.Fatalf("Expected error due to HTTP 400, got nil")
	}

	if !strings.Contains(err.Error(), "google API error") {
		t.Errorf("Expected 'google API error', got: %v", err)
	}
}

func TestGenerateQuiz_EmptyPayloadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"candidates": []}`))
	}))
	defer server.Close()

	http.DefaultTransport = &mockTransport{mockServerURL: server.URL}
	defer func() { http.DefaultTransport = originalTransport }()

	service := NewAIService("fake-key")
	_, err := service.GenerateQuiz("Sample text.")

	if err == nil {
		t.Fatalf("Expected error for empty payload, got nil")
	}

	if !strings.Contains(err.Error(), "no response text returned") {
		t.Errorf("Expected missing text error, got: %v", err)
	}
}

func TestGenerateQuiz_MalformedJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"candidates": [malformed json string...`))
	}))
	defer server.Close()

	http.DefaultTransport = &mockTransport{mockServerURL: server.URL}
	defer func() { http.DefaultTransport = originalTransport }()

	service := NewAIService("fake-key")
	_, err := service.GenerateQuiz("Sample text.")

	if err == nil {
		t.Fatalf("Expected unmarshaling error, got nil")
	}

	if !strings.Contains(err.Error(), "error unmarshaling") {
		t.Errorf("Expected unmarshaling error context, got: %v", err)
	}
}

type errorTransport struct {
	err error
}

func (e *errorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, e.err
}

func TestGenerateQuiz_NetworkFailure(t *testing.T) {
	expectedErr := errors.New("connection refused: gemini API unreachable")

	http.DefaultTransport = &errorTransport{err: expectedErr}
	defer func() { http.DefaultTransport = originalTransport }()

	service := NewAIService("fake-key")
	_, err := service.GenerateQuiz("Sample text.")

	if err == nil {
		t.Fatal("Expected network error, got nil")
	}

	if !strings.Contains(err.Error(), "error sending request") {
		t.Errorf("Expected 'error sending request' in error, got: %v", err)
	}
}
