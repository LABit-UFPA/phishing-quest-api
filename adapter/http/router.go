package http

import (
	"phishing-quest/adapter/http/middleware"
	"phishing-quest/adapter/http/router"
	"phishing-quest/container"
	"phishing-quest/domain"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cont *container.Container) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	authRequired := middleware.AuthRequired(cont.JWTService)
	requireResearcherRole := middleware.RequireRole(string(domain.RoleResearcher), string(domain.RoleAdmin))
	// Curadoria de conteudo (criar/revisar/publicar item) — issue #30.
	requireContentRole := middleware.RequireRole(string(domain.RoleAdmin), string(domain.RoleResearcher))

	router.SetupUserRoutes(r, cont.UserHandler)
	router.SetupCategoryRoutes(r, cont.CategoryHandler)
	router.SetupQuestionRoutes(r, cont.QuestionHandler)
	router.SetupAnswerRoutes(r, cont.AnswerHandler)
	router.SetupUserAnswerRoutes(r, cont.UserAnswerHandler, authRequired)
	router.SetupGameRoutes(r, cont.GameHandler, authRequired)
	router.SetupRankingRoutes(r, cont.RankingHandler)
	router.SetupItemRoutes(r, cont.ItemHandler, cont.ItemSelectionHandler, authRequired)
	router.SetupCueRoutes(r, cont.CueHandler)
	router.SetupAttemptRoutes(r, cont.AttemptHandler, authRequired)
	router.SetupTelemetryRoutes(r, cont.TelemetryHandler)
	router.SetupAssessmentRoutes(r, cont.AssessmentHandler, authRequired)
	router.SetupResearchExportRoutes(r, cont.ResearchExportHandler, authRequired, requireResearcherRole)
	router.SetupUserStatsRoutes(r, cont.UserStatsHandler, authRequired)
	router.SetupReviewScheduleRoutes(r, cont.ReviewScheduleHandler, authRequired)
	router.SetupItemReviewRoutes(r, cont.ItemReviewHandler, authRequired, requireContentRole)
	return r
}
