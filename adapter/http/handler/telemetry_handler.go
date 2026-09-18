package handler

import (
	"net/http"
	"phishing-quest/adapter/http/response"
	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/gin-gonic/gin"
)

type TelemetryHandler struct {
	telemetryUseCase *usecase.TelemetryUseCase
}

func NewTelemetryHandler(tuc *usecase.TelemetryUseCase) *TelemetryHandler {
	return &TelemetryHandler{telemetryUseCase: tuc}
}

// telemetryBatchRequest e o formato do corpo de POST /telemetry-events.
// Definido aqui (nao em dto/) porque dto ja e importado por domain
// (para os metodos ToDTO), e domain.TelemetryEvent importando de volta
// um DTO criaria um ciclo de import.
type telemetryBatchRequest struct {
	Events []*domain.TelemetryEvent `json:"events" binding:"required"`
}

// IngestBatch recebe um lote de eventos de telemetria (suporta a fila
// offline do app). Idempotente: reenviar o mesmo lote apos falha
// parcial de rede nao duplica eventos.
func (th *TelemetryHandler) IngestBatch(c *gin.Context) {
	var batch telemetryBatchRequest
	if err := c.ShouldBindJSON(&batch); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	if err := th.telemetryUseCase.IngestBatch(batch.Events); err != nil {
		if err == usecase.ErrEmptyBatch {
			c.JSON(http.StatusBadRequest, response.Error(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"ingested": len(batch.Events)})
}
