package ws_hub

import (
	"encoding/json"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/bozoteam/roshan/helpers"
	userModel "github.com/bozoteam/roshan/modules/user/models"
)

type ClientTeam struct {
	ClientI
	Team string // Team the client belongs to
}

// ClintI defines the interface for clients
type ClientI interface {
	GetID() string
	GetSender() chan []byte
	GetUser() *userModel.User
	WaitUnregister()
}

// RoomI defines the interface for rooms
type RoomI interface {
	GetID() string
	GetClients() map[string]ClientTeam
	SetSomeoneEntered(bool)
	GetSomeoneEntered() bool
	Clone() RoomI
	UserIsInRoom(userId string) bool
	GetClientsFromTeam(team string) []ClientTeam
	GetTeamMapping() map[string][]ClientTeam
	RegisterClient(client ClientI, team string)
	UnregisterClient(id string)
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
	h.mu.RLock()
	defer h.mu.RUnlock()
	return slices.Collect(maps.Values(helpers.Clone(h.rooms)))
}

func (h *Hub) Register(client ClientI, roomId string, team string) {
	h.mu.Lock()
	room, exists := h.rooms[roomId]
	if !exists {
		h.mu.Unlock()
		return
	}

	room.SetSomeoneEntered(true)
	clients := room.GetClients()
	clients[client.GetID()] = ClientTeam{
		ClientI: client,
		Team:    team,
	}

	teamMap := room.GetTeamMapping()
	teamMap[team] = append(teamMap[team], ClientTeam{
		ClientI: client,
		Team:    team,
	})
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

	room.UnregisterClient(client.GetID())
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

func (h *Hub) SendBytesToTeam(roomId string, team string, data []byte) {
	h.mu.RLock()
	room, exists := h.rooms[roomId]
	if !exists {
		h.mu.RUnlock()
		return
	}

	teamClients := room.GetClientsFromTeam(team)
	chans := make([]chan []byte, 0, len(teamClients))
	for _, client := range teamClients {
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

// RoomUserList represents a room event (like user list updates)
type RoomUserList struct {
	RoomID    string                       `json:"room_id"`
	Teams     map[string][]*userModel.User `json:"team_mapping"`
	Timestamp int64                        `json:"timestamp"`
}

func (h *Hub) sendUserList(roomId string) {
	h.mu.RLock()
	room, exists := h.rooms[roomId]
	if !exists {
		h.mu.RUnlock()
		return
	}

	tMap := room.GetTeamMapping()
	output := make(map[string][]*userModel.User, len(tMap))
	for team, clients := range tMap {
		users := make([]*userModel.User, len(clients))
		for i, client := range clients {
			users[i] = client.GetUser()
		}
		output[team] = users
	}

	roomID := room.GetID()
	h.mu.RUnlock()

	data, err := json.Marshal(&RoomUserList{
		RoomID:    roomID,
		Teams:     output,
		Timestamp: time.Now().UnixNano(),
	})
	if err != nil {
		panic(err)
	}

	go h.BroadcastBytes(roomID, data)
}
