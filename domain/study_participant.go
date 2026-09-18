package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// StudyParticipant registra o consentimento do usuario para participar
// da coleta de dados de pesquisa (POST /attempts exige um registro
// aqui antes de aceitar qualquer tentativa) e os campos do desenho
// experimental atribuidos nesse momento: coorte, condicao, versao do
// TCLE aceito, demografia autoinformada, e retirada de consentimento.
type StudyParticipant struct {
	UserId           uuid.UUID      `json:"userId" gorm:"primaryKey"`
	ConsentedAt      time.Time      `json:"consentedAt"`
	CohortId         *uuid.UUID     `json:"cohortId,omitempty"`
	Condition        string         `json:"condition,omitempty"`
	ConsentVersion   string         `json:"consentVersion,omitempty"`
	DemographicsJSON datatypes.JSON `json:"demographicsJson,omitempty" gorm:"column:demographics_json"`
	WithdrawnAt      *time.Time     `json:"withdrawnAt,omitempty"`
}

func (sp *StudyParticipant) TableName() string {
	return "phishing_quest.study_participants"
}

// IsActive indica se o participante ainda esta ativo no estudo (nao
// retirou o consentimento). Usado para bloquear novas tentativas de
// participantes que exerceram o direito de exclusao (LGPD).
func (sp *StudyParticipant) IsActive() bool {
	return sp.WithdrawnAt == nil
}
