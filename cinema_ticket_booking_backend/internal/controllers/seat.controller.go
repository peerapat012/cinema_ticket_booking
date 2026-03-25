package controllers

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/peerapat012/CinemaTicketBooking/internal/models"
	"github.com/peerapat012/CinemaTicketBooking/internal/services"
	"github.com/peerapat012/CinemaTicketBooking/internal/ws"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type seatLockResult struct {
	seatNo  string
	locked  bool
	success bool
	err     error
}

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

func LockSeats() gin.HandlerFunc {
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

		var req struct {
			MovieID string   `json:"movieId" binding:"required"`
			Seats   []string `json:"seats" binding:"required,min=1"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload."})
			return
		}

		userIDStr := objUserID.Hex()

		results := make(chan seatLockResult, len(req.Seats))
		var wg sync.WaitGroup

		for _, seatNo := range req.Seats {
			wg.Add(1)
			go func(movieID, seatNo, userID string) {
				defer wg.Done()
				locked, err := services.SeatLock.LockSeat(ctx, movieID, seatNo, userID)
				results <- seatLockResult{
					seatNo:  seatNo,
					locked:  locked,
					success: err == nil,
					err:     err,
				}
			}(req.MovieID, seatNo, userIDStr)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		lockedSeats := []string{}
		failedSeats := []string{}

		for result := range results {
			if !result.success {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lock seat: " + result.seatNo})
				return
			}
			if result.locked {
				lockedSeats = append(lockedSeats, result.seatNo)
			} else {
				failedSeats = append(failedSeats, result.seatNo)
			}
		}

		if len(failedSeats) > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"message":     "Some seats could not be locked.",
				"lockedSeats": lockedSeats,
				"failedSeats": failedSeats,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "Seats locked successfully.",
			"lockedSeats": lockedSeats,
		})
	}
}

func UnlockSeats() gin.HandlerFunc {
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

		var req struct {
			MovieID string   `json:"movieId" binding:"required"`
			Seats   []string `json:"seats" binding:"required,min=1"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload."})
			return
		}

		userIDStr := objUserID.Hex()

		var wg sync.WaitGroup
		errCh := make(chan error, len(req.Seats))

		for _, seatNo := range req.Seats {
			wg.Add(1)
			go func(movieID, seatNo, userID string) {
				defer wg.Done()
				if err := services.SeatLock.UnlockSeat(ctx, movieID, seatNo, userID); err != nil {
					errCh <- err
				}
			}(req.MovieID, seatNo, userIDStr)
		}

		wg.Wait()
		close(errCh)

		for err := range errCh {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlock seat: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Seats unlocked successfully."})
	}
}

func CheckSeatLock() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		movieID := c.Param("id")
		seatNo := c.Query("seatNo")

		if movieID == "" || seatNo == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie ID and seatNo are required."})
			return
		}

		locked, owner, err := services.SeatLock.IsLocked(ctx, movieID, seatNo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check seat lock."})
			return
		}

		if !locked {
			c.JSON(http.StatusOK, gin.H{"status": "available"})
			return
		}

		ttl, _ := services.SeatLock.GetLockTTL(ctx, movieID, seatNo)

		c.JSON(http.StatusOK, gin.H{
			"status":   "locked",
			"owner":    owner,
			"lockedAt": ttl.Seconds(),
		})
	}
}
