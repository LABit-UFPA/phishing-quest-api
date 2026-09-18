package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SignalDetectionAttempt e o dado minimo necessario para calcular
// teoria de deteccao de sinal: se o item era malicioso de fato, e se
// o usuario disse que era (verdict). Attempts sem verdict (ex.: acao
// sem julgamento binario) sao ignorados pela query.
type SignalDetectionAttempt struct {
	ItemIsMalicious bool
	Verdict         bool
}

// CueAttemptOutcome e uma linha de attempt associada a uma pista do
// item respondido, com o resultado (correto ou nao).
type CueAttemptOutcome struct {
	CueCode   string
	LabelPt   string
	IsCorrect bool
}

// CueMastery agrega o desempenho do usuario por pista JA AGREGADO no
// banco (um registro por pista), incluindo o id da pista — que o
// CueAttemptOutcome nao expoe. O id e o que permite filtrar items que
// contenham as pistas fracas na selecao adaptativa (issue #28).
type CueMastery struct {
	CueId    uuid.UUID
	CueCode  string
	Answered int
	Correct  int
}

type IUserStatsRepository interface {
	GetSignalDetectionAttempts(userID uuid.UUID) ([]SignalDetectionAttempt, error)
	GetCueOutcomes(userID uuid.UUID) ([]CueAttemptOutcome, error)
	// GetCueMastery agrega acertos/respostas por pista para o usuario,
	// usado pela selecao adaptativa para decidir quais pistas ainda
	// nao foram dominadas.
	GetCueMastery(userID uuid.UUID) ([]CueMastery, error)
}

type UserStatsRepository struct {
	db *gorm.DB
}

func NewUserStatsRepository(db *gorm.DB) IUserStatsRepository {
	return &UserStatsRepository{db: db}
}

// GetSignalDetectionAttempts busca todas as tentativas do usuario que
// tem veredito binario registrado (attempts.verdict IS NOT NULL),
// junto com se o item era de fato malicioso.
func (usr *UserStatsRepository) GetSignalDetectionAttempts(userID uuid.UUID) ([]SignalDetectionAttempt, error) {
	var rows []SignalDetectionAttempt
	err := usr.db.Table("phishing_quest.attempts AS a").
		Select("i.is_malicious AS item_is_malicious, a.verdict AS verdict").
		Joins("JOIN phishing_quest.items i ON i.id = a.item_id").
		Where("a.user_id = ? AND a.verdict IS NOT NULL", userID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// GetCueMastery agrega, por pista, quantas tentativas o usuario
// respondeu e quantas acertou. A agregacao e feita no banco (GROUP BY
// + FILTER) em vez de trazer todas as linhas para a aplicacao, porque
// esta query roda a cada chamada de GET /items/next no modo adaptativo
// — e o volume de attempts cresce por sessao.
func (usr *UserStatsRepository) GetCueMastery(userID uuid.UUID) ([]CueMastery, error) {
	var rows []CueMastery
	err := usr.db.Table("phishing_quest.attempts AS a").
		Select(`c.id AS cue_id,
			c.code AS cue_code,
			COUNT(*) AS answered,
			COUNT(*) FILTER (WHERE a.is_correct) AS correct`).
		Joins("JOIN phishing_quest.item_cues ic ON ic.item_id = a.item_id").
		Joins("JOIN phishing_quest.cues c ON c.id = ic.cue_id").
		Where("a.user_id = ? AND a.is_correct IS NOT NULL", userID).
		Group("c.id, c.code").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// GetCueOutcomes junta cada tentativa do usuario com as pistas do
// item respondido, via item_cues, para permitir o diagnostico "usuario
// erra sistematicamente quando a pista X esta presente".
func (usr *UserStatsRepository) GetCueOutcomes(userID uuid.UUID) ([]CueAttemptOutcome, error) {
	var rows []CueAttemptOutcome
	err := usr.db.Table("phishing_quest.attempts AS a").
		Select("c.code AS cue_code, c.label_pt AS label_pt, a.is_correct AS is_correct").
		Joins("JOIN phishing_quest.item_cues ic ON ic.item_id = a.item_id").
		Joins("JOIN phishing_quest.cues c ON c.id = ic.cue_id").
		Where("a.user_id = ? AND a.is_correct IS NOT NULL", userID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
