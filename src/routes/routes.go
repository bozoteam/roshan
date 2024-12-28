package routes

import (
	"github.com/bozoteam/roshan/src/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	router.POST("/users", func(c *gin.Context) { controllers.CreateUser(c) })
	router.GET("/users/:username", func(c *gin.Context) { controllers.FindUser(c) })
	router.PUT("/users/:username", func(c *gin.Context) { controllers.UpdateUser(c) })
	router.DELETE("/users/:username", func(c *gin.Context) { controllers.DeleteUser(c) })

	router.POST("/auth", controllers.AuthenticateUser)

	return router
}
