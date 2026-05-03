package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// GenerateQuizSystemPrompt defines the AI's persona and the required strict JSON output format.
const GenerateQuizSystemPrompt = `Eres un experto docente. Genera un cuestionario JSON.
Responde ÚNICAMENTE el JSON puro, sin markdown.
Formato: {"quiz_name": "...", "questions": [{"question_text": "...", "option_a": "...", "option_b": "...", "option_c": "...", "option_d": "...", "correct_answer": "A", "explanation": "..."}]}
Genera 3 preguntas.`

// GenerateQuiz processes the provided text via Google Gemini API to produce a study quiz in JSON format.
func GenerateQuiz(pdfText string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	// Using gemini-2.0-flash-lite-001 as it is the most stable version for the 2026 free tier catalog.
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1/models/gemini-2.0-flash-lite-001:generateContent?key=%s", apiKey)

	// 1. Prepare the request payload.
	payload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": fmt.Sprintf("%s\n\nTEXTO:\n%s", GenerateQuizSystemPrompt, pdfText)},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature": 0.2,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshalling payload: %w", err)
	}

	// 2. Create a manual HTTP request to inject necessary headers.
	// #nosec G704 -- URL is constructed from trusted environment variables.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Set headers to identify the request type and mimic a standard browser agent.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// 3. Execute the request using a standard HTTP client.
	client := &http.Client{}
	// #nosec G704 -- Request is sent to a trusted endpoint.
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// 4. Read the response stream.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Log the raw Google API error for backend debugging purposes.
		log.Printf("❌ Detailed Google API Error: %s", string(body))
		return "", fmt.Errorf("google API error: %s", string(body))
	}

	// 5. Parse the specific response structure required by the Gemini API.
	var googleResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &googleResp); err != nil {
		return "", fmt.Errorf("error unmarshaling google response: %w", err)
	}

	if len(googleResp.Candidates) == 0 || len(googleResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response text returned from AI")
	}

	responseText := googleResp.Candidates[0].Content.Parts[0].Text

	// 6. Final cleanup to ensure only a valid JSON object is returned to the handler.
	cleaned := strings.TrimSpace(responseText)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	// Ensure we only return the content within the main JSON braces.
	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start != -1 && end != -1 && end > start {
		cleaned = cleaned[start : end+1]
	}

	return cleaned, nil
}

// init loads environmental variables from the root or the backend directory.
func init() {
	_ = godotenv.Load()
	_ = godotenv.Load("backend/.env")
}
