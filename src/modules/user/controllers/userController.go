package controllers

import (
	"net/http"
	"time"

	"github.com/bozoteam/roshan/src/helpers"
	userDAO "github.com/bozoteam/roshan/src/modules/user/DAO"
	"github.com/bozoteam/roshan/src/modules/user/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	db *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{db: db}

}

// CreateUser creates a new user
func (c *UserController) CreateUser(context *gin.Context) {
	var json struct {
		Name     string `json:"name" binding:"required"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := context.BindJSON(&json); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	hashedPassword, err := helpers.HashPassword(json.Password)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	user := models.User{
		Name:      json.Name,
		Username:  json.Username,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := userDAO.CreateUser(&user); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// FindUser finds a user by username
func (c *UserController) FindUser(context *gin.Context) {
	username := context.Param("username")

	user, err := userDAO.FindUserByUsername(username)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"name":     user.Name,
		"username": user.Username,
	})
}

// UpdateUser updates user data
func (c *UserController) UpdateUser(context *gin.Context) {
	var json struct {
		Name     *string `json:"name"`
		Username *string `json:"username"`
		Password *string `json:"password"`
	}
	if err := context.BindJSON(&json); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	username := context.Param("username")
	updates := make(map[string]interface{})
	if json.Name != nil {
		updates["name"] = *json.Name
	}
	if json.Username != nil {
		updates["username"] = *json.Username
	}
	if json.Password != nil {
		updates["password"] = *json.Password
	}

	updates["updated_at"] = time.Now()

	if err := userDAO.UpdateUser(username, updates); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update user"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// DeleteUser deletes a user by username
func (c *UserController) DeleteUser(context *gin.Context) {
	username := context.Param("username")

	if err := userDAO.DeleteUserByUsername(username); err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
