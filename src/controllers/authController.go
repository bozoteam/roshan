package controllers

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	adapter "github.com/bozoteam/roshan/src/database"
	"github.com/bozoteam/roshan/src/helpers"
	"github.com/bozoteam/roshan/src/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var jwtKey []byte
var refreshJwtKey []byte
var tokenDuration int64
var refreshTokenDuration int64

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	jwtKey = []byte(os.Getenv("JWT_SECRET"))
	refreshJwtKey = []byte(os.Getenv("JWT_REFRESH_SECRET"))

	tokenDuration, err = strconv.ParseInt(os.Getenv("JWT_TOKEN_EXPIRATION"), 10, 64)
	if err != nil {
		log.Fatalf("Invalid JWT_TOKEN_EXPIRATION: %v", err)
	}

	refreshTokenDuration, err = strconv.ParseInt(os.Getenv("JWT_REFRESH_TOKEN_EXPIRATION"), 10, 64)
	if err != nil {
		log.Fatalf("Invalid JWT_REFRESH_TOKEN_EXPIRATION: %v", err)
	}
}

func Authenticate(context *gin.Context) {
	db, err := adapter.GetDBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	var json struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := context.BindJSON(&json); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var user models.User
	if err := db.Where("username = ? AND password = ?", json.Username, json.Password).First(&user).Error; err != nil {
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

	context.JSON(http.StatusOK, gin.H{
		"access_token":  accessTokenString,
		"refresh_token": refreshTokenString,
	})
}

func Refresh(context *gin.Context) {
	db, err := adapter.GetDBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

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
