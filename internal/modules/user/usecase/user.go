package usecase

import (
	"log/slog"
	"net/http"

	log "github.com/bozoteam/roshan/internal/adapter/log"
	"github.com/bozoteam/roshan/internal/helpers"
	"github.com/bozoteam/roshan/internal/modules/user/models"
	userRepository "github.com/bozoteam/roshan/internal/modules/user/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserUsecase struct {
	logger   *slog.Logger
	userRepo *userRepository.UserRepository
}

func NewUserUsecase(db *gorm.DB) *UserUsecase {
	return &UserUsecase{userRepo: userRepository.NewUserRepository(db), logger: log.LogWithModule("user_usecase")}
}

// CreateUser creates a new user
func (c *UserUsecase) CreateUser(context *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := context.BindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	hashedPassword, err := helpers.HashPassword(input.Password)
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
		Name:     input.Name,
		Email:    input.Email,
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
func (c *UserUsecase) FindUser(context *gin.Context) {
	user := context.MustGet("user").(*models.User)

	user, err := c.userRepo.FindUserByEmail(user.Email)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	context.JSON(http.StatusOK, user)
}

// UpdateUser updates user data
func (c *UserUsecase) UpdateUser(context *gin.Context) {
	user := context.MustGet("user").(*models.User)

	var input struct {
		Name     *string `json:"name"`
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}
	if err := context.BindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Password != nil {
		pass, err := helpers.HashPassword(*input.Password)
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
func (c *UserUsecase) DeleteUser(context *gin.Context) {
	user := context.MustGet("user").(*models.User)

	if err := c.userRepo.DeleteUser(user); err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (c *UserUsecase) GetUser(context *gin.Context) {
	user := context.MustGet("user").(*models.User)

	context.JSON(http.StatusOK, user)
}
