package domain

import (
	"phishing-quest/dto"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Question nao tem mais o campo CorrectAnswer: a coluna correct_answer
// foi removida pela migration V20250104153937 (a "resposta correta" e
// modelada por Answer.IsCorrect, associada via QuestionId). Manter o
// campo no struct fazia o GORM tentar ler/escrever uma coluna
// inexistente em qualquer create/update de questao.
type Question struct {
	Id           uuid.UUID `json:"id" gorm:"primaryKey"`
	CategoryId   uuid.UUID `json:"categoryId" validate:"required"`
	QuestionText string    `json:"questionText" validate:"required"`
}

func (q *Question) TableName() string {
	return "phishing_quest.questions"
}

func (q *Question) Validate() error {
	validate := validator.New()
	return validate.Struct(q)
}

func (q *Question) ToDTO() *dto.QuestionDTO {
	return &dto.QuestionDTO{
		QuestionId:   q.Id,
		QuestionText: q.QuestionText,
	}
}
