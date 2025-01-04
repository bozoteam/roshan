package controllers

import (
	"net/http"
	"time"

	adapter "github.com/bozoteam/roshan/src/database"
	"github.com/bozoteam/roshan/src/helpers"
	"github.com/bozoteam/roshan/src/modules/user/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var jwtKey []byte
var refreshJwtKey []byte
var tokenDuration int64
var refreshTokenDuration int64

func init() {
	jwtKey = []byte(helpers.GetEnv("JWT_SECRET"))
	refreshJwtKey = []byte(helpers.GetEnv("JWT_REFRESH_SECRET"))
	tokenDuration = helpers.GetEnvAsInt("JWT_TOKEN_EXPIRATION")
	refreshTokenDuration = helpers.GetEnvAsInt("JWT_REFRESH_TOKEN_EXPIRATION")
}

// Authenticate authenticates a user and returns an access token and a refresh token
func Authenticate(context *gin.Context) {
	db := adapter.GetDBConnection()

	var json struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := context.BindJSON(&json); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	if err := db.Where("username = ?", json.Username).First(&user).Error; err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	if !helpers.CheckPasswordHash(json.Password, user.Password) {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed"})
		return
	}

	accessTokenString, err := helpers.GenerateToken(user.Username, jwtKey, time.Duration(tokenDuration)*time.Minute)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate access token"})
		return
	}

	refreshTokenString, err := helpers.GenerateToken(user.Username, refreshJwtKey, time.Duration(refreshTokenDuration)*time.Hour)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate refresh token"})
		return
	}

	user.RefreshToken = refreshTokenString
	if err := db.Save(&user).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save refresh token"})
		return
	}

	context.SetCookie("refresh_token", refreshTokenString, int(refreshTokenDuration*3600), "/", "", false, true)
	context.JSON(http.StatusOK, gin.H{
		"access_token": accessTokenString,
	})
}

// Refresh generates a new access token using a refresh token
func Refresh(context *gin.Context) {
	db := adapter.GetDBConnection()

	var json struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := context.BindJSON(&json); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	claims := &jwt.StandardClaims{}
	token, err := jwt.ParseWithClaims(json.RefreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return refreshJwtKey, nil
	})

	if err != nil || !token.Valid {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	var user models.User
	if err := db.Where("username = ? AND refresh_token = ?", claims.Subject, json.RefreshToken).First(&user).Error; err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	accessTokenString, err := helpers.GenerateToken(user.Username, jwtKey, time.Duration(tokenDuration)*time.Minute)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate access token"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"access_token": accessTokenString})
}

// GetLoggedInUser returns the user data of the logged in user
func GetLoggedInUser(context *gin.Context) {
	tokenString := context.GetHeader("Authorization")
	if tokenString == "" {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
		return
	}

	claims := &jwt.StandardClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	db := adapter.GetDBConnection()

	var user models.User
	if err := db.Where("username = ?", claims.Subject).First(&user).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"name":     user.Name,
		"username": user.Username,
	})
}
