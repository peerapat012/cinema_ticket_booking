package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
)

type PaymentSeat struct {
	SeatNo string  `bson:"seatNo" json:"seatNo"`
	Price  float64 `bson:"price" json:"price"`
}

type Payment struct {
	ID            bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        bson.ObjectID `bson:"userId" json:"userId"`
	MovieID       bson.ObjectID `bson:"movieId" json:"movieId"`
	MovieTitle    string        `bson:"movieTitle" json:"movieTitle"`
	Seats         []PaymentSeat `bson:"seats" json:"seats"`
	Amount        float64       `bson:"amount" json:"amount"`
	Status        PaymentStatus `bson:"status" json:"status"`
	TransactionID string        `bson:"transactionId" json:"transactionId"`
	CreatedAt     time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time     `bson:"updatedAt" json:"updatedAt"`
}
