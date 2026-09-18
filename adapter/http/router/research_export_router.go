package router

import (
	"github.com/gin-gonic/gin"
	"phishing-quest/adapter/http/handler"
)

// SetupResearchExportRoutes registra a rota de export de dados de
// pesquisa. authRequired + requireResearcherRole devem ser aplicados
// nessa ordem: primeiro autentica (preenche o role no contexto),
// depois autoriza.
func SetupResearchExportRoutes(router *gin.Engine, exportHandler *handler.ResearchExportHandler, authRequired, requireResearcherRole gin.HandlerFunc) {
	router.GET("api/v1/research/export", authRequired, requireResearcherRole, exportHandler.ExportCSV)
}
