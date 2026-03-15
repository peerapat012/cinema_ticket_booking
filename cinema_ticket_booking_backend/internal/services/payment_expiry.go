package services

import (
	"context"
	"log"
	"time"

	"github.com/peerapat012/CinemaTicketBooking/internal/database"
	"github.com/peerapat012/CinemaTicketBooking/internal/models"
	"github.com/peerapat012/CinemaTicketBooking/internal/ws"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var bookingCollection *mongo.Collection = database.OpenCollection("bookings")

type PaymentExpiryService struct {
	ticker *time.Ticker
	done   chan bool
}

func NewPaymentExpiryService() *PaymentExpiryService {
	return &PaymentExpiryService{
		ticker: time.NewTicker(30 * time.Second),
		done:   make(chan bool),
	}
}

func (s *PaymentExpiryService) Start() {
	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.expirePendingBookings()
			case <-s.done:
				s.ticker.Stop()
				return
			}
		}
	}()
	log.Println("Payment expiry service started")
}

func (s *PaymentExpiryService) Stop() {
	s.done <- true
}

func (s *PaymentExpiryService) expirePendingBookings() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := bookingCollection.Find(ctx, bson.M{
		"status":          models.BookingStatusPending,
		"paymentStatus":   models.PaymentStatusPending,
		"paymentDeadline": bson.M{"$lt": time.Now()},
	})
	if err != nil {
		log.Printf("Error finding expired bookings: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var expiredBookings []models.Booking
	if err := cursor.All(ctx, &expiredBookings); err != nil {
		log.Printf("Error decoding expired bookings: %v", err)
		return
	}

	for _, booking := range expiredBookings {
		movieID := booking.MovieID.Hex()
		userID := booking.UserID.Hex()

		for _, seat := range booking.Seats {
			SeatLock.UnlockSeat(ctx, movieID, seat.SeatNo, userID)
			ws.GlobalHub.ReleaseSeat(movieID, seat.SeatNo, userID)
			seatStatus := &ws.SeatStatus{
				MovieID: movieID,
				SeatNo:  seat.SeatNo,
				Status:  "available",
			}
			ws.GlobalHub.BroadcastSeatUpdate(movieID, seatStatus)
		}

		_, err := bookingCollection.UpdateOne(ctx, bson.M{"_id": booking.ID}, bson.M{
			"$set": bson.M{
				"status":        models.BookingStatusExpired,
				"paymentStatus": models.PaymentStatusFailed,
				"updatedAt":     time.Now(),
			},
		})
		if err != nil {
			log.Printf("Error updating expired booking %s: %v", booking.BookingCode, err)
		} else {
			var seatNos []string
			for _, seat := range booking.Seats {
				seatNos = append(seatNos, seat.SeatNo)
			}

			movieCollection := database.OpenCollection("movies")
			var movie models.Movie
			movieTitle := ""
			_ = movieCollection.FindOne(ctx, bson.M{"_id": booking.MovieID}).Decode(&movie)
			if movie.Title != "" {
				movieTitle = movie.Title
			}

			RedisPub.Publish(ChannelBookingExpired, BookingEvent{
				BookingID:  booking.ID.Hex(),
				UserID:     booking.UserID.Hex(),
				MovieID:    booking.MovieID.Hex(),
				MovieTitle: movieTitle,
				Seats:      seatNos,
				Amount:     booking.TotalPrice,
				Timestamp:  time.Now(),
			})

			log.Printf("Booking %s expired and seats unlocked", booking.BookingCode)
		}
	}
}

var PaymentExpiry *PaymentExpiryService

func init() {
	PaymentExpiry = NewPaymentExpiryService()
}
