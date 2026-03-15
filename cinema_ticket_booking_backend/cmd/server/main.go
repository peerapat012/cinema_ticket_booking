package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/peerapat012/CinemaTicketBooking/internal/database"
	"github.com/peerapat012/CinemaTicketBooking/internal/routes"
	"github.com/peerapat012/CinemaTicketBooking/internal/services"
)

func main() {
	database.InitRedis()

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	routes.SetupRoutes(router)

	services.RedisPub = services.NewRedisPubSub()
	services.RedisPub.StartSubscriber()
	services.PaymentExpiry.Start()

	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
