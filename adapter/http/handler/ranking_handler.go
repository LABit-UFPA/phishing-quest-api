package handler

import (
	"net/http"
	"phishing-quest/adapter/http/response"
	"phishing-quest/core/usecase"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RankingHandler struct {
	rankingUseCase *usecase.RankingUseCase
}

func NewRankingHandler(ruc *usecase.RankingUseCase) *RankingHandler {
	return &RankingHandler{rankingUseCase: ruc}
}

// GetGlobalRanking retorna o ranking global, ou restrito a uma
// coorte/turma se ?cohortId= for informado (ROADMAP_PESQUISA_2027.md:
// leaderboard global desmotiva quem esta embaixo; comparacao por
// coorte funciona melhor).
func (rh *RankingHandler) GetGlobalRanking(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	cohortParam := c.Query("cohortId")
	if cohortParam != "" {
		cohortID, err := uuid.Parse(cohortParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.Error("cohortId invalido"))
			return
		}

		ranking, err := rh.rankingUseCase.GetCohortRanking(cohortID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Error("Failed to fetch ranking"))
			return
		}
		c.JSON(http.StatusOK, gin.H{"ranking": ranking})
		return
	}

	ranking, err := rh.rankingUseCase.GetGlobalRanking(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to fetch ranking"))
		return
	}

	c.JSON(http.StatusOK, gin.H{"ranking": ranking})
}
