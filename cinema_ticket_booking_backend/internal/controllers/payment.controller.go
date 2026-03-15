package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/peerapat012/CinemaTicketBooking/internal/database"
	"github.com/peerapat012/CinemaTicketBooking/internal/models"
	"github.com/peerapat012/CinemaTicketBooking/internal/services"
	"github.com/peerapat012/CinemaTicketBooking/internal/ws"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var paymentCollection *mongo.Collection = database.OpenCollection("booked_payment")

type CreatePaymentRequest struct {
	BookingID string `json:"bookingId" binding:"required"`
}

func CreatePayment() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated."})
			return
		}

		objUserID, ok := userID.(bson.ObjectID)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID."})
			return
		}

		var req CreatePaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload."})
			return
		}

		bookingObjID, err := bson.ObjectIDFromHex(req.BookingID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID."})
			return
		}

		bookingCollection := database.OpenCollection("bookings")
		var booking models.Booking
		if err := bookingCollection.FindOne(ctx, bson.M{"_id": bookingObjID}).Decode(&booking); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found."})
			return
		}

		if booking.UserID != objUserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to make payment for this booking."})
			return
		}

		if booking.PaymentStatus == models.PaymentStatusCompleted {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payment already completed."})
			return
		}

		if time.Now().After(booking.PaymentDeadline) {
			_, _ = bookingCollection.UpdateOne(ctx, bson.M{"_id": bookingObjID}, bson.M{
				"$set": bson.M{
					"status":        models.BookingStatusExpired,
					"paymentStatus": models.PaymentStatusFailed,
					"updatedAt":     time.Now(),
				},
			})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payment deadline expired. Please create a new booking."})
			return
		}

		movieCollection := database.OpenCollection("movies")
		var movie models.Movie
		if err := movieCollection.FindOne(ctx, bson.M{"_id": booking.MovieID}).Decode(&movie); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found."})
			return
		}

		var paymentSeats []models.PaymentSeat
		for _, seat := range booking.Seats {
			paymentSeats = append(paymentSeats, models.PaymentSeat{
				SeatNo: seat.SeatNo,
				Price:  seat.Price,
			})
		}

		transactionID := "TXN-" + time.Now().Format("20060102150405")

		payment := models.Payment{
			ID:            bson.ObjectID{},
			UserID:        objUserID,
			MovieID:       booking.MovieID,
			MovieTitle:    movie.Title,
			Seats:         paymentSeats,
			Amount:        booking.TotalPrice,
			Status:        models.PaymentStatusCompleted,
			TransactionID: transactionID,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if _, err := paymentCollection.InsertOne(ctx, payment); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment."})
			return
		}

		_, err = bookingCollection.UpdateOne(ctx, bson.M{"_id": bookingObjID}, bson.M{
			"$set": bson.M{
				"status":        models.BookingStatusConfirmed,
				"paymentStatus": models.PaymentStatusCompleted,
				"updatedAt":     time.Now(),
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update booking status."})
			return
		}

		movieID := booking.MovieID.Hex()
		for _, seat := range booking.Seats {
			ws.GlobalHub.BookSeat(movieID, seat.SeatNo, objUserID.Hex())
			seat := &ws.SeatStatus{
				MovieID: movieID,
				SeatNo:  seat.SeatNo,
				Status:  "booked",
				UserID:  objUserID.Hex(),
			}
			ws.GlobalHub.BroadcastSeatUpdate(movieID, seat)
		}

		booking.Status = models.BookingStatusConfirmed
		booking.PaymentStatus = models.PaymentStatusCompleted

		var seatNos []string
		for _, seat := range booking.Seats {
			seatNos = append(seatNos, seat.SeatNo)
		}

		services.RedisPub.Publish(services.ChannelBookingPaid, services.BookingEvent{
			BookingID:  booking.ID.Hex(),
			UserID:     booking.UserID.Hex(),
			MovieID:    booking.MovieID.Hex(),
			MovieTitle: movie.Title,
			Seats:      seatNos,
			Amount:     booking.TotalPrice,
			Timestamp:  time.Now(),
		})

		c.JSON(http.StatusOK, gin.H{"message": "Payment created successfully.", "data": map[string]interface{}{
			"payment": payment,
			"booking": booking,
		}})
	}
}

func GetMyPayments() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated."})
			return
		}

		objUserID, ok := userID.(bson.ObjectID)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID."})
			return
		}

		cursor, err := paymentCollection.Find(ctx, bson.M{"userId": objUserID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments."})
			return
		}
		defer cursor.Close(ctx)

		var payments []models.Payment
		if err := cursor.All(ctx, &payments); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode payments."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Payments fetched successfully.", "data": payments})
	}
}
