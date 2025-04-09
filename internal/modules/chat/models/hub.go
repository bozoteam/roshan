package models

import (
	"sync"

	"github.com/bozoteam/roshan/internal/modules/user/models"
	"github.com/gorilla/websocket"
)

type Message struct {
	RoomID  string `json:"room_id"`
	UserId  string `json:"user_id"`
	Content string `json:"content"`
}

type RoomClient struct {
	*models.User
	Conn *websocket.Conn
}

type Room struct {
	ID        string
	Name      string
	CreatorId string
	Clients   map[*RoomClient]bool
	Mutex     *sync.Mutex
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
				for client, _ := range room.Clients {
					err := client.Conn.WriteMessage(websocket.TextMessage, []byte(message.Content))
					if err != nil {
						client.Conn.Close()
						delete(room.Clients, client)
					}
				}
				room.Mutex.Unlock()
			}
			h.Mutex.Unlock()
		}
	}
}
