package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	controllers "github.com/peerapat012/CinemaTicketBooking/internal/controllers"
	"github.com/peerapat012/CinemaTicketBooking/internal/middleware"
	"github.com/peerapat012/CinemaTicketBooking/internal/ws"
)

func SetupRoutes(router *gin.Engine) {
	router.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})

	authRoutes(router)
	userRoutes(router)
	movieRoutes(router)
	bookingRoutes(router)
	paymentRoutes(router)
	seatRoutes(router)

	router.GET("/ws", ws.HandleWebSocket())
}

func authRoutes(router *gin.Engine) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", controllers.LoginWithGoogle())
	}
}

func userRoutes(router *gin.Engine) {
	users := router.Group("/users")
	users.Use(middleware.AuthMiddleware())
	{
		users.GET("/me", controllers.GetCurrentUser())
	}

	adminUsers := router.Group("/users")
	adminUsers.Use(middleware.AuthMiddleware(), middleware.RequireAdmin())
	{
		adminUsers.GET("", controllers.GetAllUsers())
		adminUsers.PUT("/:id/role", controllers.UpdateUserRole())
	}
}

func movieRoutes(router *gin.Engine) {
	movies := router.Group("/movies")
	{
		movies.GET("", controllers.GetMovies())
		movies.GET("/:id/seats", controllers.GetSeatsByMovieID())
		movies.GET("/:id/seat-lock", controllers.CheckSeatLock())

		movies.POST("", middleware.AuthMiddleware(), middleware.RequireAdmin(), controllers.CreateMovie())
		movies.PUT("/:id", middleware.AuthMiddleware(), middleware.RequireAdmin(), controllers.UpdateMovie())
		movies.DELETE("/:id", middleware.AuthMiddleware(), middleware.RequireAdmin(), controllers.DeleteMovie())
	}
}

func bookingRoutes(router *gin.Engine) {
	bookings := router.Group("/bookings")
	{
		bookings.GET("", middleware.AuthMiddleware(), middleware.RequireAdmin(), controllers.GetAllBookings())

		userBookings := bookings.Group("")
		userBookings.Use(middleware.AuthMiddleware(), middleware.RequireUser())
		{
			userBookings.GET("/me", controllers.GetMyBookings())
			userBookings.POST("", controllers.CreateBooking())
		}
	}
}

func seatRoutes(router *gin.Engine) {
	seats := router.Group("/seats")
	seats.Use(middleware.AuthMiddleware(), middleware.RequireUser())
	{
		seats.POST("/lock", controllers.LockSeats())
		seats.POST("/unlock", controllers.UnlockSeats())
	}
}

func paymentRoutes(router *gin.Engine) {
	payments := router.Group("/payments")
	payments.Use(middleware.AuthMiddleware(), middleware.RequireUser())
	{
		payments.POST("", controllers.CreatePayment())
		payments.GET("/me", controllers.GetMyPayments())
	}
}
