package controllers

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/peerapat012/CinemaTicketBooking/internal/database"
	"github.com/peerapat012/CinemaTicketBooking/internal/models"
	"github.com/peerapat012/CinemaTicketBooking/internal/services"
	"github.com/peerapat012/CinemaTicketBooking/internal/ws"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var bookingCollection *mongo.Collection = database.OpenCollection("bookings")

func GetMyBookings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated."})
			return
		}

		objID, ok := userID.(bson.ObjectID)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID."})
			return
		}

		cursor, err := bookingCollection.Find(ctx, bson.M{"userId": objID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings."})
			return
		}
		defer cursor.Close(ctx)

		var bookings []models.Booking
		if err := cursor.All(ctx, &bookings); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode bookings."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Bookings fetched successfully.", "data": bookings})
	}
}

type CreateBookingRequest struct {
	MovieID string               `json:"movieId" binding:"required"`
	Seats   []models.BookingSeat `json:"seats" binding:"required,min=1"`
}

func CreateBooking() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated."})
			return
		}

		var req CreateBookingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload."})
			return
		}

		objUserID, ok := userID.(bson.ObjectID)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID."})
			return
		}

		movieObjID, err := bson.ObjectIDFromHex(req.MovieID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID."})
			return
		}

		movieCollection := database.OpenCollection("movies")
		var movie models.Movie
		if err := movieCollection.FindOne(ctx, bson.M{"_id": movieObjID}).Decode(&movie); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found."})
			return
		}

		existingBookings, err := bookingCollection.Find(ctx, bson.M{
			"movieId": movieObjID,
			"status":  models.BookingStatusConfirmed,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing bookings."})
			return
		}
		defer existingBookings.Close(ctx)

		bookedSeats := make(map[string]bool)
		for existingBookings.Next(ctx) {
			var booking models.Booking
			if err := existingBookings.Decode(&booking); err != nil {
				continue
			}
			for _, seat := range booking.Seats {
				bookedSeats[seat.SeatNo] = true
			}
		}

		var validSeats []models.BookingSeat
		var totalPrice float64
		var mu sync.Mutex

		userIDStr := objUserID.Hex()
		lockKey := req.MovieID

		results := make(chan seatLockResult, len(req.Seats))
		var wg sync.WaitGroup

		for _, seat := range req.Seats {
			wg.Add(1)
			go func(seatNo, userID string) {
				defer wg.Done()
				locked, err := services.SeatLock.LockSeat(ctx, lockKey, seatNo, userID)
				results <- seatLockResult{
					seatNo:  seatNo,
					locked:  locked,
					success: err == nil,
					err:     err,
				}
			}(seat.SeatNo, userIDStr)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		seatLocks := make(map[string]bool)
		for result := range results {
			if !result.success {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lock seat: " + result.seatNo})
				return
			}
			if !result.locked {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Seat " + result.seatNo + " is already locked by another user."})
				return
			}
			seatLocks[result.seatNo] = true
		}

		for _, seat := range req.Seats {
			if bookedSeats[seat.SeatNo] {
				for lockedSeat := range seatLocks {
					services.SeatLock.UnlockSeat(ctx, lockKey, lockedSeat, userIDStr)
				}
				c.JSON(http.StatusBadRequest, gin.H{"error": "Seat " + seat.SeatNo + " is already booked."})
				return
			}
			seatPrice := movie.Price
			mu.Lock()
			validSeats = append(validSeats, models.BookingSeat{
				SeatNo: seat.SeatNo,
				Price:  seatPrice,
			})
			totalPrice += seatPrice
			mu.Unlock()
		}

		booking := models.Booking{
			ID:              bson.ObjectID{},
			BookingCode:     generateBookingCode(),
			UserID:          objUserID,
			MovieID:         movieObjID,
			Seats:           validSeats,
			TotalPrice:      totalPrice,
			Status:          models.BookingStatusPending,
			PaymentStatus:   models.PaymentStatusPending,
			PaymentDeadline: time.Now().Add(PaymentTimeout),
			BookedAt:        time.Now(),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		if result, err := bookingCollection.InsertOne(ctx, booking); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking."})
			return
		} else {
			booking.ID = result.InsertedID.(bson.ObjectID)
		}

		var seatNos []string
		for _, seat := range booking.Seats {
			seatNos = append(seatNos, seat.SeatNo)
		}

		services.SeatLock.ExtendLock(ctx, lockKey, seatNos, userIDStr)

		movieIDStr := booking.MovieID.Hex()
		for _, seat := range booking.Seats {
			ws.GlobalHub.LockSeat(movieIDStr, seat.SeatNo, userIDStr)
			ws.GlobalHub.BroadcastSeatUpdate(movieIDStr, &ws.SeatStatus{
				MovieID: movieIDStr,
				SeatNo:  seat.SeatNo,
				Status:  "locked",
				UserID:  userIDStr,
			})
		}

		services.RedisPub.Publish(services.ChannelBookingCreated, services.BookingEvent{
			BookingID:  booking.ID.Hex(),
			UserID:     booking.UserID.Hex(),
			MovieID:    booking.MovieID.Hex(),
			MovieTitle: movie.Title,
			Seats:      seatNos,
			Amount:     booking.TotalPrice,
			Timestamp:  time.Now(),
		})

		c.JSON(http.StatusOK, gin.H{"message": "Booking created successfully.", "data": booking})
	}
}

