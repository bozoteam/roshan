package models

import (
	"encoding/json"
	"fmt"
	"maps"
	"time"
)

func (h *Hub) handleRegister(reg *ClientRegistration) {
	fmt.Printf("Registration request for room: %s, client: %s\n", reg.RoomID, reg.Client.User.Id)
	room, exists := h.rooms[reg.RoomID]
	if !exists {
		return
	}
	// Add client to room
	if room.Clients == nil {
		room.Clients = make(map[string]*Client)
	}
	room.Clients[reg.Client.User.Id] = reg.Client

	// Stop the timer when a user joins
	if room.emptyTimer != nil {
		room.emptyTimer.Stop()
		room.emptyTimer = nil
	}

	// Send current user list after a client joins
	h.sendUserList(room)
}

func (h *Hub) handleUnregister(unreg *ClientUnregistration) {
	fmt.Printf("Unregistration request for client: %s, room: %s\n", unreg.Client.Id, unreg.RoomID)
	// If specific room provided
	if unreg.RoomID != "" {
		if room, ok := h.rooms[unreg.RoomID]; ok {
			if _, ok := room.Clients[unreg.Client.Id]; ok {
				delete(room.Clients, unreg.Client.Id)
				close(unreg.Client.Send)
				// Send updated user list after a client leaves
				h.sendUserList(room)

				if len(room.Clients) == 0 {
					// If no clients left in the room, delete the room
					fmt.Printf("Deleting empty room: %s\n", unreg.RoomID)
					h.deleteRoom <- unreg.RoomID
				}
			}
		}
	}
}

func (h *Hub) handleCreateRoom(room *Room) {
	fmt.Printf("Creating room: %s, name: %s\n", room.ID, room.Name)
	h.rooms[room.ID] = room
	room.emptyTimer = time.AfterFunc(time.Second*10, func() {
		fmt.Printf("Deleting empty room: %s\n", room.ID)
		h.deleteRoom <- room.ID
	})
}

func (h *Hub) handleDeleteRoom(roomID string) {
	fmt.Printf("Deleting room: %s\n", roomID)
	if _, ok := h.rooms[roomID]; ok {
		delete(h.rooms, roomID)
	}
}

func (h *Hub) handleMessage(msg *Message) {
	fmt.Printf("Broadcasting message to room: %s, from: %s\n", msg.RoomID, msg.User.Email)
	if room, ok := h.rooms[msg.RoomID]; ok {
		// Serialize the message once
		data, err := json.Marshal(msg)
		if err != nil {
			return
		}
		// Send to all clients in the room
		for clientId, client := range room.Clients {
			select {
			case client.Send <- data:
				// Message sent successfully
			default:
				// Client buffer full, remove client
				delete(room.Clients, clientId)
				close(client.Send)
				h.sendUserList(room)
			}
		}
	}
}

func (h *Hub) handleEvent(event *Event) {
	fmt.Printf("Broadcasting event type brodcasting to room: %s\n", event.RoomID)
	if room, ok := h.rooms[event.RoomID]; ok {
		// Serialize the event once
		data, err := json.Marshal(event)
		if err != nil {
			return
		}
		// Send to all clients in the room
		for clientId, client := range room.Clients {
			select {
			case client.Send <- data:
				// Event sent successfully
			default:
				// Client buffer full, remove client
				delete(room.Clients, clientId)
				close(client.Send)
				h.sendUserList(room)
			}
		}
	}
}

func (h *Hub) handleGetRoom(req *roomRequest) {
	fmt.Printf("Room request received for id: %s\n", req.id)
	req.result <- h.rooms[req.id]
}

func (h *Hub) handleListRooms(req *roomsRequest) {
	fmt.Println("List rooms request received")
	// Create a copy of the rooms to avoid data races
	rooms := make([]*Room, 0, len(h.rooms))
	for _, room := range h.rooms {
		// Create a copy of the room
		roomCopy := &Room{
			ID:        room.ID,
			Name:      room.Name,
			CreatorID: room.CreatorID,
			Clients:   make(map[string]*Client),
		}
		// Copy clients
		maps.Copy(roomCopy.Clients, room.Clients)
		rooms = append(rooms, roomCopy)
	}
	req.result <- rooms
}
