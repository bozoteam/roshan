package ws_hub

import (
	"github.com/bozoteam/roshan/modules/chat/models"
	"github.com/bozoteam/roshan/modules/websocket/ws_client"
)

type WsHub interface {
	Register(client *ws_client.Client, roomID string) *ws_client.Client
	Unregister(client *ws_client.Client, roomID string) *ws_client.Client
	GetRoom(id string) *models.Room
}
