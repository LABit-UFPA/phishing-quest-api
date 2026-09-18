package usecase

import (
	"phishing-quest/adapter/repository"
	"phishing-quest/domain"

	"github.com/google/uuid"
)

type QuestionUseCase struct {
	questionRepo repository.IQuestionRepository
	answerRepo   repository.IAnswerRepository
}

func NewQuestionUseCase(questionRepo repository.IQuestionRepository, answerRepo repository.IAnswerRepository) *QuestionUseCase {
	return &QuestionUseCase{
		questionRepo: questionRepo,
		answerRepo:   answerRepo,
	}
}

func (quc *QuestionUseCase) GetAnswersByQuestionID(questionID uuid.UUID) ([]*domain.Answer, error) {
	answers, err := quc.answerRepo.GetByQuestionID(questionID)
	if err != nil {
		return nil, err
	}
	return answers, nil
}

func (quc *QuestionUseCase) CreateQuestion(questionRequest *domain.Question) (*domain.Question, error) {
	question := &domain.Question{
		Id:           uuid.New(),
		CategoryId:   questionRequest.CategoryId,
		QuestionText: questionRequest.QuestionText,
	}

	err := question.Validate()
	if err != nil {
		return nil, err
	}

	createdQuestion, err := quc.questionRepo.Create(question)
	if err != nil {
		return nil, err
	}

	return createdQuestion, nil
}

func (quc *QuestionUseCase) GetQuestion(id uuid.UUID) (*domain.Question, error) {
	question, err := quc.questionRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return question, nil
}

func (quc *QuestionUseCase) UpdateQuestion(id uuid.UUID, questionRequest *domain.Question) (*domain.Question, error) {
	question, err := quc.questionRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if questionRequest.CategoryId != uuid.Nil {
		question.CategoryId = questionRequest.CategoryId
	}
	if questionRequest.QuestionText != "" {
		question.QuestionText = questionRequest.QuestionText
	}

	err = question.Validate()
	if err != nil {
		return nil, err
	}

	updatedQuestion, err := quc.questionRepo.Update(question)
	if err != nil {
		return nil, err
	}

	return updatedQuestion, nil
}

func (quc *QuestionUseCase) DeleteQuestion(id uuid.UUID) error {
	_, err := quc.questionRepo.GetByID(id)
	if err != nil {
		return err
	}

	err = quc.questionRepo.Delete(id)
	if err != nil {
		return err
	}

	return nil
}
