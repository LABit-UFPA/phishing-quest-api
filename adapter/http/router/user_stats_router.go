package router

import (
	"github.com/gin-gonic/gin"
	"phishing-quest/adapter/http/handler"
)

// SetupUserStatsRoutes registra a rota de estatisticas do usuario
// autenticado (deteccao de sinal + desempenho por pista).
func SetupUserStatsRoutes(router *gin.Engine, statsHandler *handler.UserStatsHandler, authRequired gin.HandlerFunc) {
	router.GET("api/v1/me/stats", authRequired, statsHandler.GetMyStats)
}
