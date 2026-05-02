// Package config handles application configuration from environment variables.
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	Port              string
	GinMode           string
	CORSAllowOrigin   string
	SupabaseURL       string
	SupabaseKey       string
	SupabaseJWTSecret string
	DatabaseURL       string
	GeminiKey         string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	_ = godotenv.Load() // Busca en el directorio de ejecución

	cfg := &Config{
		Port:              getEnv("PORT", "8080"),
		GinMode:           getEnv("GIN_MODE", "release"),
		CORSAllowOrigin:   getEnv("CORS_ALLOW_ORIGIN", "*"),
		SupabaseURL:       getEnv("SUPABASE_URL", ""),
		SupabaseKey:       getEnv("SUPABASE_KEY", ""),
		SupabaseJWTSecret: getEnv("SUPABASE_JWT_SECRET", ""),
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		GeminiKey:         getEnv("GEMINI_API_KEY", ""),
	}

	log.Printf("Configuración cargada (Puerto: %s, Modo: %s)", cfg.Port, cfg.GinMode)

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
