package main

import (
	"log"

	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/config"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/database"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"gorm.io/gorm"
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

	seedCafeOrders(db)
	fillInitialOrdersForAllUsers(db)

	log.Println("Database seeded successfully!")
}

func seedCafeOrders(db *gorm.DB) {
	var count int64
	db.Model(&models.CafeOrder{}).Count(&count)
	if count > 0 {
		log.Println("CafeOrders already seeded, skipping...")
		return
	}

	log.Println("Seeding Cafe Orders...")
	orders := []models.CafeOrder{
		{RequiredLevel: 1, Name: "Espresso", Description: "A quick caffeine hit to start the day.", EnergyCost: 10, RewardXP: 20, Category: "Coffee"},
		{RequiredLevel: 1, Name: "Butter Croissant", Description: "Flaky, buttery, and perfect for early mornings.", EnergyCost: 15, RewardXP: 25, Category: "Snack"},
		{RequiredLevel: 1, Name: "Americano", Description: "Diluted espresso for a longer focus session.", EnergyCost: 12, RewardXP: 22, Category: "Coffee"},
		{RequiredLevel: 1, Name: "Petit Pain", Description: "Small bread roll, simple and effective.", EnergyCost: 10, RewardXP: 15, Category: "Snack"},
		{RequiredLevel: 1, Name: "Macchiato", Description: "Espresso stained with a touch of milk.", EnergyCost: 15, RewardXP: 28, Category: "Coffee"},
		{RequiredLevel: 2, Name: "Cappuccino", Description: "Perfectly frothed milk with a rich coffee base.", EnergyCost: 25, RewardXP: 45, Category: "Coffee"},
		{RequiredLevel: 2, Name: "Chocolate Muffin", Description: "A sweet treat to keep the sugar levels up.", EnergyCost: 30, RewardXP: 50, Category: "Snack"},
		{RequiredLevel: 2, Name: "Espresso & Croissant", Description: "The classic combo for a solid break.", EnergyCost: 35, RewardXP: 65, Category: "Meal"},
		{RequiredLevel: 2, Name: "Flat White", Description: "Velvety microfoam over a double shot.", EnergyCost: 28, RewardXP: 50, Category: "Coffee"},
		{RequiredLevel: 2, Name: "Cookie Duo", Description: "Two chocolate chip cookies for sharing (or not).", EnergyCost: 20, RewardXP: 35, Category: "Snack"},
		{RequiredLevel: 3, Name: "Caffè Latte", Description: "Smooth, milky, and easy to drink while working.", EnergyCost: 30, RewardXP: 55, Category: "Coffee"},
		{RequiredLevel: 3, Name: "Ham & Cheese Toastie", Description: "A warm snack for deep focus hours.", EnergyCost: 45, RewardXP: 80, Category: "Meal"},
		{RequiredLevel: 3, Name: "Latte & Muffin", Description: "The ultimate mid-morning fuel.", EnergyCost: 55, RewardXP: 100, Category: "Meal"},
		{RequiredLevel: 3, Name: "Iced Americano", Description: "Refreshing caffeine for high-pressure tasks.", EnergyCost: 25, RewardXP: 40, Category: "Coffee"},
		{RequiredLevel: 3, Name: "Blueberry Scone", Description: "Crumbly and filled with fresh berries.", EnergyCost: 25, RewardXP: 45, Category: "Snack"},
		{RequiredLevel: 4, Name: "Mocha", Description: "For when you need coffee but crave chocolate.", EnergyCost: 40, RewardXP: 75, Category: "Coffee"},
		{RequiredLevel: 4, Name: "Cinnamon Roll", Description: "Sticky, sweet, and incredibly satisfying.", EnergyCost: 35, RewardXP: 65, Category: "Snack"},
		{RequiredLevel: 4, Name: "Cappuccino & Toastie", Description: "A complete lunch break for hard workers.", EnergyCost: 65, RewardXP: 130, Category: "Meal"},
		{RequiredLevel: 4, Name: "Matcha Latte", Description: "Calm energy with antioxidants.", EnergyCost: 35, RewardXP: 70, Category: "Coffee"},
		{RequiredLevel: 4, Name: "Almond Croissant", Description: "Level up your pastry game.", EnergyCost: 25, RewardXP: 50, Category: "Snack"},
		{RequiredLevel: 5, Name: "Caramel Macchiato", Description: "Sweet, layered, and premium.", EnergyCost: 45, RewardXP: 90, Category: "Coffee"},
		{RequiredLevel: 5, Name: "Avocado Toast", Description: "The modern developer's fuel.", EnergyCost: 50, RewardXP: 110, Category: "Meal"},
		{RequiredLevel: 5, Name: "Mocha & Cinnamon Roll", Description: "Maximum sugar, maximum reward.", EnergyCost: 75, RewardXP: 160, Category: "Meal"},
		{RequiredLevel: 5, Name: "Cold Brew", Description: "Steeped for 20 hours for maximum clarity.", EnergyCost: 40, RewardXP: 85, Category: "Coffee"},
		{RequiredLevel: 5, Name: "Cheesecake Slice", Description: "Reward yourself for reaching level 5.", EnergyCost: 45, RewardXP: 95, Category: "Snack"},
		{RequiredLevel: 6, Name: "Double Cortado", Description: "Short, strong, and professional.", EnergyCost: 30, RewardXP: 65, Category: "Coffee"},
		{RequiredLevel: 6, Name: "Breakfast Burrito", Description: "Heavy fuel for a marathon coding session.", EnergyCost: 70, RewardXP: 150, Category: "Meal"},
		{RequiredLevel: 6, Name: "Matcha & Scone", Description: "The \"Zen Mode\" combo.", EnergyCost: 55, RewardXP: 115, Category: "Meal"},
		{RequiredLevel: 6, Name: "Chai Latte", Description: "Spicy and aromatic.", EnergyCost: 35, RewardXP: 75, Category: "Coffee"},
		{RequiredLevel: 6, Name: "Banana Bread", Description: "Toasted with a bit of butter.", EnergyCost: 30, RewardXP: 60, Category: "Snack"},
		{RequiredLevel: 7, Name: "Affogato", Description: "Espresso poured over vanilla ice cream.", EnergyCost: 40, RewardXP: 90, Category: "Coffee"},
		{RequiredLevel: 7, Name: "Club Sandwich", Description: "Triple-decker focus fuel.", EnergyCost: 75, RewardXP: 170, Category: "Meal"},
		{RequiredLevel: 7, Name: "Cold Brew & Avocado Toast", Description: "The \"Hipster Focus\" special.", EnergyCost: 85, RewardXP: 200, Category: "Meal"},
		{RequiredLevel: 7, Name: "Red Eye", Description: "Drip coffee with an espresso shot added.", EnergyCost: 50, RewardXP: 120, Category: "Coffee"},
		{RequiredLevel: 7, Name: "Red Velvet Cake", Description: "A luxurious reward.", EnergyCost: 50, RewardXP: 110, Category: "Snack"},
		{RequiredLevel: 8, Name: "Turkish Coffee", Description: "Intense, dark, and traditional.", EnergyCost: 45, RewardXP: 100, Category: "Coffee"},
		{RequiredLevel: 8, Name: "Salmon Bagel", Description: "High-protein fuel for complex logic.", EnergyCost: 80, RewardXP: 180, Category: "Meal"},
		{RequiredLevel: 8, Name: "Chai & Banana Bread", Description: "Cozy vibes for rainy workdays.", EnergyCost: 60, RewardXP: 140, Category: "Meal"},
		{RequiredLevel: 8, Name: "Dirty Chai", Description: "Chai latte with an espresso shot.", EnergyCost: 55, RewardXP: 125, Category: "Coffee"},
		{RequiredLevel: 8, Name: "Fruit Bowl", Description: "Healthy energy for a clean mind.", EnergyCost: 40, RewardXP: 90, Category: "Snack"},
		{RequiredLevel: 9, Name: "Nitro Cold Brew", Description: "Creamy texture, high caffeine.", EnergyCost: 60, RewardXP: 140, Category: "Coffee"},
		{RequiredLevel: 9, Name: "Eggs Benedict", Description: "The king of cafe breakfasts.", EnergyCost: 90, RewardXP: 220, Category: "Meal"},
		{RequiredLevel: 9, Name: "Nitro & Salmon Bagel", Description: "The \"Executive\" combo.", EnergyCost: 130, RewardXP: 320, Category: "Meal"},
		{RequiredLevel: 9, Name: "Irish Coffee (Non-alc)", Description: "Creamy and sophisticated.", EnergyCost: 50, RewardXP: 115, Category: "Coffee"},
		{RequiredLevel: 9, Name: "Tiramisu", Description: "The ultimate coffee-themed dessert.", EnergyCost: 55, RewardXP: 130, Category: "Snack"},
		{RequiredLevel: 10, Name: "Focus Café Special", Description: "A secret blend for top-tier users.", EnergyCost: 100, RewardXP: 300, Category: "Coffee"},
		{RequiredLevel: 10, Name: "Grand Feast", Description: "Coffee, pastry, and a main dish.", EnergyCost: 150, RewardXP: 500, Category: "Meal"},
		{RequiredLevel: 10, Name: "Chemex for Two", Description: "Elegant pour-over for shared focus.", EnergyCost: 80, RewardXP: 200, Category: "Coffee"},
		{RequiredLevel: 10, Name: "Protein Power Platter", Description: "For the most demanding sprints.", EnergyCost: 120, RewardXP: 350, Category: "Meal"},
		{RequiredLevel: 10, Name: "Golden Latte", Description: "Turmeric and honey for peak health.", EnergyCost: 60, RewardXP: 150, Category: "Coffee"},
	}

	if err := db.Create(&orders).Error; err != nil {
		log.Fatalf("Failed to seed orders: %v", err)
	}
}

