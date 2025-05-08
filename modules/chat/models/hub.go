package models

import (
	"fmt"
	"time"

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
	ID         string
	Name       string
	CreatorID  string
	Clients    map[string]*ws_client.Client
	emptyTimer *time.Timer
}

// Hub manages all rooms and connections
type Hub struct {
	// Rooms indexed by ID
	rooms map[string]*Room

	// Channels for operations
	register         chan *ws_client.ClientRegistration
	unregister       chan *ws_client.ClientUnregistration
	broadcastMessage chan *sendMessage
	broadcastEvent   chan *Event
	createRoom       chan *createRoom
	deleteRoom       chan *deleteRoom
	getRoom          chan *roomRequest
	listRooms        chan *roomsRequest
}

// NewHub creates a new hub
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*Room),

		register:   make(chan *ws_client.ClientRegistration),
		unregister: make(chan *ws_client.ClientUnregistration),

		broadcastMessage: make(chan *sendMessage),
		broadcastEvent:   make(chan *Event),

		createRoom: make(chan *createRoom),
		deleteRoom: make(chan *deleteRoom),
		getRoom:    make(chan *roomRequest),
		listRooms:  make(chan *roomsRequest),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		fmt.Println("ready to consume again")
		select {
		case reg := <-h.register:
			h.handleRegister(reg)

		case unreg := <-h.unregister:
			h.handleUnregister(unreg)

		case create := <-h.createRoom:
			h.handleCreateRoom(create)

		case roomID := <-h.deleteRoom:
			h.handleDeleteRoom(roomID)

		case msg := <-h.broadcastMessage:
			h.handleMessage(msg)

		case event := <-h.broadcastEvent:
			h.handleEvent(event)

		case req := <-h.getRoom:
			h.handleGetRoom(req)

		case req := <-h.listRooms:
			h.handleListRooms(req)
		}
	}
}

// sendUserList sends the current user list to all clients in a room
func (h *Hub) sendUserList(room *Room) {
	users := make([]*userModel.User, 0, len(room.Clients))
	for clientId, client := range room.Clients {
		users = append(users, &userModel.User{
			Id:    clientId,
			Name:  client.Name,
			Email: client.Email,
		})
	}

	event := &Event{
		RoomID:    room.ID,
		Users:     users,
		Timestamp: time.Now().UnixNano(),
	}

	h.handleEvent(event)
}
