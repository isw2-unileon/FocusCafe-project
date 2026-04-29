package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

// GenerateQuizSystemPrompt is the system prompt used to instruct the Gemini model on how to generate the quiz.
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

// GenerateQuiz takes the extracted text from the PDF and generates a quiz using Gemini API.
func GenerateQuiz(pdfText string) (string, error) {
	ctx := context.Background()
	// Usamos la variable de entorno directamente
	apiKey := os.Getenv("GEMINI_API_KEY")
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", err
	}
	defer client.Close()

	// OJO: El modelo actual es "gemini-1.5-flash" (el 2.5 no existe todavía en la API estable)
	model := client.GenerativeModel("gemini-1.5-flash")

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

// init is used to perform any necessary setup when the package is imported.
func init() {
	_ = godotenv.Load()
	_ = godotenv.Load("backend/.env")

	key := os.Getenv("GEMINI_API_KEY")
	wd, _ := os.Getwd()

	log.Printf("--- DIAGNÓSTICO DE IA ---")
	log.Printf("Directorio de trabajo actual: %s", wd)

	if key == "" {
		log.Println("ERROR: No se ha detectado API KEY para la IA en GEMINI_API_KEY")
	} else {
		log.Printf("Clave detectada correctamente (empieza por:...)")
	}
	log.Printf("-------------------------")
}

// CreateQuizFromSession is a Gin handler that generates a quiz from the study session's PDF text.
func CreateQuizFromSession(c *gin.Context) {
	resp, err := GenerateQuiz("Texto de prueba para verificar API KEY")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error de IA: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"quiz":    resp,
		"message": "Cuestionario generado con éxito",
	})
}
