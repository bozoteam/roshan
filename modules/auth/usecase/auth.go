package usecase

import (
	"log/slog"

	"context"

	log "github.com/bozoteam/roshan/adapter/log"
	"github.com/bozoteam/roshan/helpers"
	jwtRepository "github.com/bozoteam/roshan/modules/auth/repository/jwt"
	userModel "github.com/bozoteam/roshan/modules/user/models"
	userRepository "github.com/bozoteam/roshan/modules/user/repository"
	"github.com/bozoteam/roshan/roshan_errors"
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

func (u *AuthUsecase) Authenticate(ctx context.Context, email string, password string) (*TokenResponse, error) {
	user, err := u.userRepo.FindUserByEmail(email)
	if err != nil || !helpers.CheckPasswordHash(password, user.Password) {
		return nil, roshan_errors.ErrAuthFailed
	}

	tokenData, err := u.jwtRepository.GenerateAccessAndRefreshTokens(user)
	if err != nil {
		return nil, roshan_errors.ErrInternalServerError
	}

	err = u.userRepo.SaveRefreshToken(user, tokenData.RefreshToken)
	if err != nil {
		return nil, roshan_errors.ErrInternalServerError
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

func (u *AuthUsecase) Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	_, claims, err := u.jwtRepository.ValidateToken(refreshToken, jwtRepository.REFRESH_TOKEN)
	if err != nil {
		return nil, roshan_errors.ErrInvalidToken
	}

	subject := claims.Subject
	if subject == "" {
		return nil, roshan_errors.ErrInvalidToken
	}

	user, err := u.userRepo.FindUserByIdAndToken(subject, refreshToken)
	if err != nil {
		return nil, roshan_errors.ErrInvalidToken
	}

	tokenData, err := u.jwtRepository.GenerateAccessAndRefreshTokens(user)
	if err != nil {
		return nil, roshan_errors.ErrInvalidToken
	}

	err = u.userRepo.SaveRefreshToken(user, tokenData.RefreshToken)
	if err != nil {
		return nil, roshan_errors.ErrInvalidToken
	}

	return &TokenResponse{
		AccessToken:      tokenData.AccessToken,
		RefreshToken:     tokenData.RefreshToken,
		TokenType:        tokenData.TokenType,
		ExpiresIn:        tokenData.ExpiresIn,
		RefreshExpiresIn: tokenData.RefreshExpiresIn,
	}, nil
}

func (u *AuthUsecase) Logout(ctx context.Context) error {
	user := ctx.Value("user").(*userModel.User)
	return u.userRepo.DeleteRefreshToken(user)
}
