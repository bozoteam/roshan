package controllers

import (
	"net/http"
	"time"

	"github.com/bozoteam/roshan/src/helpers"
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

	if err := c.db.Create(&user).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// FindUser finds a user by username
func (c *UserController) FindUser(context *gin.Context) {
	username := context.Param("username")
	var user models.User
	if err := c.db.Where("username = ?", username).First(&user).Error; err != nil {
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
	var user models.User
	if err := c.db.Where("username = ?", username).First(&user).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if json.Name != nil {
		user.Name = *json.Name
	}
	if json.Username != nil {
		user.Username = *json.Username
	}
	if json.Password != nil {
		hashedPassword, err := helpers.HashPassword(*json.Password)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
			return
		}
		user.Password = hashedPassword
	}

	user.UpdatedAt = time.Now()

	if err := c.db.Save(&user).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update user"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// DeleteUser deletes a user by username
func (c *UserController) DeleteUser(context *gin.Context) {
	username := context.Param("username")
	var user models.User
	if err := c.db.Where("username = ?", username).First(&user).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := c.db.Delete(&user).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
