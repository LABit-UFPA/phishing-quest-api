package handler

import (
	"net/http"
	"phishing-quest/adapter/http/response"
	"phishing-quest/core/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserStatsHandler struct {
	statsUseCase *usecase.UserStatsUseCase
}

func NewUserStatsHandler(suc *usecase.UserStatsUseCase) *UserStatsHandler {
	return &UserStatsHandler{statsUseCase: suc}
}

// GetMyStats retorna d', criterio c, taxas de acerto/falso alarme e o
// desempenho por pista do usuario autenticado (GET /api/v1/me/stats).
// userId vem sempre do JWT (contexto) — nao ha como consultar as
// estatisticas de outro usuario por este endpoint.
func (ush *UserStatsHandler) GetMyStats(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	userID, ok := userIDValue.(uuid.UUID)
	if !exists || !ok {
		c.JSON(http.StatusUnauthorized, response.Error("usuario nao autenticado"))
		return
	}

	stats, err := ush.statsUseCase.GetStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, stats)
}
