package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
)

func main() {
	_ = godotenv.Load()

	// 1. Leer el PDF
	fmt.Println("Leyendo PDF...")
	texto, err := services.ReadPdf("backend/uploads/pdf_prueba.pdf")
	if err != nil {
		log.Fatal("Error PDF:", err)
	}

	// 2. Enviar a la IA
	fmt.Println("Generando Quiz con Gemini 2.5 Flash...")
	quizJSON, err := services.GenerateQuiz(texto)
	if err != nil {
		log.Fatal("Error IA:", err)
	}

	// 3. Resultado
	fmt.Println("\n¡EXITO! El JSON generado es:")
	fmt.Println(quizJSON)
}
