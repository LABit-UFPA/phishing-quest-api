package dto

import "github.com/google/uuid"

type QuestionAnswersDTO struct {
	QuestionId uuid.UUID    `json:"questionId"`
	Answers    []*AnswerDTO `json:"answers"`
}

type AnswerDTO struct {
	Id         uuid.UUID `json:"id"`
	AnswerText string    `json:"answerText"`
	IsCorrect  bool      `json:"isCorrect"`
}

// SubmitAnswerDTO e AnswerResultDTO usam camelCase, alinhado com a
// convencao do resto da API (categoryName, questionText, isCorrect,
// userId...). Antes desta mudanca eram os unicos DTOs em snake_case.
type SubmitAnswerDTO struct {
	UserID     uuid.UUID `json:"userId" binding:"required"`
	QuestionID uuid.UUID `json:"questionId" binding:"required"`
	AnswerID   uuid.UUID `json:"answerId" binding:"required"`
}

type AnswerResultDTO struct {
	IsCorrect  bool   `json:"isCorrect"`
	Message    string `json:"message"`
	TotalScore int    `json:"totalScore"`
}
