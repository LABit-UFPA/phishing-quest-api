package router

import (
	"phishing-quest/adapter/http/handler"

	"github.com/gin-gonic/gin"
)

// SetupItemRoutes registra as rotas PUBLICAS de leitura de itens de
// simulacao (email, sms, whatsapp, website, phone_call, pix_qr). Todas
// devolvem apenas itens PUBLICADOS.
//
// A criacao de itens saiu daqui na issue #30: o antigo
// POST /api/v1/items era anonimo e injetava item direto no jogo, sem
// revisao. Foi substituido por POST /api/v1/admin/items (rascunho) +
// fluxo de revisao/publicacao — ver SetupItemReviewRoutes.
//
// /items/next e registrada ANTES de /items/:id (mesmo o Gin resolvendo
// rota estatica antes de wildcard por padrao) para deixar explicito
// que "next" nunca deve ser interpretado como um id.
func SetupItemRoutes(router *gin.Engine, itemHandler *handler.ItemHandler, selectionHandler *handler.ItemSelectionHandler, authRequired gin.HandlerFunc) {
	itemGroup := router.Group("api/v1/items")
	{
		itemGroup.GET("", itemHandler.ListItems)
		itemGroup.GET("/next", authRequired, selectionHandler.NextItem)
		itemGroup.GET("/:id", itemHandler.GetItem)
	}
}
