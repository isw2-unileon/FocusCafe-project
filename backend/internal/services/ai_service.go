package services

import (
	"context"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GenerateQuizSystemPrompt is the prompt used to instruct the AI.
const GenerateQuizSystemPrompt = `
Eres un experto docente universitario en pedagogía. Tu tarea es generar cuestionarios de alta calidad.

**REGLAS CRÍTICAS DE SALIDA:**
1. Responde ÚNICAMENTE con un objeto JSON puro. Sin Markdown (sin etiquetas json), sin introducciones.
2. Sigue EXACTAMENTE este formato JSON:
{
  "quiz_name": "Nombre corto del tema central",
  "questions": [
    {
      "question": "¿Pregunta...?",
      "option_a": "Opción 1",
      "option_b": "Opción 2",
      "option_c": "Opción 3",
      "option_d": "Opción 4",
      "correct_answer": "A",
      "explanation": "Explicación breve de por qué es la correcta basándote en el texto."
    }
  ]
}
3. Genera 5 preguntas de dificultad media.
`

// GenerateQuiz sends the text to the AI to create questions.
func GenerateQuiz(pdfText string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		return "", err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash")

	// UNIÓN DE PROMPTS:
	// Aquí le pasamos las reglas de comportamiento (SystemPrompt)
	// junto con el texto específico del PDF.
	fullPrompt := fmt.Sprintf("%s\n\nESTE ES EL TEXTO PARA EL CUESTIONARIO:\n%s", GenerateQuizSystemPrompt, pdfText)

	resp, err := model.GenerateContent(ctx, genai.Text(fullPrompt))
	if err != nil {
		return "", err
	}

	var responseText string
	for _, cand := range resp.Candidates {
		for _, part := range cand.Content.Parts {
			responseText += fmt.Sprintf("%v", part)
		}
	}

	return responseText, nil
}
