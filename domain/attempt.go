package domain

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// AttemptAction enumera as acoes que o usuario pode escolher ao
// avaliar um item — substitui o veredito binario simples por um
// repertorio de acao mais proximo do comportamento real
// (ROADMAP_PESQUISA_2027.md, Fase 3).
type AttemptAction string

const (
	ActionReport            AttemptAction = "report"
	ActionDelete            AttemptAction = "delete"
	ActionVerifyOtherChanel AttemptAction = "verify_other_channel"
	ActionReply             AttemptAction = "reply"
	ActionClick             AttemptAction = "click"
	ActionIgnore            AttemptAction = "ignore"
)

// Attempt e o dado central do artigo: registra a decisao completa do
// usuario sobre um item — veredito, acao, confianca, justificativa,
// avaliacao formativa por IA, latencia de decisao e se houve clique
// em link malicioso.
type Attempt struct {
	Id              uuid.UUID      `json:"id" gorm:"primaryKey"`
	UserId          uuid.UUID      `json:"userId" validate:"required"`
	ItemId          uuid.UUID      `json:"itemId" validate:"required"`
	SessionId       uuid.UUID      `json:"sessionId" validate:"required"`
	Condition       string         `json:"condition,omitempty"`
	Verdict         *bool          `json:"verdict,omitempty"`
	Action          AttemptAction  `json:"action" validate:"required,oneof=report delete verify_other_channel reply click ignore"`
	Confidence      *int           `json:"confidence,omitempty" validate:"omitempty,gte=1,lte=5"`
	Justification   string         `json:"justification,omitempty"`
	LlmRating       *int           `json:"llmRating,omitempty"`
	LlmFeedbackJSON datatypes.JSON `json:"llmFeedbackJson,omitempty" gorm:"column:llm_feedback_json"`
	IsCorrect       *bool          `json:"isCorrect,omitempty"`
	LatencyMs       *int           `json:"latencyMs,omitempty"`
	ClickedLink     bool           `json:"clickedLink"`
	CreatedAt       time.Time      `json:"createdAt,omitempty"`
}

func (a *Attempt) TableName() string {
	return "phishing_quest.attempts"
}

func (a *Attempt) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}
