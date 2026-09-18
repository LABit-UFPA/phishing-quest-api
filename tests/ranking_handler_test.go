package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"phishing-quest/adapter/http/handler"
	"phishing-quest/core/usecase"
	"phishing-quest/dto"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRankingHandler_GetGlobalRanking_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRankingRepository)
	uc := usecase.NewRankingUseCase(mockRepo)
	h := handler.NewRankingHandler(uc)

	mockRepo.On("GetGlobalRanking", 10, 0).Return([]dto.RankingEntryDTO{
		{Position: 1, UserId: uuid.New(), Username: "topo", TotalScore: 99},
	}, nil)

	r := gin.New()
	r.GET("/api/v1/rankings", h.GetGlobalRanking)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rankings", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string][]dto.RankingEntryDTO
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Len(t, response["ranking"], 1)
	assert.Equal(t, 99, response["ranking"][0].TotalScore)
	mockRepo.AssertExpectations(t)
}

// TestRankingHandler_GetGlobalRanking_ComCohortId_ChamaRepoDeCoorte
// garante que ?cohortId= desvia para GetRankingByCohort em vez do
// ranking global.
func TestRankingHandler_GetGlobalRanking_ComCohortId_ChamaRepoDeCoorte(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRankingRepository)
	uc := usecase.NewRankingUseCase(mockRepo)
	h := handler.NewRankingHandler(uc)

	cohortID := uuid.New()
	mockRepo.On("GetRankingByCohort", cohortID, 10, 0).Return([]dto.RankingEntryDTO{}, nil)

	r := gin.New()
	r.GET("/api/v1/rankings", h.GetGlobalRanking)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rankings?cohortId="+cohortID.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetGlobalRanking")
}

func TestRankingHandler_GetGlobalRanking_CohortIdInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRankingRepository)
	uc := usecase.NewRankingUseCase(mockRepo)
	h := handler.NewRankingHandler(uc)

	r := gin.New()
	r.GET("/api/v1/rankings", h.GetGlobalRanking)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/rankings?cohortId=not-a-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockRepo.AssertNotCalled(t, "GetRankingByCohort")
}
