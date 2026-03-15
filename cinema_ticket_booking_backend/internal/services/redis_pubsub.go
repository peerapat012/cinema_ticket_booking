package services

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/peerapat012/CinemaTicketBooking/internal/database"
	"github.com/peerapat012/CinemaTicketBooking/internal/models"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	ChannelBookingCreated = "booking:created"
	ChannelBookingPaid    = "booking:paid"
	ChannelBookingExpired = "booking:expired"
)

var auditLogCollection *mongo.Collection = database.OpenCollection("audit_logs")

type BookingEvent struct {
	BookingID  string    `json:"bookingId"`
	UserID     string    `json:"userId"`
	MovieID    string    `json:"movieId"`
	MovieTitle string    `json:"movieTitle"`
	Seats      []string  `json:"seats"`
	Amount     float64   `json:"amount"`
	Timestamp  time.Time `json:"timestamp"`
}

type RedisPubSub struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisPubSub() *RedisPubSub {
	client := database.GetRedis()
	if client == nil {
		return &RedisPubSub{
			client: nil,
			ctx:    context.Background(),
		}
	}
	return &RedisPubSub{
		client: client,
		ctx:    context.Background(),
	}
}

func (r *RedisPubSub) Publish(channel string, event BookingEvent) error {
	if r.client == nil {
		log.Println("Redis not connected, skipping publish to:", channel)
		return nil
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return r.client.Publish(r.ctx, channel, data).Err()
}

func (r *RedisPubSub) Subscribe(channel string, handler func(BookingEvent)) {
	if r.client == nil {
		log.Println("Redis not connected, skipping subscribe to:", channel)
		return
	}
	pubsub := r.client.Subscribe(r.ctx, channel)
	ch := pubsub.Channel()

	go func() {
		for msg := range ch {
			var event BookingEvent
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("Error unmarshaling event: %v", err)
				continue
			}
			handler(event)
		}
	}()
}

func (r *RedisPubSub) StartSubscriber() {
	r.Subscribe(ChannelBookingCreated, handleBookingCreated)
	r.Subscribe(ChannelBookingPaid, handleBookingPaid)
	r.Subscribe(ChannelBookingExpired, handleBookingExpired)
	log.Println("Redis pub/sub subscribers started")
}

func handleBookingCreated(event BookingEvent) {
	bookingID, _ := bson.ObjectIDFromHex(event.BookingID)
	userID, _ := bson.ObjectIDFromHex(event.UserID)
	movieID, _ := bson.ObjectIDFromHex(event.MovieID)

	auditLog := models.AuditLog{
		ID:          bson.ObjectID{},
		UserID:      userID,
		BookingID:   bookingID,
		MovieID:     movieID,
		MovieTitle:  event.MovieTitle,
		Seats:       event.Seats,
		Amount:      event.Amount,
		Action:      models.AuditActionBookingCreated,
		Description: "Booking created, waiting for payment",
		CreatedAt:   time.Now(),
	}

	_, err := auditLogCollection.InsertOne(context.Background(), auditLog)
	if err != nil {
		log.Printf("Error creating audit log for booking created: %v", err)
	} else {
		log.Printf("Audit log created: booking %s created", event.BookingID)
	}
}

func handleBookingPaid(event BookingEvent) {
	bookingID, _ := bson.ObjectIDFromHex(event.BookingID)
	userID, _ := bson.ObjectIDFromHex(event.UserID)
	movieID, _ := bson.ObjectIDFromHex(event.MovieID)

	auditLog := models.AuditLog{
		ID:          bson.ObjectID{},
		UserID:      userID,
		BookingID:   bookingID,
		MovieID:     movieID,
		MovieTitle:  event.MovieTitle,
		Seats:       event.Seats,
		Amount:      event.Amount,
		Action:      models.AuditActionBookingPaid,
		Description: "Payment completed successfully",
		CreatedAt:   time.Now(),
	}

	_, err := auditLogCollection.InsertOne(context.Background(), auditLog)
	if err != nil {
		log.Printf("Error creating audit log for booking paid: %v", err)
	} else {
		log.Printf("Audit log created: booking %s paid", event.BookingID)
	}
}

func handleBookingExpired(event BookingEvent) {
	bookingID, _ := bson.ObjectIDFromHex(event.BookingID)
	userID, _ := bson.ObjectIDFromHex(event.UserID)
	movieID, _ := bson.ObjectIDFromHex(event.MovieID)

	auditLog := models.AuditLog{
		ID:          bson.ObjectID{},
		UserID:      userID,
		BookingID:   bookingID,
		MovieID:     movieID,
		MovieTitle:  event.MovieTitle,
		Seats:       event.Seats,
		Amount:      event.Amount,
		Action:      models.AuditActionBookingExpired,
		Description: "Payment deadline expired",
		CreatedAt:   time.Now(),
	}

	_, err := auditLogCollection.InsertOne(context.Background(), auditLog)
	if err != nil {
		log.Printf("Error creating audit log for booking expired: %v", err)
	} else {
		log.Printf("Audit log created: booking %s expired", event.BookingID)
	}
}

var RedisPub *RedisPubSub
