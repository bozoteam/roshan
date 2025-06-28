package models

import userModel "github.com/bozoteam/roshan/modules/user/models"

// Message represents a chat message
type Message struct {
	RoomID    string          `json:"room_id"`
	User      *userModel.User `json:"user"`
	Content   string          `json:"content"`
	Timestamp int64           `json:"timestamp"`
}