func fillInitialOrdersForAllUsers(db *gorm.DB) {
	log.Println("Checking and filling orders for all existing users...")

	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("Could not fetch users: %v", err)
	}

	for _, user := range users {
		// 1. User Level from user_progress
		var progress models.UserProgress
		if err := db.Where("user_id = ?", user.ID).First(&progress).Error; err != nil {
			log.Printf("Skip user %s: No progress record found", user.Username)
			continue
		}

		// 2. Check pending orders count
		var pendingCount int64
		db.Model(&models.UserOrder{}).Where("user_id = ? AND status = ?", user.ID, "pending").Count(&pendingCount)

		if pendingCount < 3 {
			needed := 3 - int(pendingCount)

			// 3. Search for random cafe orders that match the user's level
			var randomCafes []models.CafeOrder
			db.Where("required_level <= ?", progress.Level).Order("RANDOM()").Limit(needed).Find(&randomCafes)

			// 4. Assign these cafe orders to the user
			for _, cafe := range randomCafes {
				newOrder := models.UserOrder{
					UserID:      user.ID,
					CafeOrderID: cafe.ID,
					Status:      "pending",
				}
				db.Create(&newOrder)
			}
			log.Printf("Added %d orders to user %s (Level %d)", needed, user.Username, progress.Level)
		}
	}
}
