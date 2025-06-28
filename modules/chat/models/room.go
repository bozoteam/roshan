package models

import (
	"github.com/bozoteam/roshan/helpers"
	ws_hub "github.com/bozoteam/roshan/modules/websocket/hub"
)

func (r *Room) UserIsInRoom(userId string) bool {
	_, exists := r.Clients[userId]
	return exists
}

// Room implements the RoomI interface for chat rooms
type Room struct {
	ID        string
	Name      string
	CreatorID string
	Clients   map[string]ws_hub.ClientI

	someoneEntered bool
}

func (r *Room) GetID() string {
	return r.ID
}

func (r *Room) GetClients() map[string]ws_hub.ClientI {
	return r.Clients
}

func (r *Room) SetSomeoneEntered(entered bool) {
	r.someoneEntered = entered
}

func (r *Room) GetSomeoneEntered() bool {
	return r.someoneEntered
}

func (r *Room) Clone() ws_hub.RoomI {
	return helpers.Clone(r)
}

func NewRoom(name string, creatorId string) *Room {
	return &Room{
		ID:        helpers.GenUUID(),
		Name:      name,
		CreatorID: creatorId,
		Clients:   make(map[string]ws_hub.ClientI),

		someoneEntered: false,
	}
}
