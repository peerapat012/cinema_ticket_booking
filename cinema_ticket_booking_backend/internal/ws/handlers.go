package ws

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var GlobalHub *Hub

func init() {
	GlobalHub = NewHub()
	go GlobalHub.Run()
}

func HandleWebSocket() gin.HandlerFunc {
	return func(c *gin.Context) {
		movieID := c.Query("movieId")
		userID := c.Query("userId")

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		client := &Client{
			hub:    GlobalHub,
			conn:   conn,
			send:   make(chan []byte, 256),
			movie:  movieID,
			userID: userID,
		}

		client.hub.register <- client

		go client.writePump()
		go client.readPump()

		if movieID != "" {
			seats := GlobalHub.GetSeats(movieID)
			for _, seat := range seats {
				data, _ := json.Marshal(seat)
				client.send <- data
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "lock_seat":
			success := GlobalHub.LockSeat(msg.Movie, msg.Seat, msg.UserID)
			if success {
				seat := &SeatStatus{
					MovieID: msg.Movie,
					SeatNo:  msg.Seat,
					Status:  "locked",
					UserID:  msg.UserID,
				}
				GlobalHub.BroadcastSeatUpdate(msg.Movie, seat)
			}

		case "book_seat":
			success := GlobalHub.BookSeat(msg.Movie, msg.Seat, msg.UserID)
			if success {
				seat := &SeatStatus{
					MovieID: msg.Movie,
					SeatNo:  msg.Seat,
					Status:  "booked",
					UserID:  msg.UserID,
				}
				GlobalHub.BroadcastSeatUpdate(msg.Movie, seat)
			}

		case "release_seat":
			GlobalHub.ReleaseSeat(msg.Movie, msg.Seat, msg.UserID)
			seat := &SeatStatus{
				MovieID: msg.Movie,
				SeatNo:  msg.Seat,
				Status:  "available",
			}
			GlobalHub.BroadcastSeatUpdate(msg.Movie, seat)
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		message, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}
