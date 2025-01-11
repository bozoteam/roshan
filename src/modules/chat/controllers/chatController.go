package controllers

import (
	"net/http"

	"github.com/bozoteam/roshan/src/modules/chat/models"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var hub = models.NewHub()

func init() {
	go hub.Run()
}

var upgrader = websocket.Upgrader{}

type RoomResponse struct {
	ID    string              `json:"id"`
	Users map[string]UserInfo `json:"users"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

func CreateRoom(c *gin.Context) {
	var room models.Room
	if err := c.ShouldBindJSON(&room); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	room.Users = make(map[*websocket.Conn]models.User)
	hub.Register <- &room
	roomResponse := RoomResponse{
		ID:    room.ID,
		Users: make(map[string]UserInfo),
	}
	c.JSON(http.StatusOK, roomResponse)
}

func ListRooms(c *gin.Context) {
	hub.Mutex.Lock()
	defer hub.Mutex.Unlock()
	rooms := make([]RoomResponse, 0, len(hub.Rooms))
	for _, room := range hub.Rooms {
		room.Mutex.Lock()
		users := make(map[string]UserInfo)
		for _, user := range room.Users {
			users[user.Conn.RemoteAddr().String()] = UserInfo{ID: user.ID, Username: user.Username}
		}
		roomResponse := RoomResponse{
			ID:    room.ID,
			Users: users,
		}
		room.Mutex.Unlock()
		rooms = append(rooms, roomResponse)
	}
	c.JSON(http.StatusOK, rooms)
}

func ListUsers(c *gin.Context) {
	roomID := c.Param("id")
	hub.Mutex.Lock()
	room, exists := hub.Rooms[roomID]
	hub.Mutex.Unlock()
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}
	room.Mutex.Lock()
	users := make([]UserInfo, 0, len(room.Users))
	for _, user := range room.Users {
		users = append(users, UserInfo{ID: user.ID, Username: user.Username})
	}
	room.Mutex.Unlock()
	c.JSON(http.StatusOK, users)
}

func DeleteRoom(c *gin.Context) {
	roomID := c.Param("id")
	hub.Mutex.Lock()
	room, exists := hub.Rooms[roomID]
	hub.Mutex.Unlock()
	if exists {
		hub.Unregister <- room
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
	}
}

func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	userID := c.Query("userID")
	username := c.Query("username")

	roomID := c.Param("id")
	hub.Mutex.Lock()
	room, exists := hub.Rooms[roomID]
	hub.Mutex.Unlock()
	if !exists {
		conn.Close()
		return
	}

	user := models.User{
		ID:       userID,
		Username: username,
		Conn:     conn,
	}

	room.Mutex.Lock()
	room.Users[conn] = user
	room.Mutex.Unlock()

	defer func() {
		conn.Close()
		room.Mutex.Lock()
		delete(room.Users, conn)
		room.Mutex.Unlock()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		hub.Broadcast <- models.Message{RoomID: roomID, Content: string(msg)}
	}
}
