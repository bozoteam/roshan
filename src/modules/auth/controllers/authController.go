package controllers

import (
	"net/http"
	"time"

	"github.com/bozoteam/roshan/src/helpers"
	userDAO "github.com/bozoteam/roshan/src/modules/user/DAO"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
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
}

func NewAuthController(db *gorm.DB, jwtConf *JWTConfig) *AuthController {
	return &AuthController{
		db:        db,
		jwtConfig: jwtConf,
	}
}

func (c *JWTConfig) GetRefreshTokenKeyFunc(token *jwt.Token) (interface{}, error) {
	return c.refreshKey, nil
}

func (c *JWTConfig) GetTokenKeyFunc(token *jwt.Token) (interface{}, error) {
	return c.key, nil
}

func (c *AuthController) GenerateToken(subject string, secretKey []byte, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
	})
	return token.SignedString(secretKey)
}

// Authenticate authenticates a user and returns an access token and a refresh token
func (c *AuthController) Authenticate(context *gin.Context) {
	var json struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := context.BindJSON(&json); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := userDAO.FindUserByUsername(json.Username)
	if err != nil || !helpers.CheckPasswordHash(json.Password, user.Password) {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	accessTokenString, err := c.GenerateToken(user.Username, c.jwtConfig.key, time.Duration(c.jwtConfig.tokenDuration)*time.Minute)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate access token"})
		return
	}

	refreshTokenString, err := c.GenerateToken(user.Username, c.jwtConfig.refreshKey, time.Duration(c.jwtConfig.refreshTokenDuration)*time.Hour)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate refresh token"})
		return
	}

	user.RefreshToken = refreshTokenString
	if err := userDAO.SaveUser(user); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save refresh token"})
		return
	}

	context.SetCookie("refresh_token", refreshTokenString, int(c.jwtConfig.tokenDuration*3600), "/", "", false, true)
	context.JSON(http.StatusOK, gin.H{
		"access_token": accessTokenString,
	})
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

	user, err := userDAO.FindUserByUsernameAndToken(subject, json.RefreshToken)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	accessTokenString, err := c.GenerateToken(user.Username, c.jwtConfig.key, time.Duration(c.jwtConfig.tokenDuration)*time.Minute)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate access token"})
		return
	}
	context.JSON(http.StatusOK, gin.H{"access_token": accessTokenString})
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

	user, err := userDAO.FindUserByUsername(subject)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"name":     user.Name,
		"username": user.Username,
	})
}
