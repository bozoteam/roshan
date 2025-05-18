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
	hub            ws_hub.WsHub
	logger         *slog.Logger
	jwtRepository  *jwtRepository.JWTRepository
	userRepository *userRepository.UserRepository
}

func NewChatUsecase(userRepository *userRepository.UserRepository, jwtRepository *jwtRepository.JWTRepository) *ChatUsecase {
	hub := models.NewHub()
	// go hub.Run()
	return &ChatUsecase{
		hub:            hub,
		logger:         log.LogWithModule("chat_usecase"),
		userRepository: userRepository,
		jwtRepository:  jwtRepository,
	}
}

// RoomResponse represents a chat room with its users
type RoomResponse struct {
	Id        string
	Name      string
	CreatorId string
	Users     []*userModel.User
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

	// Check if user is in room
	foundUser := false
	for _, client := range room.Clients {
		if client.User.Id == user.Id {
			foundUser = true
			break
		}
	}

	if !foundUser {
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

func (u *ChatUsecase) CreateRoom(ctx context.Context, name string) (*models.Room, error) {
	user := ctx.Value("user").(*userModel.User)

	room := models.NewRoom(name, user.Id)

	u.hub.CreateRoom(room)

	return room, nil
}

func (u *ChatUsecase) ListRooms(ctx context.Context) ([]*RoomResponse, error) {
	rooms := u.hub.ListRooms()

	responseRooms := make([]*RoomResponse, 0, len(rooms))
	for _, room := range rooms {
		users := make([]*userModel.User, 0, len(room.Clients))
		for _, client := range room.Clients {
			users = append(users, client.User)
		}

		responseRoom := &RoomResponse{
			CreatorId: room.CreatorID,
			Id:        room.ID,
			Name:      room.Name,
			Users:     users,
		}

		responseRooms = append(responseRooms, responseRoom)
	}

	return responseRooms, nil
}

func (u *ChatUsecase) DeleteRoom(ctx context.Context, roomId string) (*RoomResponse, error) {
	user := ctx.Value("user").(*userModel.User)

	room := u.hub.GetRoom(roomId)
	if room == nil {
		return nil, ErrRoomNotFound
	}

	if room.CreatorID != user.Id {
		return nil, ErrUserNotCreator
	}

	users := make([]*userModel.User, 0, len(room.Clients))
	for _, user := range room.Clients {
		users = append(users, user.User)
	}

	u.hub.DeleteRoom(room.ID)

	return &RoomResponse{
		Id:        room.ID,
		Name:      room.Name,
		CreatorId: room.CreatorID,
		Users:     users,
	}, nil
}

func (u *ChatUsecase) HandleWebSocket(ctx *gin.Context) {
	roomID := ctx.Param("id")
	token, ok := ctx.GetQuery("token")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}
	_, claims, err := u.jwtRepository.ValidateToken(token, jwtRepository.ACCESS_TOKEN)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	user, err := u.userRepository.FindUserById(claims.Subject)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

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
	u.hub.Register(client, roomID)

	u.logger.Info("User connected to room", "user_id", user.Id, "room_id", roomID)

	// Handle unregistration when the client disconnects
	// This runs in the same goroutine as HandleWebSocket
	client.WaitUnregister()
	u.hub.Unregister(client, roomID)
	u.logger.Info("User disconnected from room", "user_id", user.Id, "room_id", roomID)
}
