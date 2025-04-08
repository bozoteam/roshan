package controllers

import (
	"log/slog"
	"net/http"

	"github.com/bozoteam/roshan/src/helpers"
	log "github.com/bozoteam/roshan/src/log"
	"github.com/bozoteam/roshan/src/modules/user/models"
	userRepository "github.com/bozoteam/roshan/src/modules/user/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserController struct {
	logger   *slog.Logger
	userRepo *userRepository.UserRepository
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{userRepo: userRepository.NewUserRepository(db), logger: log.WithModule("user_controller")}
}

// CreateUser creates a new user
func (c *UserController) CreateUser(context *gin.Context) {
	var json struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required"`
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

	id, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}
	user := &models.User{
		Id:       id.String(),
		Name:     json.Name,
		Email:    json.Email,
		Password: hashedPassword,
	}

	if err := models.ValidateUser(user); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user data"})
		return
	}

	if err := c.userRepo.SaveUser(user); err != nil {
		if helpers.IsErrorCode(err, "23505") {
			context.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
			return
		}
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// FindUser finds a user by username
func (c *UserController) FindUser(context *gin.Context) {
	user := context.MustGet("user").(*models.User)

	user, err := c.userRepo.FindUserByEmail(user.Email)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	context.JSON(http.StatusOK, user)
}

// UpdateUser updates user data
func (c *UserController) UpdateUser(context *gin.Context) {
	user := context.MustGet("user").(*models.User)

	var json struct {
		Name     *string `json:"name"`
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}
	if err := context.BindJSON(&json); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if json.Name != nil {
		user.Name = *json.Name
	}
	if json.Email != nil {
		user.Email = *json.Email
	}
	if json.Password != nil {
		pass, err := helpers.HashPassword(*json.Password)
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
			return
		}
		user.Password = pass
	}

	if err := models.ValidateUser(user); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user data"})
		return
	}

	if err := c.userRepo.SaveUser(user); err != nil {
		if helpers.IsErrorCode(err, "23505") {
			context.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
			return
		}
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// DeleteUser deletes a user by username
func (c *UserController) DeleteUser(context *gin.Context) {
	user := context.MustGet("user").(*models.User)

	if err := c.userRepo.DeleteUser(user); err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (c *UserController) GetUser(context *gin.Context) {
	user := context.MustGet("user").(*models.User)

	context.JSON(http.StatusOK, user)
}
