package main

import (
	"log"

	"example.com/event-app/config"
	"example.com/event-app/controllers"
	"example.com/event-app/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// load file .env
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// koneksi ke database
	config.ConnectDB()

	router := gin.Default()

	// Route group ke dalam endpoint "api"
	api := router.Group("/api")
	{
		// events
		api.GET("/events", controllers.GetEvents)
		api.GET("/events/:id", controllers.GetEventById)

		// auth
		api.POST("/auth/register", controllers.RegisterUser)
		api.POST("/auth/login", controllers.LoginUser)

		protected := api.Group("/")
		protected.Use(middlewares.RequiredAuth())

		{
			// protected auth
			protected.GET("/auth/me", controllers.GetCurrentUser)
			protected.PUT("/auth/me", controllers.UpdateCurrentUser)
			protected.PUT("/auth/me/password", controllers.ChangePassword)

			// protected events
			protected.GET("/events/user", controllers.GetEventsByUser)
			protected.POST("/events", controllers.CreateEvent)
			protected.PUT("/events/:id", controllers.UpdateEvent)
			protected.DELETE("/events/:id", controllers.DeleteEvent)

			// protected booking
			protected.POST("/booking", controllers.CreateBookingEvent)
			protected.GET("/booking/user", controllers.GetBookingByUser)
			protected.DELETE("/booking/:id", controllers.DeleteBooking)
		}
	}

	// Jalankan server dan port
	router.Run(":8080")
}