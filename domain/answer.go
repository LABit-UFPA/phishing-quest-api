package domain

import (
	"phishing-quest/dto"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type Answer struct {
	Id         uuid.UUID `json:"id" gorm:"primaryKey"`
	QuestionId uuid.UUID `json:"questionId" validate:"required"`
	AnswerText string    `json:"answerText" validate:"required"`
	// IsCorrect nao tem a tag "required": o pacote go-playground/validator
	// trata bool false como valor-zero e reprova a validacao "required",
	// o que impedia criar respostas incorretas (isCorrect:false) via
	// POST /api/v1/answers.
	IsCorrect bool `json:"isCorrect"`
}

func (a *Answer) TableName() string {
	return "phishing_quest.answers"
}

func (a *Answer) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}

func (a *Answer) ToDTO() *dto.AnswerDTO {
	return &dto.AnswerDTO{
		Id:         a.Id,
		AnswerText: a.AnswerText,
		IsCorrect:  a.IsCorrect,
	}
}
