package ws_hub

import (
	"github.com/bozoteam/roshan/modules/chat/models"
	"github.com/bozoteam/roshan/modules/websocket/ws_client"
)

type WsHub interface {
	Register(client *ws_client.Client, roomID string)
	Unregister(client *ws_client.Client, roomID string)

	CreateRoom(room *models.Room)
	DeleteRoom(roomId string)

	GetRoom(roomId string) *models.Room
	ListRooms() []*models.Room

	BroadcastBytes(roomId string, data []byte)
}
