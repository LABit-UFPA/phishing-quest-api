package repository

import (
	"phishing-quest/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ITelemetryEventRepository nao embute IRepository[T]: a unica
// operacao necessaria e ingestao em lote idempotente. Nao ha
// Update/Delete/GetByID previstos para telemetria (e so um log de
// eventos, nunca editado).
type ITelemetryEventRepository interface {
	CreateBatch(events []*domain.TelemetryEvent) error
}

type TelemetryEventRepository struct {
	db *gorm.DB
}

func NewTelemetryEventRepository(db *gorm.DB) ITelemetryEventRepository {
	return &TelemetryEventRepository{db: db}
}

// CreateBatch insere os eventos em lote com ON CONFLICT (id) DO
// NOTHING: como o id e gerado pelo cliente, reenviar o mesmo lote
// (fila offline apos falha de rede parcial) nao duplica eventos nem
// retorna erro de PK violada.
func (ter *TelemetryEventRepository) CreateBatch(events []*domain.TelemetryEvent) error {
	if len(events) == 0 {
		return nil
	}
	return ter.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&events).Error
}
