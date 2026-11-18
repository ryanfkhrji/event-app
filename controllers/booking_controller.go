package controllers

import (
	"fmt"
	"net/http"
	"time"

	"example.com/event-app/config"
	"example.com/event-app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BookingInput struct {
	Phone   string `json:"phone" binding:"required"`
	EventID int `json:"event_id" binding:"required"`
}

func CreateBookingEvent(ctx *gin.Context) {
	// ambil user id yang login
	userId, _ := ctx.Get("userID")

	// ambil struct input booking
	var input BookingInput

	// ambil dari struct booking model
	var booking models.Booking

	errValidasi := ctx.ShouldBindJSON(&input);

	if errValidasi != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": errValidasi.Error(),
		})
		return
	}

	// kondisi jika user sudah pernah booking event
	bookingData := config.DB.Where("user_id = ? AND event_id = ?", userId.(int), input.EventID).First(&booking).Error

	if bookingData == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "You have already booked this event",
		})
		return
	}

	// Generate booking code
	bookingCode := fmt.Sprintf("BK-%sE%dU%d",
		time.Now().Format("20060102"),
		input.EventID,
		userId.(int),
	)

	// simpan ke database
	booking = models.Booking{
		Phone:       input.Phone,
		EventId:     input.EventID,
		BookingCode: bookingCode,
		UserId:      userId.(int),
	}

	errCreateBooing := config.DB.Create(&booking).Error

	if errCreateBooing != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create booking",
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Booking created successfully",
		"booking_code": bookingCode,
		"phone":        input.Phone,
		"user_id":      userId.(int),
		"event_id":     input.EventID,
	})
}

func GetBookingByUser(ctx *gin.Context) {
	// ambil data booking dari model struct
	var bookings []models.Booking

	// ambil user id yang login
	userId, _ := ctx.Get("userID")

	// ambil data user dan event yang sudah booking
	err := config.DB.Preload("Event").Preload("Event.User", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "name", "email")
	}).Where("user_id = ?", userId).Find(&bookings).Error

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get booking by user",
		})
		return
	}

	// response berhasil
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Get booking by user successfully",
		"bookings": bookings,
	})
}

func DeleteBooking (ctx *gin.Context) {
	// ambil user id yang login
	userId, _ := ctx.Get("userID")

	// ambil parameter id, sesuaikan dengan route kalo "id" tulis "id", kalo "event_id" tulis "event_id"
	paramId := ctx.Param("id")

	// kalo datanya tidak ditemukan
	var bookingData = config.DB.First(&models.Booking{}, paramId).Error

	// jika datanya tidak ditemukan
	if bookingData != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "Booking not found",
		})
		return
	}

	// cek apakah user yang delete sama dengan user yang login
	var booking models.Booking
	config.DB.First(&booking, paramId)
	if booking.UserId != userId.(int) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"message": "You are not authorized to delete this booking",
		})
		return
	}

	// hapus booking
	config.DB.Unscoped().Delete(&models.Booking{}, paramId)

	// response berhasil
	ctx.JSON(http.StatusOK, gin.H{
		"message": "Booking deleted successfully",
	})
}