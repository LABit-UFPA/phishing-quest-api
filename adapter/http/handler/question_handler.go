package handler

import (
	"net/http"
	"phishing-quest/adapter/http/response"
	"phishing-quest/core/usecase"
	"phishing-quest/domain"
	"phishing-quest/dto"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type QuestionHandler struct {
	questionUseCase *usecase.QuestionUseCase
}

func NewQuestionHandler(quc *usecase.QuestionUseCase) *QuestionHandler {
	return &QuestionHandler{questionUseCase: quc}
}

func (qh *QuestionHandler) CreateQuestion(c *gin.Context) {
	var questionDTO *domain.Question
	if err := c.ShouldBindJSON(&questionDTO); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	createdQuestion, err := qh.questionUseCase.CreateQuestion(questionDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, createdQuestion)
}

func (qh *QuestionHandler) ListAnswersByQuestion(c *gin.Context) {
	idParam := c.Param("id")
	questionID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid Question ID"))
		return
	}

	answers, err := qh.questionUseCase.GetAnswersByQuestionID(questionID)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Answers not found"))
		return
	}

	var answerDTOs []*dto.AnswerDTO
	for _, answer := range answers {
		answerDTOs = append(answerDTOs, answer.ToDTO())
	}

	responseBody := dto.QuestionAnswersDTO{
		QuestionId: questionID,
		Answers:    answerDTOs,
	}

	c.JSON(http.StatusOK, responseBody)
}

func (qh *QuestionHandler) GetQuestion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.ErrInvalidIDFormat))
		return
	}

	question, err := qh.questionUseCase.GetQuestion(id)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Question not found"))
		return
	}

	c.JSON(http.StatusOK, question)
}

func (qh *QuestionHandler) UpdateQuestion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.ErrInvalidIDFormat))
		return
	}

	var questionDTO *domain.Question
	if err := c.ShouldBindJSON(&questionDTO); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	updatedQuestion, err := qh.questionUseCase.UpdateQuestion(id, questionDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, updatedQuestion)
}

func (qh *QuestionHandler) DeleteQuestion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.ErrInvalidIDFormat))
		return
	}

	err = qh.questionUseCase.DeleteQuestion(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
