package handler

import (
	"net/http"
	"phishing-quest/adapter/http/response"
	"phishing-quest/core/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReviewScheduleHandler struct {
	scheduleUseCase *usecase.ReviewScheduleUseCase
}

func NewReviewScheduleHandler(suc *usecase.ReviewScheduleUseCase) *ReviewScheduleHandler {
	return &ReviewScheduleHandler{scheduleUseCase: suc}
}

// GetDueReviews retorna os agendamentos de revisao espacada devidos
// para o usuario autenticado (GET /api/v1/review/due?limit=).
// userId vem sempre do JWT (contexto).
func (rsh *ReviewScheduleHandler) GetDueReviews(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(uuid.UUID)
	if !exists || !ok {
		c.JSON(http.StatusUnauthorized, response.Error("usuario nao autenticado"))
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	due, err := rsh.scheduleUseCase.GetDueReviews(userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"due": due})
}
