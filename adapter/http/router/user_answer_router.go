package router

import (
	"phishing-quest/adapter/http/handler"

	"github.com/gin-gonic/gin"
)

// SetupUserAnswerRoutes registra as rotas de historico de respostas.
// authRequired exige usuario logado em todo o grupo.
func SetupUserAnswerRoutes(router *gin.Engine, userAnswerHandler *handler.UserAnswerHandler, authRequired gin.HandlerFunc) {
	userAnswersGroup := router.Group("api/v1/user-answers")
	userAnswersGroup.Use(authRequired)
	{
		userAnswersGroup.POST("", userAnswerHandler.CreateUserAnswer)
		userAnswersGroup.GET("", userAnswerHandler.ListUserAnswers)
		userAnswersGroup.GET("/:id", userAnswerHandler.GetUserAnswer)
		userAnswersGroup.PUT("/:id", userAnswerHandler.UpdateUserAnswer)
		userAnswersGroup.DELETE("/:id", userAnswerHandler.DeleteUserAnswer)
	}
}
