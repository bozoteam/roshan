package jwtRepository

import (
	"errors"
	"time"

	"github.com/bozoteam/roshan/helpers"
	"github.com/bozoteam/roshan/modules/user/models"
	"github.com/golang-jwt/jwt/v5"
)

type TokenType int

const (
	ACCESS_TOKEN  TokenType = 0
	REFRESH_TOKEN TokenType = 1
)

func NewJWTRepository() *JWTRepository {
	return &JWTRepository{
		secretKey:            []byte(helpers.GetEnv("JWT_SECRET")),
		refreshSecretKey:     []byte(helpers.GetEnv("JWT_REFRESH_SECRET")),
		tokenDuration:        time.Duration(helpers.GetEnvAsInt("JWT_TOKEN_EXPIRATION")) * time.Second,
		refreshTokenDuration: time.Duration(helpers.GetEnvAsInt("JWT_REFRESH_TOKEN_EXPIRATION")) * time.Second,
		issuer:               "roshan",
	}
}

type JWTRepository struct {
	secretKey            []byte
	refreshSecretKey     []byte
	tokenDuration        time.Duration
	refreshTokenDuration time.Duration
	issuer               string
}

type CustomClaims struct {
	Email     string    `json:"email"`
	TokenType TokenType `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

func (c *JWTRepository) GetRefreshTokenKeyFunc(token *jwt.Token) (any, error) {
	return c.refreshSecretKey, nil
}

func (c *JWTRepository) GetTokenKeyFunc(token *jwt.Token) (any, error) {
	return c.secretKey, nil
}

type TokenData struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        uint64 `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn uint64 `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
}

func (s *JWTRepository) GenerateAccessAndRefreshTokens(user *models.User) (*TokenData, error) {
	now := time.Now()

	accessToken, err := s.generateToken(user, ACCESS_TOKEN, now)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(user, REFRESH_TOKEN, now)
	if err != nil {
		return nil, err
	}

	return &TokenData{
		AccessToken:      accessToken,
		ExpiresIn:        uint64(s.tokenDuration.Seconds()),
		RefreshToken:     refreshToken,
		RefreshExpiresIn: uint64((s.refreshTokenDuration).Seconds()),
		TokenType:        "Bearer",
		Scope:            "email",
	}, nil
}

func (s *JWTRepository) generateToken(user *models.User, tokenType TokenType, now time.Time) (string, error) {
	uuid := helpers.GenUUID()

	var duration time.Duration
	var signingKey []byte

	switch tokenType {
	case ACCESS_TOKEN:
		duration = s.tokenDuration
		signingKey = s.secretKey
	case REFRESH_TOKEN:
		duration = s.refreshTokenDuration
		signingKey = s.refreshSecretKey
	}

	claims := CustomClaims{
		TokenType: tokenType,
		Email:     user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid,
			Subject:   user.Id,
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "roshan",
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signingKey)
}

func (s *JWTRepository) ValidateToken(tokenString string, expectedKind TokenType) (*jwt.Token, *CustomClaims, error) {
	claims := &CustomClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		switch expectedKind {
		case ACCESS_TOKEN:
			return s.secretKey, nil
		case REFRESH_TOKEN:
			return s.refreshSecretKey, nil
		default:
			panic("invalid token kind")
		}
	}, jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(s.issuer),
		jwt.WithExpirationRequired())
	if err != nil {
		return nil, nil, err
	}

	if claims.TokenType != expectedKind {
		return nil, nil, errors.New("invalid token kind")
	}

	if token.Valid == false {
		return nil, nil, errors.New("invalid token")
	}

	return token, claims, err
}
