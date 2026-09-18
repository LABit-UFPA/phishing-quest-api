package router

import (
	"phishing-quest/adapter/http/handler"

	"github.com/gin-gonic/gin"
)

// SetupUserRoutes configura as rotas relacionadas a usuários
func SetupUserRoutes(router *gin.Engine, userHandler *handler.UserHandler) {
	userGroup := router.Group("/api/v1/users")
	{
		userGroup.POST("/register", userHandler.CreateUser)
		userGroup.POST("/login", userHandler.Login)
		userGroup.GET("/:id", userHandler.GetUser)
	}
}
