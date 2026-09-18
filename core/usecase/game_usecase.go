package usecase

import (
	"errors"
	"phishing-quest/adapter/repository"
	"phishing-quest/domain"
	"phishing-quest/dto"

	"github.com/google/uuid"
)

// pointsPerCorrectAnswer e fixo por enquanto. Fica sujeito a mudar
// quando o calculo de pontuacao por teoria de deteccao de sinal for
// implementado (ROADMAP_PESQUISA_2027.md, Fase 3).
const pointsPerCorrectAnswer = 10

type GameUseCase struct {
	answerRepo     repository.IAnswerRepository
	userRepo       repository.IUserRepository
	userScoreRepo  repository.IUserScoreRepository
	userAnswerRepo repository.IUserAnswerRepository
}

func NewGameUseCase(
	answerRepo repository.IAnswerRepository,
	userRepo repository.IUserRepository,
	userScoreRepo repository.IUserScoreRepository,
	userAnswerRepo repository.IUserAnswerRepository,
) *GameUseCase {
	return &GameUseCase{
		answerRepo:     answerRepo,
		userRepo:       userRepo,
		userScoreRepo:  userScoreRepo,
		userAnswerRepo: userAnswerRepo,
	}
}

func (guc *GameUseCase) ProcessAnswer(sbaDTO *dto.SubmitAnswerDTO) (*dto.AnswerResultDTO, error) {
	answer, err := guc.answerRepo.GetByID(sbaDTO.AnswerID)
	if err != nil {
		return nil, err
	}

	if answer.QuestionId != sbaDTO.QuestionID {
		return nil, errors.New("answer does not belong to the given question")
	}

	isCorrect := answer.IsCorrect

	// Grava o historico da tentativa independente do resultado — antes
	// disso o fluxo de /game/answer nao deixava nenhum rastro em
	// user_answers, apesar de a tabela existir com esse proposito.
	isCorrectCopy := isCorrect
	userAnswer := &domain.UserAnswer{
		UserAnswerId: uuid.New(),
		UserId:       sbaDTO.UserID,
		QuestionId:   sbaDTO.QuestionID,
		AnswerId:     sbaDTO.AnswerID,
		IsCorrect:    &isCorrectCopy,
	}
	if _, err = guc.userAnswerRepo.Create(userAnswer); err != nil {
		return nil, err
	}

	if isCorrect {
		if err = guc.userScoreRepo.IncrementScore(sbaDTO.UserID, pointsPerCorrectAnswer); err != nil {
			return nil, err
		}
	}

	totalScore := 0
	if userScore, scoreErr := guc.userScoreRepo.GetUserScore(sbaDTO.UserID); scoreErr == nil {
		totalScore = userScore.Score
	}

	result := &dto.AnswerResultDTO{
		IsCorrect:  isCorrect,
		Message:    "Resposta processada com sucesso!",
		TotalScore: totalScore,
	}

	return result, nil
}
