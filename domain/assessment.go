package domain

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// AssessmentPhase enumera as fases de coleta do desenho experimental
// pre/pos/pos-tardio (ROADMAP_PESQUISA_2027.md, Fase 5 e 7).
type AssessmentPhase string

const (
	PhasePre       AssessmentPhase = "pre"
	PhasePost      AssessmentPhase = "post"
	PhaseDelayed4w AssessmentPhase = "delayed_4w"
)

// Assessment registra as respostas de um instrumento de avaliacao
// (teste de deteccao, escalas de confianca/autoeficacia, SUS, IMI)
// numa fase especifica do estudo.
type Assessment struct {
	Id                uuid.UUID       `json:"id" gorm:"primaryKey"`
	UserId            uuid.UUID       `json:"userId" validate:"required"`
	Phase             AssessmentPhase `json:"phase" validate:"required,oneof=pre post delayed_4w"`
	InstrumentVersion string          `json:"instrumentVersion" validate:"required"`
	StartedAt         *time.Time      `json:"startedAt,omitempty"`
	FinishedAt        *time.Time      `json:"finishedAt,omitempty"`
	ResponsesJSON     datatypes.JSON  `json:"responsesJson,omitempty" gorm:"column:responses_json"`
	CreatedAt         time.Time       `json:"createdAt,omitempty"`
}

func (a *Assessment) TableName() string {
	return "phishing_quest.assessments"
}

func (a *Assessment) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}
