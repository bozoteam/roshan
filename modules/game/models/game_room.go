package models

import (
	"github.com/bozoteam/roshan/modules/chat/models"
)

type GameRoom struct {
	models.Room
	Game string
}

func NewGameRoom(name string, creatorId string, game string) *GameRoom {
	return &GameRoom{
		Room: *models.NewRoom(name, creatorId, []string{"team1", "team2", "all", "watcher"}),
		Game: game,
	}
}
