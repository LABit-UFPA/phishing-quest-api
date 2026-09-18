package router

import (
	"github.com/gin-gonic/gin"
	"phishing-quest/adapter/http/handler"
)

// SetupItemRoutes registra as rotas do modulo de itens de simulacao
// (email, sms, whatsapp, website, phone_call, pix_qr). Leitura e
// aberta; criacao/edicao serao restritas por role na issue #26.
func SetupItemRoutes(router *gin.Engine, itemHandler *handler.ItemHandler) {
	itemGroup := router.Group("api/v1/items")
	{
		itemGroup.POST("", itemHandler.CreateItem)
		itemGroup.GET("", itemHandler.ListItems)
		itemGroup.GET("/:id", itemHandler.GetItem)
	}
}
