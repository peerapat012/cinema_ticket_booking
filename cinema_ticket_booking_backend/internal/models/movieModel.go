package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type MovieStatus string

const (
	MovieStatusActive   MovieStatus = "active"
	MovieStatusInactive MovieStatus = "inactive"
)

type Movie struct {
	ID              bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Title           string        `bson:"title" json:"title"`
	Description     string        `bson:"description,omitempty" json:"description,omitempty"`
	DurationMinutes int           `bson:"durationMinutes" json:"durationMinutes"`
	PosterURL       string        `bson:"posterUrl,omitempty" json:"posterUrl,omitempty"`
	Price           float64       `bson:"price" json:"price"`
	Status          MovieStatus   `bson:"status" json:"status"`
	CreatedAt       time.Time     `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time     `bson:"updatedAt" json:"updatedAt"`
}
