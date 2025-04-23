package models

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"time"
)

func (h *Hub) handleRegister(reg *clientRegistration) {
	fmt.Printf("Registration request for room: %s, client: %s\n", reg.RoomID, reg.Client.User.Id)
	room, exists := h.rooms[reg.RoomID]
	if !exists {
		reg.result <- nil
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

	reg.result <- reg.Client
}

func (h *Hub) handleUnregister(unreg *clientUnregistration) {
	fmt.Printf("Unregistration request for client: %s, room: %s\n", unreg.Client.Id, unreg.RoomID)
	// If specific room provided
	if room, ok := h.rooms[unreg.RoomID]; ok {
		if _, ok := room.Clients[unreg.Client.Id]; ok {
			delete(room.Clients, unreg.Client.Id)
			close(unreg.Client.Send)
			// Send updated user list after a client leaves
			h.sendUserList(room)

			if len(room.Clients) == 0 {
				// If no clients left in the room, delete the room
				fmt.Printf("Deleting empty room: %s\n", unreg.RoomID)
				h.handleDeleteRoom(&deleteRoom{
					roomId: unreg.RoomID,
					result: make(chan *Room, 1),
				})
			}
		}
	}

	unreg.result <- unreg.Client
}

func (h *Hub) handleCreateRoom(create *createRoom) {
	room := create.Room
	fmt.Printf("Creating room: %s, name: %s\n", room.ID, room.Name)
	h.rooms[room.ID] = room
	room.emptyTimer = time.AfterFunc(time.Second*10, func() {
		fmt.Printf("Deleting empty room: %s\n", room.ID)
		h.handleDeleteRoom(&deleteRoom{
			roomId: room.ID,
			result: make(chan *Room, 1),
		})
	})

	create.result <- room
}

func (h *Hub) handleDeleteRoom(deleteRoom *deleteRoom) {
	fmt.Printf("Deleting room: %s\n", deleteRoom.roomId)
	room, ok := h.rooms[deleteRoom.roomId]
	if ok {
		delete(h.rooms, room.ID)
	}

	deleteRoom.result <- room
}

func (h *Hub) handleMessage(send *sendMessage) {
	msg := send.Msg
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

	send.result <- msg
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
	req.result <- slices.Collect(maps.Values(h.rooms))
}
