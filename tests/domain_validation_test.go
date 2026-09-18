package tests

import (
	"testing"

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
