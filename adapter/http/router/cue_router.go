package router

import (
	"github.com/gin-gonic/gin"
	"phishing-quest/adapter/http/handler"
)

// SetupCueRoutes registra as rotas da taxonomia de pistas e sua
// associacao com items.
func SetupCueRoutes(router *gin.Engine, cueHandler *handler.CueHandler) {
	cueGroup := router.Group("api/v1/cues")
	{
		cueGroup.GET("", cueHandler.ListCues)
		cueGroup.POST("", cueHandler.CreateCue)
	}

	router.POST("api/v1/item-cues", cueHandler.AssociateItemCue)
	router.GET("api/v1/items/:id/cues", cueHandler.GetCuesByItem)
}
