package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/peerapat012/CinemaTicketBooking/internal/database"
	"github.com/peerapat012/CinemaTicketBooking/internal/models"
	"github.com/peerapat012/CinemaTicketBooking/internal/ws"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var showtimeCollection *mongo.Collection = database.OpenCollection("showtimes")

func GetSeatsByMovieID() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		movieID := c.Params.ByName("id")

		if movieID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie ID is required."})
			return
		}

		objID, err := bson.ObjectIDFromHex(movieID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID."})
			return
		}

		cursor, err := bookingCollection.Find(ctx, bson.M{
			"movieId": objID,
			"status":  models.BookingStatusConfirmed,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings."})
			return
		}
		defer cursor.Close(ctx)

		var bookedSeats []string
		for cursor.Next(ctx) {
			var booking models.Booking
			if err := cursor.Decode(&booking); err != nil {
				continue
			}
			for _, seat := range booking.Seats {
				bookedSeats = append(bookedSeats, seat.SeatNo)
			}
		}

		var lockedSeats []string
		wsSeats := ws.GlobalHub.GetSeats(movieID)
		for _, seat := range wsSeats {
			if seat.Status == "locked" {
				lockedSeats = append(lockedSeats, seat.SeatNo)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "Seats fetched successfully.",
			"movieId":     movieID,
			"bookedSeats": bookedSeats,
			"lockedSeats": lockedSeats,
		})
	}
}
