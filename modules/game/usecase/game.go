package usecase

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/bozoteam/roshan/adapter/log"
	"github.com/bozoteam/roshan/modules/chat/models"
	gameModel "github.com/bozoteam/roshan/modules/game/models"
	userModel "github.com/bozoteam/roshan/modules/user/models"
	ws_hub "github.com/bozoteam/roshan/modules/websocket/hub"
	"github.com/bozoteam/roshan/modules/websocket/ws_client"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type GameUsecase struct {
	hub    *ws_hub.Hub
	logger *slog.Logger
}

type GameRoomResponse struct {
	*models.Room
}

func NewGameUsecase() *GameUsecase {
	return &GameUsecase{
		hub:    ws_hub.NewHub(),
		logger: log.LogWithModule("game_usecase"),
	}
}

func (u *GameUsecase) CreateRoom(ctx context.Context, name string, game string) (*GameRoomResponse, error) {
	user := ctx.Value("user").(*userModel.User)

	// TODO: validate game exists
	room := gameModel.NewGameRoom(name, user.Id, game)

	u.hub.CreateRoom(room)

	return &GameRoomResponse{
		Room: room,
	}, nil
}

func (u *GameUsecase) JoinRoom(ctx *gin.Context, roomID string, team string) {
	user := ctx.MustGet("user").(*userModel.User)

	room := u.hub.GetRoom(roomID)
	if room == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// TODO: validate team

	// Allow all origins for the WebSocket upgrade
	upgrader := websocket.Upgrader{
		HandshakeTimeout:  time.Second * 5,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		CheckOrigin:       func(r *http.Request) bool { return true },
		EnableCompression: false,
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		u.logger.Error("Failed to upgrade connection", "error", err)
		return
	}

	// Create client
	client := ws_client.NewClient(conn, user, roomID)

	// Register client to room
	u.hub.Register(client, roomID, team)

	u.logger.Info("User connected to room", "user_id", user.Id, "room_id", roomID)

	// Handle unregistration when the client disconnects
	// This runs in the same goroutine as HandleWebSocket
	client.WaitUnregister()
	u.hub.Unregister(client, roomID)
	u.logger.Info("User disconnected from room", "user_id", user.Id, "room_id", roomID)
}

func (u *GameUsecase) ListRooms(ctx context.Context) ([]*GameRoomResponse, error) {
	rooms := u.hub.ListRooms()

	responseRooms := make([]*GameRoomResponse, 0, len(rooms))
	for _, room := range rooms {
		responseRooms = append(responseRooms, &GameRoomResponse{
			Room: room.(*models.Room),
		},
		)
	}

	return responseRooms, nil
}
