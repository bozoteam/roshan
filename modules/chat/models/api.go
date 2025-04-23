package models

import "fmt"

// Register adds a client to a room
func (h *Hub) Register(client *Client, roomID string) *Client {
	result := make(chan *Client, 1)
	h.register <- &clientRegistration{
		Client: client,
		RoomID: roomID,

		result: result,
	}

	return <-result
}

// Unregister removes a client from a room
func (h *Hub) Unregister(client *Client, roomID string) *Client {
	result := make(chan *Client, 1)
	h.unregister <- &clientUnregistration{
		Client: client,
		RoomID: roomID,

		result: result,
	}

	return <-result
}

// CreateRoom adds a new room to the hub
func (h *Hub) CreateRoom(room *Room) *Room {
	result := make(chan *Room, 1)
	h.createRoom <- &createRoom{
		Room: room,

		result: result,
	}
	return <-result
}

// DeleteRoom removes a room from the hub
func (h *Hub) DeleteRoom(roomID string) *Room {
	result := make(chan *Room, 1)
	h.deleteRoom <- &deleteRoom{
		roomId: roomID,

		result: result,
	}

	fmt.Println("==========Deleting room:", roomID)
	return <-result
}

// BroadcastMessage sends a message to all clients in a room
func (h *Hub) BroadcastMessage(msg *Message) *Message {
	result := make(chan *Message, 1)
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
	result := make(chan *Room, 1)
	h.getRoom <- &roomRequest{id: id, result: result}
	return <-result
}

// ListRooms returns a list of all rooms
func (h *Hub) ListRooms() []*Room {
	result := make(chan []*Room, 1)
	h.listRooms <- &roomsRequest{result: result}
	return <-result
}
