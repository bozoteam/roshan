package models

import "github.com/bozoteam/roshan/modules/websocket/ws_client"

// Register adds a client to a room
func (h *Hub) Register(client *ws_client.Client, roomID string) *ws_client.Client {
	result := make(chan *ws_client.Client)
	h.register <- &ws_client.ClientRegistration{
		Client: client,
		RoomID: roomID,

		Result: result,
	}

	return <-result
}

// Unregister removes a client from a room
func (h *Hub) Unregister(client *ws_client.Client, roomID string) *ws_client.Client {
	result := make(chan *ws_client.Client)
	h.unregister <- &ws_client.ClientUnregistration{
		Client: client,
		RoomID: roomID,

		Result: result,
	}

	return <-result
}

// CreateRoom adds a new room to the hub
func (h *Hub) CreateRoom(room *Room) *Room {
	result := make(chan *Room)
	h.createRoom <- &createRoom{
		Room: room,

		result: result,
	}

	return <-result
}

// DeleteRoom removes a room from the hub
func (h *Hub) DeleteRoom(room *Room) *Room {
	result := make(chan *Room)
	h.deleteRoom <- &deleteRoom{
		roomId: room.ID,

		result: result,
	}

	return <-result
}

// BroadcastMessage sends a message to all clients in a room
func (h *Hub) BroadcastMessage(msg *Message) *Message {
	result := make(chan *Message)
	h.broadcastMessage <- &sendMessage{
		Msg: msg,

		result: result,
	}

	return <-result
}

// BroadcastEvent sends an event to all clients in a room
func (h *Hub) BroadcastEvent(event *Event) {
	h.broadcastEvent <- event
}

// GetRoom returns a room by ID
func (h *Hub) GetRoom(id string) *Room {
	result := make(chan *Room)
	h.getRoom <- &roomRequest{
		id:     id,
		result: result,
	}

	return <-result
}

// ListRooms returns a list of all rooms
func (h *Hub) ListRooms() []*Room {
	result := make(chan []*Room)
	h.listRooms <- &roomsRequest{
		result: result,
	}

	return <-result
}
