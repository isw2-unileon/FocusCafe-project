// Package config handles application configuration from environment variables.
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	Port                   string
	GinMode                string
	CORSAllowOrigin        string
	SupabaseURL            string
	SupabaseKey            string
	SupabaseServiceRoleKey string
	SupabaseJWTSecret      string
	DatabaseURL            string
	GeminiKey              string
}

// Load reads configuration from environment variables.
// Loads .env only.
func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env not found or error: %v", err)
	}

	cfg := &Config{
		Port:                   getEnv("PORT", "8080"),
		GinMode:                getEnv("GIN_MODE", "release"),
		CORSAllowOrigin:        getEnv("CORS_ALLOW_ORIGIN", "*"),
		SupabaseURL:            getEnv("SUPABASE_URL", ""),
		SupabaseKey:            getEnv("SUPABASE_KEY", ""),
		SupabaseServiceRoleKey: getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		SupabaseJWTSecret:      getEnv("SUPABASE_JWT_SECRET", ""),
		DatabaseURL:            getEnv("DATABASE_URL", ""),
		GeminiKey:              getEnv("GEMINI_API_KEY", ""),
	}

	log.Printf("Loaded Settings (Port: %s, Mode: %s)", cfg.Port, cfg.GinMode)
	log.Printf("Supabase URL: %s", cfg.SupabaseURL)
	log.Printf("Database URL: %s", cfg.DatabaseURL)

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
