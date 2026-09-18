package usecase

import (
	"errors"
	"phishing-quest/adapter/repository"
	"phishing-quest/domain"
)

var ErrEmptyBatch = errors.New("lote de eventos vazio")

type TelemetryUseCase struct {
	telemetryRepo repository.ITelemetryEventRepository
}

func NewTelemetryUseCase(telemetryRepo repository.ITelemetryEventRepository) *TelemetryUseCase {
	return &TelemetryUseCase{telemetryRepo: telemetryRepo}
}

// IngestBatch valida cada evento do lote (id e eventType obrigatorios)
// antes de persistir. A idempotencia (reenvio seguro) e garantida pelo
// repositorio via ON CONFLICT DO NOTHING sobre o id gerado pelo
// cliente.
func (tuc *TelemetryUseCase) IngestBatch(events []*domain.TelemetryEvent) error {
	if len(events) == 0 {
		return ErrEmptyBatch
	}

	for _, event := range events {
		if err := event.Validate(); err != nil {
			return err
		}
	}

	return tuc.telemetryRepo.CreateBatch(events)
}
