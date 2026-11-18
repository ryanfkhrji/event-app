package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string  `json:"name" gorm:"not null"`
	Email    string  `json:"email" gorm:"uniqueIndex;not null"`
	Password string  `json:"-" gorm:"not null"`
	Events   []Event `json:"events,omitempty"` // relasi one to many ke tabel events
}