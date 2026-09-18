package domain

import (
	"time"

	"github.com/google/uuid"
)

// StudyParticipant e o registro minimo de consentimento do usuario
// para participar da coleta de dados de pesquisa (POST /attempts
// exige um registro aqui antes de aceitar qualquer tentativa).
//
// Campos adicionais do desenho experimental (cohort_id, condition,
// consent_version, demographics_json, withdrawn_at) sao adicionados
// pela issue #25, que expande esta tabela via ALTER TABLE.
type StudyParticipant struct {
	UserId      uuid.UUID `json:"userId" gorm:"primaryKey"`
	ConsentedAt time.Time `json:"consentedAt"`
}

func (sp *StudyParticipant) TableName() string {
	return "phishing_quest.study_participants"
}
