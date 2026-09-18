package router

import (
	"phishing-quest/adapter/http/handler"

	"github.com/gin-gonic/gin"
)

// SetupItemRoutes registra as rotas do modulo de itens de simulacao
// (email, sms, whatsapp, website, phone_call, pix_qr). Leitura e
// aberta; criacao/edicao serao restritas por role na issue #26.
// /items/next e registrada ANTES de /items/:id (mesmo o Gin resolvendo
// rota estatica antes de wildcard por padrao) para deixar explicito
// que "next" nunca deve ser interpretado como um id.
func SetupItemRoutes(router *gin.Engine, itemHandler *handler.ItemHandler, selectionHandler *handler.ItemSelectionHandler, authRequired gin.HandlerFunc) {
	itemGroup := router.Group("api/v1/items")
	{
		itemGroup.POST("", itemHandler.CreateItem)
		itemGroup.GET("", itemHandler.ListItems)
		itemGroup.GET("/next", authRequired, selectionHandler.NextItem)
		itemGroup.GET("/:id", itemHandler.GetItem)
	}
}
