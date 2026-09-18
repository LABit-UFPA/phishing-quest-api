package domain

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// TelemetryEvent registra um evento de uso do app (tela aberta,
// notificacao de revisao respondida, erro de rede, etc). Complementa
// Attempt, que registra so a decisao sobre um item especifico.
//
// Id e gerado pelo CLIENTE (nao pelo servidor) de proposito: e o que
// permite a fila offline do app reenviar o mesmo lote apos falha de
// rede parcial sem duplicar eventos (ver TelemetryUseCase.IngestBatch).
type TelemetryEvent struct {
	Id          uuid.UUID      `json:"id" validate:"required" gorm:"primaryKey"`
	UserId      *uuid.UUID     `json:"userId,omitempty"`
	SessionId   *uuid.UUID     `json:"sessionId,omitempty"`
	EventType   string         `json:"eventType" validate:"required,min=1,max=64"`
	PayloadJSON datatypes.JSON `json:"payloadJson,omitempty" gorm:"column:payload_json"`
	CreatedAt   time.Time      `json:"createdAt,omitempty"`
}

func (te *TelemetryEvent) TableName() string {
	return "phishing_quest.telemetry_events"
}

func (te *TelemetryEvent) Validate() error {
	validate := validator.New()
	return validate.Struct(te)
}
