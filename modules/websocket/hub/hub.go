package ws_hub

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/bozoteam/roshan/modules/user/models"
	userModel "github.com/bozoteam/roshan/modules/user/models"
)

// Event represents a room event (like user list updates)
type Event struct {
	RoomID    string            `json:"room_id"`
	Users     []*userModel.User `json:"users"`
	Timestamp int64             `json:"timestamp"`
}

// ClientI defines the interface for clients
type ClientI interface {
	GetID() string
	GetSender() chan []byte
	GetUser() *userModel.User
	WaitUnregister()
}

// RoomI defines the interface for rooms
type RoomI interface {
	GetID() string
	GetClients() map[string]ClientI
	SetSomeoneEntered(bool)
	GetSomeoneEntered() bool
	UserIsInRoom(userId string) bool

	Clone() RoomI
}

// Hub manages all rooms and connections
type Hub struct {
	rooms map[string]RoomI
	mu    *sync.RWMutex
}

// NewHub creates a new hub
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]RoomI, 1024),
		mu:    new(sync.RWMutex),
	}
}

func (h *Hub) GetRoom(roomId string) RoomI {
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

func (h *Hub) CreateRoom(room RoomI) {
	h.mu.Lock()
	h.rooms[room.GetID()] = room
	h.mu.Unlock()

	go func() {
		<-time.NewTimer(time.Second * 5).C
		h.mu.RLock()
		entered := room.GetSomeoneEntered()
		id := room.GetID()
		h.mu.RUnlock()
		if !entered {
			h.DeleteRoom(id)
		}
	}()
}

func (h *Hub) ListRooms() []RoomI {
	rooms := make([]RoomI, 0, len(h.rooms))
	h.mu.RLock()
	for _, room := range h.rooms {
		rooms = append(rooms, room.Clone())
	}
	h.mu.RUnlock()
	return rooms
}

func (h *Hub) Register(client ClientI, roomId string) {
	h.mu.Lock()
	room, exists := h.rooms[roomId]
	if !exists {
		h.mu.Unlock()
		return
	}

	room.SetSomeoneEntered(true)
	clients := room.GetClients()
	clients[client.GetID()] = client
	h.mu.Unlock()

	h.sendUserList(roomId)
}

func (h *Hub) Unregister(client ClientI, roomId string) {
	h.mu.Lock()
	room, exists := h.rooms[roomId]
	if !exists {
		h.mu.Unlock()
		return
	}

	clients := room.GetClients()
	delete(clients, client.GetID())
	if len(clients) == 0 {
		delete(h.rooms, roomId)
	}
	h.mu.Unlock()

	h.sendUserList(roomId)
}

// BroadcastBytes sends data to all clients in a room
func (h *Hub) BroadcastBytes(roomId string, data []byte) {
	h.mu.RLock()
	room, exists := h.rooms[roomId]
	if !exists {
		h.mu.RUnlock()
		return
	}

	clients := room.GetClients()
	chans := make([]chan []byte, 0, len(clients))
	for _, client := range clients {
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

// SendBytes sends data to a specific client in a room
func (h *Hub) SendBytes(roomId string, clientId string, data []byte) bool {
	h.mu.RLock()
	room, exists := h.rooms[roomId]
	if !exists {
		h.mu.RUnlock()
		return false
	}

	clients := room.GetClients()
	client, exists := clients[clientId]
	if !exists {
		h.mu.RUnlock()
		return false
	}

	sender := client.GetSender()
	h.mu.RUnlock()

	select {
	case sender <- data:
		return true
	default:
		return false
	}
}

// SendBytesToClients sends different data to specific clients in a room
func (h *Hub) SendBytesToClients(roomId string, clientData map[string][]byte) {
	h.mu.RLock()
	room, exists := h.rooms[roomId]
	if !exists {
		h.mu.RUnlock()
		return
	}

	clients := room.GetClients()

	// Create a map of client senders
	senders := make(map[string]chan []byte, len(clientData))
	for clientId := range clientData {
		if client, ok := clients[clientId]; ok {
			senders[clientId] = client.GetSender()
		}
	}
	h.mu.RUnlock()

	// Send data to each client
	for clientId, data := range clientData {
		if sender, ok := senders[clientId]; ok {
			select {
			case sender <- data:
			default:
			}
		}
	}
}

// BroadcastBytesExcept sends data to all clients in a room except the specified ones
func (h *Hub) BroadcastBytesExcept(roomId string, excludedClientIds []string, data []byte) {
	h.mu.RLock()
	room, exists := h.rooms[roomId]
	if !exists {
		h.mu.RUnlock()
		return
	}

	clients := room.GetClients()

	// Create exclude map for O(1) lookups
	excludeMap := make(map[string]struct{}, len(excludedClientIds))
	for _, id := range excludedClientIds {
		excludeMap[id] = struct{}{}
	}

	// Collect senders for clients that aren't excluded
	chans := make([]chan []byte, 0, len(clients))
	for clientId, client := range clients {
		if _, excluded := excludeMap[clientId]; !excluded {
			chans = append(chans, client.GetSender())
		}
	}
	h.mu.RUnlock()

	// Send data
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

	clients := room.GetClients()
	users := make([]*models.User, 0, len(clients))
	for _, client := range clients {
		users = append(users, client.GetUser().Clone())
	}
	roomID := room.GetID()
	h.mu.RUnlock()

	data, err := json.Marshal(&Event{
		RoomID:    roomID,
		Users:     users,
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		return
	}

	go h.BroadcastBytes(roomID, data)
}
