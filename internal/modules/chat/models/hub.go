package models

import (
	"fmt"
	"time"

	userModel "github.com/bozoteam/roshan/internal/modules/user/models"
	"github.com/gorilla/websocket"
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

// Client represents a connected user
type Client struct {
	*userModel.User
	Conn       *websocket.Conn `json:"-"`
	Send       chan []byte     `json:"-"`
	RoomID     string          `json:"-"`
	Unregister chan<- *Client  `json:"-"`
	PingNotify chan struct{}   `json:"-"`
}

// Room represents a chat room
type Room struct {
	ID         string
	Name       string
	CreatorID  string
	Clients    map[string]*Client
	emptyTimer *time.Timer
}

// ClientRegistration holds data for registering a client to a room
type ClientRegistration struct {
	Client *Client
	RoomID string
}

// ClientUnregistration holds data for unregistering a client
type ClientUnregistration struct {
	Client *Client
	RoomID string
}

// roomRequest is used to safely get room data
type roomRequest struct {
	id     string
	result chan *Room
}

// roomsRequest is used to get a list of all rooms
type roomsRequest struct {
	result chan []*Room
}

// Hub manages all rooms and connections
type Hub struct {
	// Rooms indexed by ID
	rooms map[string]*Room

	// Channels for operations
	register         chan *ClientRegistration
	unregister       chan *ClientUnregistration
	broadcastMessage chan *Message
	broadcastEvent   chan *Event
	createRoom       chan *Room
	deleteRoom       chan string
	getRoom          chan roomRequest
	listRooms        chan roomsRequest
}

// NewClient creates a new client
func NewClient(user *userModel.User, conn *websocket.Conn, roomID string, unregister chan<- *Client) *Client {
	return &Client{
		PingNotify: make(chan struct{}),
		User:       user,
		Conn:       conn,
		Send:       make(chan []byte, 1024),
		RoomID:     roomID,
		Unregister: unregister,
	}
}

// NewHub creates a new hub
func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*Room),

		register:   make(chan *ClientRegistration, 1),
		unregister: make(chan *ClientUnregistration, 1),

		broadcastMessage: make(chan *Message, 1),
		broadcastEvent:   make(chan *Event, 1),

		createRoom: make(chan *Room, 1),
		deleteRoom: make(chan string, 1),
		getRoom:    make(chan roomRequest, 1),
		listRooms:  make(chan roomsRequest, 1),
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
			fmt.Println(unreg)
			h.handleUnregister(unreg)

		case room := <-h.createRoom:
			h.handleCreateRoom(room)

		case roomID := <-h.deleteRoom:
			h.handleDeleteRoom(roomID)

		case msg := <-h.broadcastMessage:
			h.handleMessage(msg)

		case event := <-h.broadcastEvent:
			h.handleEvent(event)

		case req := <-h.getRoom:
			h.handleGetRoom(&req)

		case req := <-h.listRooms:
			h.handleListRooms(&req)
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
		Timestamp: time.Now().Unix(),
	}

	h.broadcastEvent <- event
}
