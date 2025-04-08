package controllers

import (
	"net/http"
	"time"

	"github.com/bozoteam/roshan/src/helpers"
	"github.com/bozoteam/roshan/src/modules/user/models"
	userRepository "github.com/bozoteam/roshan/src/modules/user/repository"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func NewJWTConfig() *JWTConfig {
	return &JWTConfig{
		key:                  []byte(helpers.GetEnv("JWT_SECRET")),
		refreshKey:           []byte(helpers.GetEnv("JWT_REFRESH_SECRET")),
		tokenDuration:        helpers.GetEnvAsInt("JWT_TOKEN_EXPIRATION"),
		refreshTokenDuration: helpers.GetEnvAsInt("JWT_REFRESH_TOKEN_EXPIRATION"),
	}
}

type JWTConfig struct {
	key                  []byte
	refreshKey           []byte
	tokenDuration        int64
	refreshTokenDuration int64
}

type AuthController struct {
	db        *gorm.DB
	jwtConfig *JWTConfig
	userRepo  *userRepository.UserRepository
}

func NewAuthController(db *gorm.DB, jwtConf *JWTConfig) *AuthController {
	return &AuthController{
		db:        db,
		jwtConfig: jwtConf,
		userRepo:  userRepository.NewUserRepository(db),
	}
}

func (c *JWTConfig) GetRefreshTokenKeyFunc(token *jwt.Token) (any, error) {
	return c.refreshKey, nil
}

func (c *JWTConfig) GetTokenKeyFunc(token *jwt.Token) (any, error) {
	return c.key, nil
}

func (c *AuthController) generateToken(user *models.User, secretKey []byte, duration time.Duration, notBefore time.Time) (string, error) {
	type CustomClaims struct {
		Email string `json:"email"`
		jwt.RegisteredClaims
	}

	uuid, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}

	claims := CustomClaims{
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.String(),
			Subject:   user.Id,
			ExpiresAt: jwt.NewNumericDate(notBefore.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(notBefore),
			Issuer:    "roshan",
			NotBefore: jwt.NewNumericDate(notBefore),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secretKey)
}

func (c *AuthController) generateAndReturnToken(context *gin.Context, user *models.User) {
	notBefore := time.Now()

	accessTokenString, err := c.generateToken(user, c.jwtConfig.key, time.Duration(c.jwtConfig.tokenDuration)*time.Second, notBefore)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate access token"})
		return
	}

	refreshTokenString, err := c.generateToken(user, c.jwtConfig.refreshKey, time.Duration(c.jwtConfig.refreshTokenDuration)*time.Second, notBefore)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate refresh token"})
		return
	}

	user.RefreshToken = refreshTokenString
	if err := c.userRepo.SaveUser(user); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save refresh token"})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"access_token":       accessTokenString,
		"expires_in":         c.jwtConfig.tokenDuration,
		"refresh_token":      refreshTokenString,
		"refresh_expires_in": c.jwtConfig.refreshTokenDuration,
		"token_type":         "Bearer",
		"scope":              "email",
	})
}

// Authenticate authenticates a user and returns an access token and a refresh token
func (c *AuthController) Authenticate(context *gin.Context) {
	var json struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := context.BindJSON(&json); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := c.userRepo.FindUserByEmail(json.Email)
	if err != nil || !helpers.CheckPasswordHash(json.Password, user.Password) {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	c.generateAndReturnToken(context, user)
}

// Refresh generates a new access token using a refresh token
func (c *AuthController) Refresh(context *gin.Context) {
	var json struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := context.BindJSON(&json); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var claims jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(json.RefreshToken, &claims, c.jwtConfig.GetRefreshTokenKeyFunc)
	if err != nil || !token.Valid {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	subject := claims.Subject
	if subject == "" {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	user, err := c.userRepo.FindUserByIdAndToken(subject, json.RefreshToken)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.generateAndReturnToken(context, user)
}

// GetLoggedInUser returns the user data of the logged in user
func (c *AuthController) GetLoggedInUser(context *gin.Context) {
	tokenString := context.GetHeader("Authorization")
	if tokenString == "" {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
		return
	}

	var claims jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, c.jwtConfig.GetTokenKeyFunc)
	if err != nil || !token.Valid {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	subject, err := token.Claims.GetSubject()
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	user, err := c.userRepo.FindUserByEmail(subject)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	context.JSON(http.StatusOK, user)
}
