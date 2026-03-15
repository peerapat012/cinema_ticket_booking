package ws

import (
	"encoding/json"
	"sync"
)

type Hub struct {
	clients    map[*Client]bool
	seats      map[string]map[string]*SeatStatus
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		seats:      make(map[string]map[string]*SeatStatus),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if client.movie == message.Movie || message.Movie == "" {
					select {
					case client.send <- message.Data:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) LockSeat(movieID, seatNo, userID string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := movieID
	if h.seats[key] == nil {
		h.seats[key] = make(map[string]*SeatStatus)
	}

	seat := h.seats[key][seatNo]
	if seat != nil && seat.Status == "locked" && seat.UserID != userID {
		return false
	}

	h.seats[key][seatNo] = &SeatStatus{
		MovieID: movieID,
		SeatNo:  seatNo,
		Status:  "locked",
		UserID:  userID,
	}
	return true
}

func (h *Hub) BookSeat(movieID, seatNo, userID string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := movieID
	if h.seats[key] == nil {
		h.seats[key] = make(map[string]*SeatStatus)
	}

	seat := h.seats[key][seatNo]
	if seat != nil && seat.Status == "locked" && seat.UserID != userID {
		return false
	}

	h.seats[key][seatNo] = &SeatStatus{
		MovieID: movieID,
		SeatNo:  seatNo,
		Status:  "booked",
		UserID:  userID,
	}
	return true
}

func (h *Hub) ReleaseSeat(movieID, seatNo, userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := movieID
	if seat := h.seats[key][seatNo]; seat != nil && seat.UserID == userID {
		delete(h.seats[key], seatNo)
	}
}

func (h *Hub) GetSeats(movieID string) []*SeatStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	key := movieID
	var result []*SeatStatus
	for _, seat := range h.seats[key] {
		result = append(result, seat)
	}
	return result
}

func (h *Hub) BroadcastSeatUpdate(movieID string, seat *SeatStatus) {
	data, _ := json.Marshal(seat)
	msg := Message{
		Type:  "seat_update",
		Movie: movieID,
		Data:  data,
	}
	h.broadcast <- &msg
}

func (h *Hub) BroadcastToMovie(movieID string, msg *Message) {
	h.broadcast <- msg
}
