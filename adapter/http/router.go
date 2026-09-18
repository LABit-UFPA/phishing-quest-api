package http

import (
	"phishing-quest/adapter/http/middleware"
	"phishing-quest/adapter/http/router"
	"phishing-quest/container"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cont *container.Container) *gin.Engine {
	r := gin.Default()

	authRequired := middleware.AuthRequired(cont.JWTService)

	router.SetupUserRoutes(r, cont.UserHandler)
	router.SetupCategoryRoutes(r, cont.CategoryHandler)
	router.SetupQuestionRoutes(r, cont.QuestionHandler)
	router.SetupAnswerRoutes(r, cont.AnswerHandler)
	router.SetupUserAnswerRoutes(r, cont.UserAnswerHandler, authRequired)
	router.SetupGameRoutes(r, cont.GameHandler, authRequired)
	router.SetupRankingRoutes(r, cont.RankingHandler)
	return r
}
