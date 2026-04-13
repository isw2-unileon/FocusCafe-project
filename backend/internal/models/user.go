package models

import "gorm.io/gorm"

// User represents a student in the system with their game statistics.
type User struct {
	gorm.Model
	Name      string      `json:"name"`
	Energy    int         `json:"energy" gorm:"default:500"`
	MaxEnergy int         `json:"max_energy" gorm:"default:500"`
	Level     int         `json:"level" gorm:"default:1"`
	Orders    []UserOrder `json:"orders,omitempty" gorm:"foreignKey:UserID"`
}
