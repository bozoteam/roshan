package controllers

import (
	"log"
	"net/http"
	"time"

	adapter "github.com/bozoteam/roshan/src/database"
	"github.com/bozoteam/roshan/src/modules/user/models"
	"github.com/gin-gonic/gin"
)

// User creation handler
func CreateUser(context *gin.Context) {
	db, err := adapter.GetDBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	var json struct {
		Name     string `json:"name" binding:"required"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := context.BindJSON(&json); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := models.User{
		Name:      json.Name,
		Username:  json.Username,
		Password:  json.Password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&user).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// User search handler
func FindUser(context *gin.Context) {
	db, err := adapter.GetDBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	username := context.Param("username")

	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	context.JSON(http.StatusOK, user)
}

// User update handler
func UpdateUser(context *gin.Context) {
	db, err := adapter.GetDBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	var json struct {
		Name     *string `json:"name"`
		Username *string `json:"username"`
		Password *string `json:"password"`
	}

	if err := context.BindJSON(&json); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	username := context.Param("username")
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if json.Name != nil {
		user.Name = *json.Name
	}
	if json.Username != nil {
		user.Username = *json.Username
	}
	if json.Password != nil {
		user.Password = *json.Password
	}

	user.UpdatedAt = time.Now()
	if err := db.Save(&user).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update user"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// User deletion handler
func DeleteUser(context *gin.Context) {
	db, err := adapter.GetDBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	username := context.Param("username")

	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := db.Delete(&user).Error; err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
