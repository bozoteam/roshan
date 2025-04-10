package usecase

import (
	"log/slog"
	"net/http"

	log "github.com/bozoteam/roshan/internal/adapter/log"
	"github.com/bozoteam/roshan/internal/helpers"
	jwtRepository "github.com/bozoteam/roshan/internal/modules/auth/repository/jwt"
	userRepository "github.com/bozoteam/roshan/internal/modules/user/repository"
	"github.com/gin-gonic/gin"
)

type AuthUsecase struct {
	logger        *slog.Logger
	jwtRepository *jwtRepository.JWTRepository
	userRepo      *userRepository.UserRepository
}

func NewAuthUsecase(userRepository *userRepository.UserRepository, jwtRepository *jwtRepository.JWTRepository) *AuthUsecase {
	return &AuthUsecase{
		logger:        log.LogWithModule("auth_usecase"),
		jwtRepository: jwtRepository,
		userRepo:      userRepository,
	}
}

// AuthRequest represents the login credentials
type AuthRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// TokenResponse represents the JWT tokens returned on successful authentication
type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string `json:"token_type" example:"Bearer"`
	ExpiresIn    int    `json:"expires_in" example:"3600"`
}

// Authenticate godoc
// @Summary Authenticate user
// @Description Authenticate a user with email and password, returns JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body AuthRequest true "Login credentials"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Authentication failed"
// @Failure 500 {object} map[string]string "Server error"
// @Router /auth [post]
func (c *AuthUsecase) Authenticate(context *gin.Context) {
	var input AuthRequest
	if err := context.BindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := c.userRepo.FindUserByEmail(input.Email)
	if err != nil || !helpers.CheckPasswordHash(input.Password, user.Password) {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	tokenData, err := c.jwtRepository.GenerateAccessAndRefreshTokens(user)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate tokens"})
		return
	}

	err = c.userRepo.SaveRefreshToken(user, tokenData.RefreshToken)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate tokens"})
		return
	}

	context.JSON(http.StatusOK, tokenData)
}

// RefreshRequest represents the refresh token request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// Refresh godoc
// @Summary Refresh access token
// @Description Generate a new access token using a valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh_token body RefreshRequest true "Refresh token"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Invalid refresh token"
// @Failure 500 {object} map[string]string "Server error"
// @Router /auth/refresh [post]
func (c *AuthUsecase) Refresh(context *gin.Context) {
	var input RefreshRequest
	if err := context.BindJSON(&input); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	_, claims, err := c.jwtRepository.ValidateToken(input.RefreshToken, jwtRepository.REFRESH_TOKEN)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	subject := claims.Subject
	if subject == "" {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	user, err := c.userRepo.FindUserByIdAndToken(subject, input.RefreshToken)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	tokenData, err := c.jwtRepository.GenerateAccessAndRefreshTokens(user)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate tokens"})
		return
	}

	err = c.userRepo.SaveRefreshToken(user, tokenData.RefreshToken)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate tokens"})
		return
	}

	context.JSON(http.StatusOK, tokenData)
}
