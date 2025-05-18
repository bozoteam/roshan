package models

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"sync"
	"time"

	"github.com/bozoteam/roshan/helpers"
	"github.com/bozoteam/roshan/modules/user/models"
	userModel "github.com/bozoteam/roshan/modules/user/models"
	"github.com/bozoteam/roshan/modules/websocket/ws_client"
)

// Message represents a chat message
type Message struct {
	RoomID    string          `json:"room_id"`
	User      *userModel.User `json:"user"`
	Content   string          `json:"content"`
	Timestamp int64           `json:"timestamp"`
}

// Event represents a room event (like user list updates)
type Event struct {
	RoomID    string            `json:"room_id"`
	Users     []*userModel.User `json:"users"`
	Timestamp int64             `json:"timestamp"`
}

// Room represents a chat room
type Room struct {
	ID        string
	Name      string
	CreatorID string
	Clients   map[string]*ws_client.Client

	someoneEntered bool
	// mu         *sync.RWMutex // Add mutex for individual room
}

func (r *Room) Clone() *Room {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)

	err := enc.Encode(r)
	if err != nil {
		panic(err)
	}

	var result Room
	err = dec.Decode(&result)
	if err != nil {
		panic(err)
	}

	return &result
}

func NewRoom(name string, creatorId string) *Room {
	return &Room{
		someoneEntered: false,

		ID:        helpers.GenUUID(),
		Name:      name,
		CreatorID: creatorId,
		Clients:   make(map[string]*ws_client.Client),
	}
}

// Hub manages all rooms and connections
type Hub struct {
	rooms map[string]*Room
	mu    *sync.RWMutex // Changed from pointer to embedded, using RWMutex
}

// NewHub creates a new hub
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*Room, 1024),
		mu:    new(sync.RWMutex),
	}
}

func (h *Hub) GetRoom(roomId string) *Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	room, exists := h.rooms[roomId]
	if !exists {
		return nil
	}

	return room.Clone()
}

func (h *Hub) DeleteRoom(roomId string) {
	h.mu.Lock()
	delete(h.rooms, roomId)
	h.mu.Unlock()
}

func (h *Hub) CreateRoom(room *Room) {
	h.mu.Lock()
	h.rooms[room.ID] = room
	h.mu.Unlock()

	go func() {
		<-time.NewTimer(time.Second * 5).C
		h.mu.RLock()
		entered := room.someoneEntered
		id := room.ID
		h.mu.RUnlock()
		if entered == false {
			h.DeleteRoom(id)
		}
	}()
}

func (h *Hub) ListRooms() []*Room {
	h.mu.RLock()
	rooms := make([]*Room, 0, len(h.rooms))
	for _, room := range h.rooms {
		rooms = append(rooms, room.Clone())
	}
	h.mu.RUnlock()

	return rooms
}

func (h *Hub) Register(client *ws_client.Client, roomId string) {
	h.mu.Lock()
	room, exists := h.rooms[roomId]
	room.someoneEntered = true
	if !exists {
		h.mu.Unlock()
		return
	}
	room.Clients[client.Id] = client
	h.mu.Unlock()

	h.sendUserList(roomId)
}

func (h *Hub) Unregister(client *ws_client.Client, roomId string) {
	h.mu.Lock()
	room, exists := h.rooms[roomId]
	if !exists {
		h.mu.Unlock()
		return
	}
	delete(room.Clients, client.Id)
	if len(room.Clients) == 0 {
		delete(h.rooms, roomId)
	}
	h.mu.Unlock()

	h.sendUserList(roomId)
}

func (h *Hub) BroadcastBytes(roomId string, data []byte) {
	h.mu.RLock()
	room, exists := h.rooms[roomId]
	if !exists {
		h.mu.RUnlock()
		return
	}
	chans := make([](chan []byte), 0, len(room.Clients))
	for _, client := range room.Clients {
		chans = append(chans, client.GetSender())
	}
	h.mu.RUnlock()

	for _, c := range chans {
		select {
		case c <- data:
		default:
		}
	}
}

func (h *Hub) sendUserList(roomId string) {
	h.mu.RLock()
	room, exists := h.rooms[roomId]
	if !exists {
		h.mu.RUnlock()
		return
	}
	clients := make([]*models.User, 0, len(room.Clients))
	for _, client := range room.Clients {
		clients = append(clients, client.User.Clone())
	}
	h.mu.RUnlock()

	data, err := json.Marshal(&Event{
		RoomID:    room.ID,
		Users:     clients,
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return
	}

	go h.BroadcastBytes(room.ID, data)
}
