package usecase

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"context"

	log "github.com/bozoteam/roshan/adapter/log"
	"github.com/bozoteam/roshan/helpers"
	jwtRepository "github.com/bozoteam/roshan/modules/auth/repository/jwt"
	"github.com/bozoteam/roshan/modules/chat/models"
	userModel "github.com/bozoteam/roshan/modules/user/models"
	userRepository "github.com/bozoteam/roshan/modules/user/repository"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ChatUsecase struct {
	hub            *models.Hub
	upgrader       *websocket.Upgrader
	logger         *slog.Logger
	jwtRepository  *jwtRepository.JWTRepository
	userRepository *userRepository.UserRepository
}

func NewChatUsecase(userRepository *userRepository.UserRepository, jwtRepository *jwtRepository.JWTRepository) *ChatUsecase {
	hub := models.NewHub()
	go hub.Run()
	return &ChatUsecase{
		hub:            hub,
		upgrader:       new(websocket.Upgrader),
		logger:         log.LogWithModule("chat_usecase"),
		userRepository: userRepository,
		jwtRepository:  jwtRepository,
	}
}

// RoomResponse represents a chat room with its users
type RoomResponse struct {
	Id        string `json:"id" example:"f81d4fae-7dec-11d0-a765-00a0c91e6bf6"`
	Name      string `json:"name" example:"General Discussion"`
	CreatorId string
	Users     []*userModel.User `json:"users"`
}

// RoomCreateRequest represents data needed to create a chat room
type RoomCreateRequest struct {
	RoomName string `json:"room_name" binding:"required" example:"General Discussion"`
}

// RoomCreateResponse represents the response after creating a room
type RoomCreateResponse struct {
	Id string `json:"id" example:"f81d4fae-7dec-11d0-a765-00a0c91e6bf6"`
}

// MessageRequest represents the structure of a chat message
type MessageRequest struct {
	Message string `json:"message" binding:"required" example:"Hello, everyone!"`
}

func (cc *ChatUsecase) SendMessage(context context.Context, content string, roomId string) error {
	user := context.Value("user").(*userModel.User)

	room := cc.hub.GetRoom(roomId)
	if room == nil {
		return errors.New("room not found")
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
		return errors.New("user not found in room")
	}

	// Create message with proper metadata
	message := &models.Message{
		RoomID:    roomId,
		User:      user,
		Content:   content,
		Timestamp: time.Now().Unix(),
	}

	// Broadcast the message
	cc.hub.BroadcastMessage(message)
	return nil
}

func (cc *ChatUsecase) CreateRoom(ctx context.Context, name string) (string, error) {
	user := ctx.Value("user").(*userModel.User)

	uuid := helpers.GenUUID()

	room := &models.Room{
		ID:        uuid,
		Name:      name,
		CreatorID: user.Id,
		Clients:   make(map[string]*models.Client),
	}

	cc.hub.CreateRoom(room)

	return uuid, nil
}

func (cc *ChatUsecase) ListRooms(ctx context.Context) ([]*RoomResponse, error) {
	rooms := cc.hub.ListRooms()

	responseRooms := make([]*RoomResponse, 0, len(rooms))
	for _, room := range rooms {
		clients := make([]*userModel.User, 0, len(room.Clients))
		for _, client := range room.Clients {
			clients = append(clients, client.User)
		}

		responseRoom := &RoomResponse{
			CreatorId: room.CreatorID,
			Id:        room.ID,
			Name:      room.Name,
			Users:     clients,
		}

		responseRooms = append(responseRooms, responseRoom)
	}

	return responseRooms, nil
}

func (cc *ChatUsecase) DeleteRoom(ctx context.Context, roomId string) (*RoomResponse, error) {
	user := ctx.Value("user").(*userModel.User)

	room := cc.hub.GetRoom(roomId)
	if room == nil {
		return nil, errors.New("room does not exist")
	}

	if room.CreatorID != user.Id {
		return nil, errors.New("not room creator!")
	}

	users := make([]*userModel.User, 0, len(room.Clients))
	for _, user := range room.Clients {
		users = append(users, user.User)
	}

	cc.hub.DeleteRoom(roomId)

	return &RoomResponse{
		Id:        room.ID,
		Name:      room.Name,
		CreatorId: room.CreatorID,
		Users:     users,
	}, nil
}

func (cc *ChatUsecase) HandleWebSocket(ctx *gin.Context) {
	roomID := ctx.Param("id")
	token, ok := ctx.GetQuery("token")
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}
	_, claims, err := cc.jwtRepository.ValidateToken(token, jwtRepository.ACCESS_TOKEN)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	user, err := cc.userRepository.FindUserById(claims.Subject)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	room := cc.hub.GetRoom(roomID)
	if room == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Allow all origins for the WebSocket upgrade
	cc.upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := cc.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		cc.logger.Error("Failed to upgrade connection", "error", err)
		return
	}

	// Create unregister channel
	unregister := make(chan *models.Client)

	// Create client
	client := models.NewClient(user, conn, roomID, unregister)

	// Start goroutines for reading and writing
	go client.ReadPump(cc.hub)
	go client.WritePump(cc.hub)

	// Register client to room
	cc.hub.Register(client, roomID)

	cc.logger.Info("User connected to room", "user_id", user.Id, "room_id", roomID)

	// Handle unregistration when the client disconnects
	// This runs in the same goroutine as HandleWebSocket
	clientToUnregister := <-unregister
	cc.hub.Unregister(clientToUnregister, roomID)
	cc.logger.Info("User disconnected from room", "user_id", user.Id, "room_id", roomID)
}
