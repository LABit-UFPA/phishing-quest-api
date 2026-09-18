package handler

import (
	"net/http"
	"phishing-quest/adapter/http/response"
	"phishing-quest/core/usecase"
	"phishing-quest/dto"

	"github.com/gin-gonic/gin"
)

type GameHandler struct {
	gameUseCase *usecase.GameUseCase
}

func NewGameHandler(guc *usecase.GameUseCase) *GameHandler {
	return &GameHandler{gameUseCase: guc}
}

func (gh *GameHandler) SubmitAnswer(c *gin.Context) {
	var answerDTO dto.SubmitAnswerDTO
	if err := c.ShouldBindJSON(&answerDTO); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	result, err := gh.gameUseCase.ProcessAnswer(&answerDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, result)
}
