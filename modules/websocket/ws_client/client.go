package ws_client

import (
	userModel "github.com/bozoteam/roshan/modules/user/models"
	"github.com/bozoteam/roshan/modules/websocket/ws_pump"
	"github.com/gorilla/websocket"
)

// Client represents a connected user
type Client struct {
	*userModel.User
	send   chan []byte `json:"-"`
	RoomID string      `json:"-"`

	pump *ws_pump.Pump `json:"-"`
}

func (c *Client) WaitUnregister() {
	<-c.pump.Unregister
}

func (c *Client) GetSender() chan []byte {
	return c.send
}

// NewClient creates a new client
func NewClient(conn *websocket.Conn, user *userModel.User, roomID string) *Client {
	send := make(chan []byte, 8)

	c := &Client{
		User:   user,
		send:   send,
		RoomID: roomID,
		pump:   ws_pump.NewPump(conn, send),
	}

	c.pump.Start()

	return c
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
