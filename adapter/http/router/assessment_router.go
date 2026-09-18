package router

import (
	"github.com/gin-gonic/gin"
	"phishing-quest/adapter/http/handler"
)

// SetupAssessmentRoutes registra as rotas dos instrumentos de
// avaliacao (pre/pos/pos-tardio de 4 semanas).
func SetupAssessmentRoutes(router *gin.Engine, assessmentHandler *handler.AssessmentHandler, authRequired gin.HandlerFunc) {
	assessmentGroup := router.Group("api/v1/assessments")
	assessmentGroup.Use(authRequired)
	{
		assessmentGroup.POST("/:phase", assessmentHandler.SubmitAssessment)
		assessmentGroup.GET("/:phase/users/:id", assessmentHandler.GetAssessment)
	}
}
