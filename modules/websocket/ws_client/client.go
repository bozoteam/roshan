package ws_client

import (
	userModel "github.com/bozoteam/roshan/modules/user/models"
	"github.com/bozoteam/roshan/modules/websocket/ws_pump"
	"github.com/gorilla/websocket"
)

// Client represents a connected user
type Client struct {
	*userModel.User
	Send   chan []byte `json:"-"`
	RoomID string      `json:"-"`

	Pump *ws_pump.Pump `json:"-"`
}

// NewClient creates a new client
func NewClient(conn *websocket.Conn, user *userModel.User, roomID string) *Client {
	send := make(chan []byte, 1024)
	return &Client{
		User:   user,
		Send:   send,
		RoomID: roomID,
		Pump:   ws_pump.NewPump(conn, send),
	}
}

// ClientRegistration holds data for registering a client to a room
type ClientRegistration struct {
	Client *Client
	RoomID string

	Result chan *Client
}

// ClientUnregistration holds data for unregistering a client
type ClientUnregistration struct {
	Client *Client
	RoomID string

	Result chan *Client
}
