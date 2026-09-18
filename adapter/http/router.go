package http

import (
	"phishing-quest/adapter/http/middleware"
	"phishing-quest/adapter/http/router"
	"phishing-quest/container"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cont *container.Container) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	authRequired := middleware.AuthRequired(cont.JWTService)

	router.SetupUserRoutes(r, cont.UserHandler)
	router.SetupCategoryRoutes(r, cont.CategoryHandler)
	router.SetupQuestionRoutes(r, cont.QuestionHandler)
	router.SetupAnswerRoutes(r, cont.AnswerHandler)
	router.SetupUserAnswerRoutes(r, cont.UserAnswerHandler, authRequired)
	router.SetupGameRoutes(r, cont.GameHandler, authRequired)
	router.SetupRankingRoutes(r, cont.RankingHandler)
	router.SetupItemRoutes(r, cont.ItemHandler)
	router.SetupCueRoutes(r, cont.CueHandler)
	router.SetupAttemptRoutes(r, cont.AttemptHandler, authRequired)
	router.SetupTelemetryRoutes(r, cont.TelemetryHandler)
	router.SetupAssessmentRoutes(r, cont.AssessmentHandler, authRequired)
	return r
}
