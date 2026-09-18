package tests

import (
	"testing"
	"time"

	"phishing-quest/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestAnswer_Validate_AceitaIsCorrectFalso e a regressao da issue #17:
// Answer.IsCorrect tinha a tag validate:"required", e o pacote
// go-playground/validator trata bool false como valor-zero, reprovando
// a validacao. Isso impedia criar respostas incorretas via API — metade
// de qualquer conjunto de respostas plausivel.
func TestAnswer_Validate_AceitaIsCorrectFalso(t *testing.T) {
	answer := &domain.Answer{
		Id:         uuid.New(),
		QuestionId: uuid.New(),
		AnswerText: "Resposta errada de proposito",
		IsCorrect:  false,
	}

	err := answer.Validate()

	assert.NoError(t, err)
}

func TestAnswer_Validate_AceitaIsCorrectVerdadeiro(t *testing.T) {
	answer := &domain.Answer{
		Id:         uuid.New(),
		QuestionId: uuid.New(),
		AnswerText: "Resposta certa",
		IsCorrect:  true,
	}

	err := answer.Validate()

	assert.NoError(t, err)
}

func TestAnswer_Validate_RejeitaTextoVazio(t *testing.T) {
	answer := &domain.Answer{
		Id:         uuid.New(),
		QuestionId: uuid.New(),
		AnswerText: "",
		IsCorrect:  false,
	}

	err := answer.Validate()

	assert.Error(t, err)
}

// TestQuestion_Validate_NaoExigeMaisCorrectAnswer e a regressao da
// segunda parte da issue #17: a coluna correct_answer foi removida da
// tabela questions pela migration V20250104153937, mas o struct
// domain.Question ainda tinha o campo CorrectAnswer com
// validate:"required", fazendo toda validacao falhar e, se corrigida
// manualmente, o GORM tentaria gravar uma coluna inexistente.
func TestQuestion_Validate_NaoExigeMaisCorrectAnswer(t *testing.T) {
	question := &domain.Question{
		Id:           uuid.New(),
		CategoryId:   uuid.New(),
		QuestionText: "Este email e phishing?",
	}

	err := question.Validate()

	assert.NoError(t, err)
}

func TestQuestion_Validate_AindaExigeCategoryIdEQuestionText(t *testing.T) {
	question := &domain.Question{}

	err := question.Validate()

	assert.Error(t, err)
}

// TestReviewSchedule_ApplyResult_AcertoAvancaCaixa garante o
// comportamento central do algoritmo Leitner (issue #27): uma
// resposta correta avanca a pista para a proxima caixa, aumentando o
// intervalo de revisao (regressao espacada).
func TestReviewSchedule_ApplyResult_AcertoAvancaCaixa(t *testing.T) {
	rs := &domain.ReviewSchedule{Box: domain.MinLeitnerBox}
	now := time.Now()

	rs.ApplyResult(true, now)

	assert.Equal(t, 2, rs.Box)
	assert.NotNil(t, rs.LastResult)
	assert.True(t, *rs.LastResult)
	// Caixa 2 tem intervalo de 1 dia (domain.LeitnerIntervalDays).
	assert.WithinDuration(t, now.AddDate(0, 0, 1), rs.DueAt, time.Second)
}

// TestReviewSchedule_ApplyResult_ErroVoltaParaCaixaUm garante que um
// erro sempre reseta a pista para a caixa 1 (revisao quase imediata),
// independentemente de quao avancada ela estava — e o criterio de
// aceite explicito da issue #27: "item errado reaparece antes".
func TestReviewSchedule_ApplyResult_ErroVoltaParaCaixaUm(t *testing.T) {
	rs := &domain.ReviewSchedule{Box: 4}
	now := time.Now()

	rs.ApplyResult(false, now)

	assert.Equal(t, domain.MinLeitnerBox, rs.Box)
	assert.NotNil(t, rs.LastResult)
	assert.False(t, *rs.LastResult)
	// Caixa 1 tem intervalo 0 (devido imediatamente).
	assert.WithinDuration(t, now, rs.DueAt, time.Second)
}

// TestReviewSchedule_ApplyResult_NaoPassaDaCaixaMaxima garante que a
// caixa nunca excede domain.MaxLeitnerBox mesmo apos varios acertos
// consecutivos.
func TestReviewSchedule_ApplyResult_NaoPassaDaCaixaMaxima(t *testing.T) {
	rs := &domain.ReviewSchedule{Box: domain.MaxLeitnerBox}
	now := time.Now()

	rs.ApplyResult(true, now)

	assert.Equal(t, domain.MaxLeitnerBox, rs.Box)
}
