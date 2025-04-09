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
	logger *slog.Logger

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

// Authenticate authenticates a user and returns an access token and a refresh token
func (c *AuthUsecase) Authenticate(context *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
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

	context.JSON(http.StatusOK, tokenData)
}

// Refresh generates a new access token using a refresh token
func (c *AuthUsecase) Refresh(context *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
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

	context.JSON(http.StatusOK, tokenData)
}
