package config

import (
	"log"
	"os"

	"example.com/event-app/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := os.Getenv("DATABASE_URI")

	// jika database_uri kosong
	if dsn == " " {
		log.Fatal("Environment variable DATABASE_URI is not set")
	}

	// koneksi ke supabase
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	// jika terjadi error gagal terkoneksi ke supabase
	if err != nil {
		log.Fatal("Failed to connect to database")
	}

	// migrate database, pertabel dimasukan kalo ada tabel baru dimasukan
	err = db.AutoMigrate(&models.Event{}, &models.User{}, &models.Booking{})

	// jika terjadi error migrate database
	if err != nil {
		log.Fatal("Failed to migrate database")
	}

	// set database
	DB = db

	// jika berhasil terkoneksi ke supabase
	log.Println("Connected to database")
}