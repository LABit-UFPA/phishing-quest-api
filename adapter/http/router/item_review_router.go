package router

import (
	"phishing-quest/adapter/http/handler"

	"github.com/gin-gonic/gin"
)

// SetupItemReviewRoutes registra o pipeline administrativo de curadoria
// de itens (issue #30). Todas as rotas exigem autenticacao E role de
// conteudo (admin/researcher): criar, gerar, revisar e publicar item
// nunca podem ser acoes anonimas, porque item errado ensina errado.
func SetupItemReviewRoutes(
	router *gin.Engine,
	reviewHandler *handler.ItemReviewHandler,
	authRequired gin.HandlerFunc,
	requireContentRole gin.HandlerFunc,
) {
	adminItems := router.Group("api/v1/admin/items", authRequired, requireContentRole)
	{
		adminItems.POST("", reviewHandler.CreateDraft)
		adminItems.POST("/generate", reviewHandler.GenerateDraft)
		adminItems.GET("", reviewHandler.ListByStatus)
		adminItems.GET("/:id", reviewHandler.GetItem)
		adminItems.POST("/:id/review", reviewHandler.Review)
		adminItems.POST("/:id/publish", reviewHandler.Publish)
	}
}
