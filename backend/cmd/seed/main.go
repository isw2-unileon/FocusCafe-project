package main

import (
	"log"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/config"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/database"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
)

func main() {
	cfg := config.Load()
	database.InitDB(cfg)
	db := database.DB

	log.Println("Migrating database...")
	err := db.AutoMigrate(
		&models.User{},
		&models.UserProgress{},
		&models.StudyMaterial{},
		&models.StudySession{},
		&models.CafeOrder{},
		&models.UserOrder{},
		&models.Quiz{},
		&models.Question{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Seeding Cafe Orders...")
	orders := []models.CafeOrder{
		{Name: "Espresso", Category: "Coffee", EnergyCost: 20, RewardXP: 10},
		{Name: "Cappuccino", Category: "Coffee", EnergyCost: 40, RewardXP: 25},
		{Name: "Croissant", Category: "Snack", EnergyCost: 30, RewardXP: 15},
		{Name: "Full Breakfast", Category: "Meal", EnergyCost: 100, RewardXP: 60},
	}

	for _, o := range orders {
		db.Where(models.CafeOrder{Name: o.Name}).FirstOrCreate(&o)
	}

	log.Println("Seeding Sample User...")
	// Note: In a real scenario, this UUID should match an existing auth.user id in Supabase
	sampleID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
	user := models.User{
		ID:        sampleID,
		FirstName: "Test",
		LastName:  "User",
		Username:  "testuser",
		Email:     "test@focuscafe.com",
		Role:      "user",
	}
	db.Where(models.User{Email: user.Email}).FirstOrCreate(&user)

	log.Println("Seeding User Progress...")
	progress := models.UserProgress{
		UserID: user.ID,
		Energy: 100,
		Level:  1,
	}
	db.Where(models.UserProgress{UserID: user.ID}).FirstOrCreate(&progress)

	log.Println("Database seeded successfully!")
}
