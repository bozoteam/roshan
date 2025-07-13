package usecase

import (
	"log/slog"
	"net/http"
	"time"

	"context"

	"encoding/json"

	log "github.com/bozoteam/roshan/adapter/log"
	jwtRepository "github.com/bozoteam/roshan/modules/auth/repository/jwt"
	"github.com/bozoteam/roshan/modules/chat/models"
	chatModel "github.com/bozoteam/roshan/modules/chat/models"

	userModel "github.com/bozoteam/roshan/modules/user/models"
	userRepository "github.com/bozoteam/roshan/modules/user/repository"
	ws_hub "github.com/bozoteam/roshan/modules/websocket/hub"
	"github.com/bozoteam/roshan/modules/websocket/ws_client"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChatUsecase struct {
	hub            *ws_hub.Hub
	logger         *slog.Logger
	jwtRepository  *jwtRepository.JWTRepository
	userRepository *userRepository.UserRepository
}

func NewChatUsecase(
	userRepository *userRepository.UserRepository,
	jwtRepository *jwtRepository.JWTRepository,
) *ChatUsecase {
	hub := ws_hub.NewHub()
	// go hub.Run()
	return &ChatUsecase{
		hub:            hub,
		logger:         log.LogWithModule("chat_usecase"),
		userRepository: userRepository,
		jwtRepository:  jwtRepository,
	}
}

// ChatRoomResponse represents a chat room with its users
type ChatRoomResponse struct {
	*models.Room
}

var (
	ErrRoomNotFound       = status.Error(codes.NotFound, "room not found")
	ErrUserNotFoundInRoom = status.Error(codes.PermissionDenied, "user not found in room")
	ErrUserNotCreator     = status.Error(codes.PermissionDenied, "user cannot delete room, not creator")
)

func (u *ChatUsecase) SendMessage(ctx context.Context, content string, roomId string) error {
	user := ctx.Value("user").(*userModel.User)

	room := u.hub.GetRoom(roomId)
	if room == nil {
		return ErrRoomNotFound
	}

	if room.UserIsInRoom(user.Id) == false {
		return ErrUserNotFoundInRoom
	}

	// Create message with proper metadata
	message := &models.Message{
		RoomID:    roomId,
		User:      user,
		Content:   content,
		Timestamp: time.Now().UnixNano(),
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Broadcast the message
	go u.hub.BroadcastBytes(roomId, data)
	return nil
}

func (u *ChatUsecase) CreateRoom(ctx context.Context, name string) (*ChatRoomResponse, error) {
	user := ctx.Value("user").(*userModel.User)

	room := models.NewRoom(name, user.Id, []string{"chat"}, "chat")

	u.hub.CreateRoom(room)

	return &ChatRoomResponse{
		Room: room,
	}, nil
}

func (u *ChatUsecase) ListRooms(ctx context.Context) ([]*ChatRoomResponse, error) {
	rooms := u.hub.ListRooms()

	responseRooms := make([]*ChatRoomResponse, 0, len(rooms))
	for _, room := range rooms {
		responseRooms = append(responseRooms, &ChatRoomResponse{
			Room: room.(*models.Room),
		},
		)
	}

	return responseRooms, nil
}

func (u *ChatUsecase) DeleteRoom(ctx context.Context, roomId string) (*ChatRoomResponse, error) {
	user := ctx.Value("user").(*userModel.User)

	room := u.hub.GetRoom(roomId).(*chatModel.Room)
	if room == nil {
		return nil, ErrRoomNotFound
	}

	if room.CreatorID != user.Id {
		return nil, ErrUserNotCreator
	}

	u.hub.DeleteRoom(room.ID)

	return &ChatRoomResponse{
		Room: room,
	}, nil
}

func (u *ChatUsecase) JoinRoom(ctx *gin.Context, roomID string) {
	user := ctx.MustGet("user").(*userModel.User)

	room := u.hub.GetRoom(roomID)
	if room == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

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
	u.hub.Register(client, roomID, "chat")

	u.logger.Info("User connected to room", "user_id", user.Id, "room_id", roomID)

	// Handle unregistration when the client disconnects
	// This runs in the same goroutine as HandleWebSocket
	client.WaitUnregister()
	u.hub.Unregister(client, roomID)
	u.logger.Info("User disconnected from room", "user_id", user.Id, "room_id", roomID)
}
