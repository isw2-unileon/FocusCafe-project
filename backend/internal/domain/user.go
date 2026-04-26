package domain

import "github.com/google/uuid"

// UserProfile represents the profile information of a user, including their gamified stats (energy, level, XP)
type UserProfile struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	Username  string    `json:"username"`
	Energy    int       `json:"energy"`
	MaxEnergy int       `json:"max_energy"`
	XP        int       `json:"xp"`
	Level     int       `json:"level"`
}
