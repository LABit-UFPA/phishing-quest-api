package handler

import (
	"net/http"
	"phishing-quest/adapter/http/response"
	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AssessmentHandler struct {
	assessmentUseCase *usecase.AssessmentUseCase
}

func NewAssessmentHandler(auc *usecase.AssessmentUseCase) *AssessmentHandler {
	return &AssessmentHandler{assessmentUseCase: auc}
}

// SubmitAssessment registra as respostas de um instrumento na fase
// informada pela rota (POST /api/v1/assessments/:phase). A fase do
// path tem precedencia sobre qualquer "phase" enviado no corpo, para
// a rota ser a fonte de verdade do fluxo (pre -> post -> delayed_4w).
func (ah *AssessmentHandler) SubmitAssessment(c *gin.Context) {
	phaseParam := domain.AssessmentPhase(c.Param("phase"))

	var assessmentDTO *domain.Assessment
	if err := c.ShouldBindJSON(&assessmentDTO); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	assessmentDTO.Phase = phaseParam

	created, err := ah.assessmentUseCase.SubmitAssessment(assessmentDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, created)
}

// GetAssessment busca a resposta de um usuario numa fase especifica
// (GET /api/v1/assessments/:phase/users/:id).
func (ah *AssessmentHandler) GetAssessment(c *gin.Context) {
	phaseParam := domain.AssessmentPhase(c.Param("phase"))

	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.ErrInvalidIDFormat))
		return
	}

	assessment, err := ah.assessmentUseCase.GetAssessment(userID, phaseParam)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Assessment not found"))
		return
	}

	c.JSON(http.StatusOK, assessment)
}
