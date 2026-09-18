package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"phishing-quest/adapter/http/handler"
	"phishing-quest/core/usecase"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTelemetryHandler_IngestBatch_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockTelemetryEventRepository)
	uc := usecase.NewTelemetryUseCase(mockRepo)
	h := handler.NewTelemetryHandler(uc)

	mockRepo.On("CreateBatch", mock.Anything).Return(nil)

	r := gin.New()
	r.POST("/api/v1/telemetry-events", h.IngestBatch)

	body := map[string]interface{}{
		"events": []map[string]interface{}{
			{"id": "11111111-1111-1111-1111-111111111111", "eventType": "screen_opened"},
			{"id": "22222222-2222-2222-2222-222222222222", "eventType": "review_notification_answered"},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry-events", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.Equal(t, float64(2), response["ingested"])

	mockRepo.AssertExpectations(t)
}

func TestTelemetryHandler_IngestBatch_LoteVazioRetorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockTelemetryEventRepository)
	uc := usecase.NewTelemetryUseCase(mockRepo)
	h := handler.NewTelemetryHandler(uc)

	r := gin.New()
	r.POST("/api/v1/telemetry-events", h.IngestBatch)

	body := map[string]interface{}{"events": []map[string]interface{}{}}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry-events", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockRepo.AssertNotCalled(t, "CreateBatch", mock.Anything)
}

func TestTelemetryHandler_IngestBatch_CampoEventsAusenteRetorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockTelemetryEventRepository)
	uc := usecase.NewTelemetryUseCase(mockRepo)
	h := handler.NewTelemetryHandler(uc)

	r := gin.New()
	r.POST("/api/v1/telemetry-events", h.IngestBatch)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry-events", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
