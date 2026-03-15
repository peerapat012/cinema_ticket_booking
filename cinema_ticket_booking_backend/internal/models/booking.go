package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusExpired   BookingStatus = "expired"
)

type BookingSeat struct {
	SeatNo string  `bson:"seatNo" json:"seatNo"`
	Price  float64 `bson:"price" json:"price"`
}

type Booking struct {
	ID              bson.ObjectID `bson:"_id,omitempty" json:"id"`
	BookingCode     string        `bson:"bookingCode" json:"bookingCode"`
	UserID          bson.ObjectID `bson:"userId" json:"userId"`
	MovieID         bson.ObjectID `bson:"movieId" json:"movieId"`
	Seats           []BookingSeat `bson:"seats" json:"seats"`
	TotalPrice      float64       `bson:"totalPrice" json:"totalPrice"`
	Status          BookingStatus `bson:"status" json:"status"`
	PaymentStatus   PaymentStatus `bson:"paymentStatus" json:"paymentStatus"`
	PaymentDeadline time.Time     `bson:"paymentDeadline" json:"paymentDeadline"`
	BookedAt        time.Time     `bson:"bookedAt" json:"bookedAt"`
	CreatedAt       time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time     `bson:"updatedAt" json:"updatedAt"`
}
