package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuditAction string

const (
	AuditActionBookingCreated   AuditAction = "booking_created"
	AuditActionBookingPaid      AuditAction = "booking_paid"
	AuditActionBookingExpired   AuditAction = "booking_expired"
	AuditActionBookingCancelled AuditAction = "booking_cancelled"
)

type AuditLog struct {
	ID          bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      bson.ObjectID `bson:"userId" json:"userId"`
	BookingID   bson.ObjectID `bson:"bookingId" json:"bookingId"`
	MovieID     bson.ObjectID `bson:"movieId" json:"movieId"`
	MovieTitle  string        `bson:"movieTitle" json:"movieTitle"`
	Seats       []string      `bson:"seats" json:"seats"`
	Amount      float64       `bson:"amount" json:"amount"`
	Action      AuditAction   `bson:"action" json:"action"`
	Description string        `bson:"description" json:"description"`
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
}
