package ws

import (
	"encoding/json"
	"time"
)

type SeatStatus struct {
	MovieID  string    `json:"movieId"`
	SeatNo   string    `json:"seatNo"`
	Status   string    `json:"status"`
	UserID   string    `json:"userId,omitempty"`
	LockedAt time.Time `json:"lockedAt,omitempty"`
}

type Message struct {
	Type   string          `json:"type"`
	Movie  string          `json:"movieId"`
	Seat   string          `json:"seatNo"`
	Status string          `json:"status"`
	UserID string          `json:"userId"`
	Data   json.RawMessage `json:"data,omitempty"`
}
