package router

import (
	"github.com/gin-gonic/gin"
	"phishing-quest/adapter/http/handler"
)

// SetupReviewScheduleRoutes registra a rota da fila de revisao
// espacada (Leitner por pista).
func SetupReviewScheduleRoutes(router *gin.Engine, reviewHandler *handler.ReviewScheduleHandler, authRequired gin.HandlerFunc) {
	router.GET("api/v1/review/due", authRequired, reviewHandler.GetDueReviews)
}
