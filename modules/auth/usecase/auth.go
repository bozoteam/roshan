package usecase

import (
	"errors"
	"log/slog"

	"context"

	log "github.com/bozoteam/roshan/adapter/log"
	"github.com/bozoteam/roshan/helpers"
	jwtRepository "github.com/bozoteam/roshan/modules/auth/repository/jwt"
	userRepository "github.com/bozoteam/roshan/modules/user/repository"
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

// TokenResponse represents the JWT tokens returned on successful authentication
type TokenResponse struct {
	AccessToken      string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken     string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType        string `json:"token_type" example:"Bearer"`
	ExpiresIn        uint64 `json:"expires_in" example:"3600"`
	RefreshExpiresIn uint64 `json:"refresh_expires_in" example:"3600"`
}

var (
	ErrAuthFailed     = errors.New("authentication failed")
	ErrInvalidRequest = errors.New("invalid request")
	ErrInvalidToken   = errors.New("invalid token")
)

func (c *AuthUsecase) Authenticate(ctx context.Context, email string, password string) (*TokenResponse, error) {
	user, err := c.userRepo.FindUserByEmail(email)
	if err != nil || !helpers.CheckPasswordHash(password, user.Password) {
		return nil, ErrAuthFailed
	}

	tokenData, err := c.jwtRepository.GenerateAccessAndRefreshTokens(user)
	if err != nil {
		return nil, errors.New("could not generate token")
	}

	err = c.userRepo.SaveRefreshToken(user, tokenData.RefreshToken)
	if err != nil {
		return nil, errors.New("could not generate token")
	}

	return &TokenResponse{
		AccessToken:      tokenData.AccessToken,
		RefreshToken:     tokenData.RefreshToken,
		TokenType:        tokenData.TokenType,
		ExpiresIn:        tokenData.ExpiresIn,
		RefreshExpiresIn: tokenData.RefreshExpiresIn,
	}, nil
}

// RefreshRequest represents the refresh token request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

func (c *AuthUsecase) Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	_, claims, err := c.jwtRepository.ValidateToken(refreshToken, jwtRepository.REFRESH_TOKEN)
	if err != nil {
		return nil, ErrInvalidToken
	}

	subject := claims.Subject
	if subject == "" {
		return nil, ErrInvalidToken
	}

	user, err := c.userRepo.FindUserByIdAndToken(subject, refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	tokenData, err := c.jwtRepository.GenerateAccessAndRefreshTokens(user)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = c.userRepo.SaveRefreshToken(user, tokenData.RefreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return &TokenResponse{
		AccessToken:      tokenData.AccessToken,
		RefreshToken:     tokenData.RefreshToken,
		TokenType:        tokenData.TokenType,
		ExpiresIn:        tokenData.ExpiresIn,
		RefreshExpiresIn: tokenData.RefreshExpiresIn,
	}, nil
}
