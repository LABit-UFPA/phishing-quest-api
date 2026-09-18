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

type SubmitAnswerDTO struct {
	UserID     uuid.UUID `json:"user_id" binding:"required"`
	QuestionID uuid.UUID `json:"question_id" binding:"required"`
	AnswerID   uuid.UUID `json:"answer_id" binding:"required"`
}

type AnswerResultDTO struct {
	IsCorrect  bool   `json:"is_correct"`
	Message    string `json:"message"`
	TotalScore int    `json:"total_score"`
}
