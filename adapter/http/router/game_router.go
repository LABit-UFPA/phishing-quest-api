package router

import (
	"phishing-quest/adapter/http/handler"

	"github.com/gin-gonic/gin"
)

// SetupGameRoutes registra as rotas do modulo de jogo. authRequired e o
// middleware de autenticacao (adapter/http/middleware.AuthRequired) —
// submeter resposta exige usuario logado.
func SetupGameRoutes(router *gin.Engine, gameHandler *handler.GameHandler, authRequired gin.HandlerFunc) {
	gameGroup := router.Group("api/v1/game")
	gameGroup.Use(authRequired)
	{
		gameGroup.POST("/answer", gameHandler.SubmitAnswer)
	}
}
