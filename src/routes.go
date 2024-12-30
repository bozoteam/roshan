package routes

import (
	"fmt"
	"log"
	"os"

	authRouter "github.com/bozoteam/roshan/src/modules/auth/routes"
	userRouter "github.com/bozoteam/roshan/src/modules/user/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// RegisterRoutes registers all routes
func RegisterRoutes() *gin.Engine {
	router := gin.Default()
	userRouter.RegisterUserRoutes(router)
	authRouter.RegisterAuthRoutes(router)
	return router
}

// RunServer starts the API server
func RunServer() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}
	port := os.Getenv("API_PORT")
	router := RegisterRoutes()
	if port == "" {
		fmt.Printf("API_PORT is not set. Defaulting to 8080\n")
		port = "8080"
	}
	router.Run(":" + port)
}
