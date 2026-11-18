package models

import "gorm.io/gorm"

type Booking struct {
	gorm.Model
	BookingCode string `json:"booking_code"`
	Phone string `json:"phone"`

	// relasi
	UserId int `json:"user_id"`
	User User `gorm:"foreignKey:UserId" json:"user"`
	EventId int `json:"event_id"`
	Event Event `gorm:"foreignKey:EventId" json:"event"`
}