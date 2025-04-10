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
	return &UserUsecase{
		userRepo: userRepository.NewUserRepository(db),
		logger:   log.LogWithModule("user_usecase"),
	}
}

// UserCreateInput represents the input for creating a user
type UserCreateInput struct {
	Name     string `json:"name" binding:"required" example:"John Doe"`
	Email    string `json:"email" binding:"required" example:"john@example.com"`
	Password string `json:"password" binding:"required" example:"securepassword"`
}

// UserCreateResponse represents the response for a successful user creation
type UserCreateResponse struct {
	Message string `json:"message" example:"User created successfully"`
}

// CreateUser godoc
// @Summary Create a new user
// @Description Register a new user with name, email, and password
// @Tags users
// @Accept json
// @Produce json
// @Param user body UserCreateInput true "User information"
// @Success 201 {object} UserCreateResponse "User created successfully"
// @Failure 400 {object} map[string]string "Invalid request or user already exists"
// @Failure 500 {object} map[string]string "Server error"
// @Router /user [post]
func (c *UserUsecase) CreateUser(context *gin.Context) {
	var input UserCreateInput
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

	context.JSON(http.StatusCreated, UserCreateResponse{Message: "User created successfully"})
}

// UserResponse represents a user for API responses
type UserResponse struct {
	Id    string `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name  string `json:"name" example:"John Doe"`
	Email string `json:"email" example:"john@example.com"`
}

// FindUser godoc
// @Summary Find user by email
// @Description Find a user by their email address
// @Tags users
// @Accept json
// @Produce json
// @Param email path string true "Email address"
// @Success 200 {object} UserResponse
// @Failure 404 {object} map[string]string "User not found"
// @Router /user/{email} [get]
func (c *UserUsecase) FindUser(context *gin.Context) {
	email := context.Param("email")
	user, err := c.userRepo.FindUserByEmail(email)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	response := UserResponse{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
	}

	context.JSON(http.StatusOK, response)
}

// UserUpdateInput represents the input for updating a user
type UserUpdateInput struct {
	Name     *string `json:"name" example:"John Smith"`
	Email    *string `json:"email" example:"john.smith@example.com"`
	Password *string `json:"password" example:"newsecurepassword"`
}

// UserUpdateResponse represents the response for a successful user update
type UserUpdateResponse struct {
	Message string `json:"message" example:"User updated successfully"`
}

// UpdateUser godoc
// @Summary Update user information
// @Description Update the authenticated user's information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body UserUpdateInput true "User information to update"
// @Success 200 {object} UserUpdateResponse "User updated successfully"
// @Failure 400 {object} map[string]string "Invalid request or user already exists"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /user [put]
func (c *UserUsecase) UpdateUser(context *gin.Context) {
	user := context.MustGet("user").(*models.User)
	var input UserUpdateInput
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

	context.JSON(http.StatusOK, UserUpdateResponse{Message: "User updated successfully"})
}

// DeleteUserResponse represents the response for a successful user deletion
type DeleteUserResponse struct {
	Message string `json:"message" example:"User deleted successfully"`
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete the authenticated user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} DeleteUserResponse "User deleted successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Router /user [delete]
func (c *UserUsecase) DeleteUser(context *gin.Context) {
	user := context.MustGet("user").(*models.User)
	if err := c.userRepo.DeleteUser(user); err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	context.JSON(http.StatusOK, DeleteUserResponse{Message: "User deleted successfully"})
}

// GetUser godoc
// @Summary Get user information
// @Description Get the authenticated user's information
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /user [get]
func (c *UserUsecase) GetUser(context *gin.Context) {
	user := context.MustGet("user").(*models.User)

	response := UserResponse{
		Id:    user.Id,
		Name:  user.Name,
		Email: user.Email,
	}

	context.JSON(http.StatusOK, response)
}
