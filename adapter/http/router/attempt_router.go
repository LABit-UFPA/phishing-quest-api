package router

import (
	"github.com/gin-gonic/gin"
	"phishing-quest/adapter/http/handler"
)

// SetupAttemptRoutes registra as rotas de consentimento e de tentativas
// (dado central do estudo de retencao).
func SetupAttemptRoutes(router *gin.Engine, attemptHandler *handler.AttemptHandler, authRequired gin.HandlerFunc) {
	router.POST("api/v1/auth/consent", attemptHandler.RegisterConsent)

	attemptGroup := router.Group("api/v1/attempts")
	attemptGroup.Use(authRequired)
	{
		attemptGroup.POST("", attemptHandler.CreateAttempt)
		attemptGroup.GET("/users/:id", attemptHandler.ListAttemptsByUser)
	}
}
