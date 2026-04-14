package models

// CafeOrder represents an item available in the cafe menu.
type CafeOrder struct {
	ID         uint64 `json:"id" gorm:"primaryKey;autoIncrement"`
	Name       string `json:"name" gorm:"not null"`
	Category   string `json:"category"` // e.g., "Coffee", "Snack", "Meal"
	EnergyCost int64  `json:"energy_cost" gorm:"not null"`
	RewardXP   int64  `json:"reward_xp" gorm:"not null"`
}
