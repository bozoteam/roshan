package models

import (
	"encoding/json"
	"sync"

	"github.com/bozoteam/roshan/internal/modules/user/models"
	userModel "github.com/bozoteam/roshan/internal/modules/user/models"
	"github.com/gorilla/websocket"
)

type Message struct {
	RoomID    string `json:"room_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

type Event struct {
	RoomID    string `json:"room_id"`
	Users     []*userModel.User
	Timestamp int64 `json:"timestamp"`
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
	Rooms            map[string]*Room
	Register         chan *Room
	Unregister       chan *Room
	BroadcastMessage chan *Message
	BrodcastEvent    chan *Event
	Mutex            sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Rooms:            make(map[string]*Room),
		Register:         make(chan *Room),
		Unregister:       make(chan *Room),
		BroadcastMessage: make(chan *Message),
		BrodcastEvent:    make(chan *Event),
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
		case event := <-h.BrodcastEvent:
			h.Mutex.Lock()
			if room, ok := h.Rooms[event.RoomID]; ok {
				room.Mutex.Lock()
				eventJSON, _ := json.Marshal(event)
				for client := range room.Clients {
					err := client.Conn.WriteMessage(websocket.TextMessage, eventJSON)
					if err != nil {
						client.Conn.Close()
						delete(room.Clients, client)
					}
				}
				room.Mutex.Unlock()
			}
			h.Mutex.Unlock()
		case message := <-h.BroadcastMessage:
			h.Mutex.Lock()
			if room, ok := h.Rooms[message.RoomID]; ok {
				room.Mutex.Lock()
				messageJSON, _ := json.Marshal(message)
				for client := range room.Clients {
					err := client.Conn.WriteMessage(websocket.TextMessage, messageJSON)
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
