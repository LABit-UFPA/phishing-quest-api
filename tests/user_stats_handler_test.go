package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"phishing-quest/adapter/http/handler"
	"phishing-quest/adapter/repository"
	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserStatsHandler_GetMyStats_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserStatsRepository)
	uc := usecase.NewUserStatsUseCase(mockRepo)
	h := handler.NewUserStatsHandler(uc)

	userID := uuid.New()
	mockRepo.On("GetSignalDetectionAttempts", userID).Return([]repository.SignalDetectionAttempt{
		{ItemIsMalicious: true, Verdict: true},
	}, nil)
	mockRepo.On("GetCueOutcomes", userID).Return([]repository.CueAttemptOutcome{}, nil)

	r := gin.New()
	r.GET("/api/v1/me/stats", func(c *gin.Context) {
		c.Set("userId", userID)
		c.Next()
	}, h.GetMyStats)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats domain.UserStats
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &stats))
	assert.Equal(t, 1, stats.TotalAttempts)
}

// TestUserStatsHandler_GetMyStats_SemUserIdNoContexto garante que o
// endpoint nunca aceita userId vindo do cliente — so do JWT/contexto.
func TestUserStatsHandler_GetMyStats_SemUserIdNoContexto(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockUserStatsRepository)
	uc := usecase.NewUserStatsUseCase(mockRepo)
	h := handler.NewUserStatsHandler(uc)

	r := gin.New()
	r.GET("/api/v1/me/stats", h.GetMyStats) // sem middleware simulando auth

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockRepo.AssertNotCalled(t, "GetSignalDetectionAttempts", mock.Anything)
}
