package models

// Register adds a client to a room
func (h *Hub) Register(client *Client, roomID string) {
	h.register <- &ClientRegistration{
		Client: client,
		RoomID: roomID,
	}
}

// Unregister removes a client from a room
func (h *Hub) Unregister(client *Client, roomID string) {
	h.unregister <- &ClientUnregistration{
		Client: client,
		RoomID: roomID,
	}
}

// CreateRoom adds a new room to the hub
func (h *Hub) CreateRoom(room *Room) {
	h.createRoom <- room
}

// DeleteRoom removes a room from the hub
func (h *Hub) DeleteRoom(roomID string) {
	h.deleteRoom <- roomID
}

// BroadcastMessage sends a message to all clients in a room
func (h *Hub) BroadcastMessage(msg *Message) {
	h.broadcastMessage <- msg
}

// BroadcastEvent sends an event to all clients in a room
func (h *Hub) BroadcastEvent(event *Event) {
	h.broadcastEvent <- event
}

// GetRoom returns a room by ID
func (h *Hub) GetRoom(id string) *Room {
	result := make(chan *Room, 1)
	h.getRoom <- roomRequest{id: id, result: result}
	return <-result
}

// ListRooms returns a list of all rooms
func (h *Hub) ListRooms() []*Room {
	result := make(chan []*Room, 1)
	h.listRooms <- roomsRequest{result: result}
	return <-result
}
