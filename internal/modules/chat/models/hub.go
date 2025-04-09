package models

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	RoomID  string `json:"roomId"`
	Content string `json:"content"`
}

type User struct {
	ID       string
	Username string
	Conn     *websocket.Conn
}

type Room struct {
	ID    string
	Name  string
	Users map[*websocket.Conn]User
	Mutex sync.Mutex
}

type Hub struct {
	Rooms      map[string]*Room
	Register   chan *Room
	Unregister chan *Room
	Broadcast  chan Message
	Mutex      sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Rooms:      make(map[string]*Room),
		Register:   make(chan *Room),
		Unregister: make(chan *Room),
		Broadcast:  make(chan Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case room := <-h.Register:
			h.Mutex.Lock()
			h.Rooms[room.ID] = room
			h.Mutex.Unlock()
		case room := <-h.Unregister:
			h.Mutex.Lock()
			delete(h.Rooms, room.ID)
			h.Mutex.Unlock()
		case message := <-h.Broadcast:
			h.Mutex.Lock()
			if room, ok := h.Rooms[message.RoomID]; ok {
				room.Mutex.Lock()
				for conn, user := range room.Users {
					err := user.Conn.WriteMessage(websocket.TextMessage, []byte(message.Content))
					if err != nil {
						user.Conn.Close()
						delete(room.Users, conn)
					}
				}
				room.Mutex.Unlock()
			}
			h.Mutex.Unlock()
		}
	}
}
