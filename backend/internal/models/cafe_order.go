package models

import "gorm.io/gorm"

// CafeOrder represents an item available in the cafe menu.
type CafeOrder struct {
	gorm.Model
	Name       string `json:"name"`
	Category   string `json:"category"` // e.g., "Coffee", "Snack", "Meal"
	EnergyCost int    `json:"energy_cost"`
	RewardXP   int    `json:"reward_xp"`
}
