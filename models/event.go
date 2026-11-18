package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	// Id          int //otomatis
	NameEvent   string `json:"name_event" binding:"required"`
	Description string `json:"description" binding:"required"`
	Location    string `json:"location" binding:"required"`
	Image 			string `json:"image"`
	ImageId 		string `json:"imageId"`
	UserId 			int		 `json:"userId"`
	User 				User `grom:"foreignKey:UserId" json:"user"` // relasi one to many ke tabel user
	DateTime    time.Time `json:"date_time" binding:"required"`
	Bookings    []Booking `gorm:"foreignKey:EventId" json:"bookings"`
}
