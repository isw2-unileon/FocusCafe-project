package database

import (
	"log"

	"github.com/isw2-unileon/FocusCafe-project/backend/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB initializes the connection to the database (Supabase/PostgreSQL).
func InitDB(cfg *config.Config) {
	var err error

	// We use the DatabaseURL from the config (DSN).
	DB, err = gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established successfully")
}
