package models

import (
	"github.com/bozoteam/roshan/modules/chat/models"
)

func NewGameRoom(name string, creatorId string, game string) *models.Room {
	return models.NewRoom(name, creatorId, []string{"team1", "team2", "all", "watcher"}, game)
}
