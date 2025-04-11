package usecase

import (
	"log/slog"
	"net/http"
	"time"

	log "github.com/bozoteam/roshan/internal/adapter/log"
	jwtRepository "github.com/bozoteam/roshan/internal/modules/auth/repository/jwt"
	"github.com/bozoteam/roshan/internal/modules/chat/models"
	userModel "github.com/bozoteam/roshan/internal/modules/user/models"
	userRepository "github.com/bozoteam/roshan/internal/modules/user/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	Id    string            `json:"id" example:"f81d4fae-7dec-11d0-a765-00a0c91e6bf6"`
	Name  string            `json:"name" example:"General Discussion"`
	Users []*userModel.User `json:"users"`
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

// SendMessage godoc
// @Summary Send a message to a chat room
// @Description Send a message to a specific chat room that the user is a member of
// @Tags chat
// @Accept json
// @Produce json
// @Param id path string true "Room ID"
// @Param message body MessageRequest true "Message content"
// @Security BearerAuth
// @Success 200 {object} nil
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 403 {object} map[string]string "User not in room"
// @Failure 404 {object} map[string]string "Room not found"
// @Router /chat/rooms/message/{id} [post]
func (cc *ChatUsecase) SendMessage(context *gin.Context) {
	user := context.MustGet("user").(*userModel.User)
	roomID := context.Param("id")

	room := cc.hub.GetRoom(roomID)
	if room == nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
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
		context.JSON(http.StatusForbidden, gin.H{"error": "User not found in room"})
		return
	}

	var input MessageRequest
	if err := context.BindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Create message with proper metadata
	message := &models.Message{
		RoomID:    roomID,
		User:      user,
		Content:   input.Message,
		Timestamp: time.Now().Unix(),
	}

	// Broadcast the message
	cc.hub.BroadcastMessage(message)
	context.JSON(http.StatusOK, nil)
}

// CreateRoom godoc
// @Summary Create a new chat room
// @Description Create a new chat room with the authenticated user as creator
// @Tags chat
// @Accept json
// @Produce json
// @Param room body RoomCreateRequest true "Room information"
// @Security BearerAuth
// @Success 200 {object} RoomCreateResponse
// @Failure 400 {object} map[string]string "Invalid request"
// @Router /chat/rooms [post]
func (cc *ChatUsecase) CreateRoom(context *gin.Context) {
	user := context.MustGet("user").(*userModel.User)

	var input RoomCreateRequest
	if err := context.BindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	uuid, err := uuid.NewV7()
	if err != nil {
		cc.logger.Error("Failed to generate UUID", "error", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create room"})
		return
	}

	room := &models.Room{
		ID:        uuid.String(),
		Name:      input.RoomName,
		CreatorID: user.Id,
		Clients:   make(map[string]*models.Client),
	}

	cc.hub.CreateRoom(room)
	context.JSON(http.StatusOK, RoomCreateResponse{Id: room.ID})
}

// ListRooms godoc
// @Summary List all chat rooms
// @Description Get a list of all available chat rooms and their members
// @Tags chat
// @Produce json
// @Success 200 {array} RoomResponse
// @Router /chat/rooms [get]
func (cc *ChatUsecase) ListRooms(context *gin.Context) {
	rooms := cc.hub.ListRooms()

	responseRooms := make([]RoomResponse, 0, len(rooms))
	for _, room := range rooms {
		clients := make([]*userModel.User, 0, len(room.Clients))
		for _, client := range room.Clients {
			clients = append(clients, client.User)
		}

		responseRoom := RoomResponse{
			Id:    room.ID,
			Name:  room.Name,
			Users: clients,
		}

		responseRooms = append(responseRooms, responseRoom)
	}

	context.JSON(http.StatusOK, responseRooms)
}

// DeleteRoom godoc
// @Summary Delete a chat room
// @Description Delete a specific chat room
// @Tags chat
// @Produce json
// @Param id path string true "Room ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string "Room not found"
// @Router /chat/rooms/{id} [delete]
func (cc *ChatUsecase) DeleteRoom(c *gin.Context) {
	roomID := c.Param("id")

	room := cc.hub.GetRoom(roomID)
	if room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	cc.hub.DeleteRoom(roomID)
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// HandleWebSocket godoc
// @Summary Connect to a chat room via WebSocket
// @Description Establish a WebSocket connection to a specific chat room
// @Tags chat
// @Param id path string true "Room ID"
// @Param token query string true "Auth token"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Token is required"
// @Failure 401 {object} map[string]string "Invalid token or user not found"
// @Failure 404 {object} map[string]string "Room not found"
// @Router /chat/rooms/{id}/ws [get]
func (cc *ChatUsecase) HandleWebSocket(context *gin.Context) {
	roomID := context.Param("id")
	token, ok := context.GetQuery("token")
	if !ok {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}
	_, claims, err := cc.jwtRepository.ValidateToken(token, jwtRepository.ACCESS_TOKEN)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	user, err := cc.userRepository.FindUserById(claims.Subject)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	room := cc.hub.GetRoom(roomID)
	if room == nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Allow all origins for the WebSocket upgrade
	cc.upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := cc.upgrader.Upgrade(context.Writer, context.Request, nil)
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
