package dto

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ConsentRequestDTO e o corpo de POST /api/v1/auth/consent. CohortId,
// ConsentVersion e DemographicsJSON sao opcionais na primeira versao
// do fluxo (o front pode enviar so userId + consentVersion); Condition
// nao vem do cliente — e atribuida pelo servidor (ver AttemptUseCase.
// RegisterConsent) para nao permitir que o participante escolha a
// propria condicao experimental.
type ConsentRequestDTO struct {
	UserID           uuid.UUID      `json:"userId" binding:"required"`
	ConsentVersion   string         `json:"consentVersion" binding:"required"`
	CohortID         *uuid.UUID     `json:"cohortId,omitempty"`
	DemographicsJSON datatypes.JSON `json:"demographicsJson,omitempty"`
}
