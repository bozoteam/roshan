package models

import (
	"github.com/bozoteam/roshan/helpers"
	"github.com/bozoteam/roshan/modules/websocket/ws_client"
)

type GameRoom struct {
	ID        string
	Name      string
	CreatorID string
	Clients   map[string]*ws_client.Client

	someoneEntered bool
}

var _ helpers.Cloneable[GameRoom] = (*GameRoom)(nil)

func (r *GameRoom) Clone() *GameRoom {
	return helpers.Clone(r)
}

func NewGameRoom(name string, creatorId string) *GameRoom {
	return &GameRoom{
		someoneEntered: false,

		ID:        helpers.GenUUID(),
		Name:      name,
		CreatorID: creatorId,
		Clients:   make(map[string]*ws_client.Client),
	}
}