const PaymentTimeout = 5 * time.Minute

func generateBookingCode() string {
	return time.Now().Format("20060102150405")
}

type BookingFilter struct {
	MovieID   string `form:"movieId"`
	MovieName string `form:"movieName"`
	UserID    string `form:"userId"`
	Status    string `form:"status"`
	Date      string `form:"date"`
}

func GetAllBookings() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var filter BookingFilter
		if err := c.ShouldBindQuery(&filter); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters."})
			return
		}

		query := bson.M{}

		if filter.MovieID != "" {
			objID, err := bson.ObjectIDFromHex(filter.MovieID)
			if err == nil {
				query["movieId"] = objID
			}
		}

		if filter.UserID != "" {
			objID, err := bson.ObjectIDFromHex(filter.UserID)
			if err == nil {
				query["userId"] = objID
			}
		}

		if filter.Status != "" {
			query["status"] = filter.Status
		}

		if filter.Date != "" {
			startDate, err := time.Parse("2006-01-02", filter.Date)
			if err == nil {
				endDate := startDate.Add(24 * time.Hour)
				query["createdAt"] = bson.M{
					"$gte": startDate,
					"$lt":  endDate,
				}
			}
		}

		cursor, err := bookingCollection.Find(ctx, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings."})
			return
		}
		defer cursor.Close(ctx)

		var bookings []models.Booking
		if err := cursor.All(ctx, &bookings); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode bookings."})
			return
		}

		if filter.MovieName != "" {
			movieCollection := database.OpenCollection("movies")
			movieIDs := make(map[bson.ObjectID]bool)
			for _, b := range bookings {
				movieIDs[b.MovieID] = true
			}

			movieCursor, err := movieCollection.Find(ctx, bson.M{
				"$or": []bson.M{
					{"title": bson.M{"$regex": filter.MovieName, "$options": "i"}},
					{"_id": bson.M{"$in": getKeys(movieIDs)}},
				},
			})
			if err == nil {
				defer movieCursor.Close(ctx)
				var movies []models.Movie
				movieCursor.All(ctx, &movies)

				filteredMovieIDs := make(map[bson.ObjectID]bool)
				for _, m := range movies {
					filteredMovieIDs[m.ID] = true
				}

				var filteredBookings []models.Booking
				for _, b := range bookings {
					if filteredMovieIDs[b.MovieID] {
						filteredBookings = append(filteredBookings, b)
					}
				}
				bookings = filteredBookings
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Bookings fetched successfully.", "data": bookings, "total": len(bookings)})
	}
}

func getKeys(m map[bson.ObjectID]bool) []bson.ObjectID {
	keys := make([]bson.ObjectID, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
