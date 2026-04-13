// Package config handles application configuration from environment variables.
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration loaded from environment variables.
type Config struct {
	Port            string
	GinMode         string
	CORSAllowOrigin string
	DatabaseURL     string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	err := godotenv.Load("../../../.env")
	if err != nil {
		log.Printf("No .env file found, using environment variables")
	}
	return &Config{
		Port:            getEnv("PORT", "8080"),
		GinMode:         getEnv("GIN_MODE", "release"),
		CORSAllowOrigin: getEnv("CORS_ALLOW_ORIGIN", "*"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgresql://postgres.uhetjanenchdncjtmybm:inso2026focuscafe@aws-1-eu-west-1.pooler.supabase.com:5432/postgres"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
