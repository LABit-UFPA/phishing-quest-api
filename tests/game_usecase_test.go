package tests

import (
	"testing"

	"phishing-quest/core/usecase"
	"phishing-quest/domain"
	"phishing-quest/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAnswerRepository implementa repository.IAnswerRepository.
type MockAnswerRepository struct {
	mock.Mock
}

func (m *MockAnswerRepository) Create(answer *domain.Answer) (*domain.Answer, error) {
	args := m.Called(answer)
	if err := args.Error(0); err != nil {
		return nil, err
	}
	return answer, nil
}

func (m *MockAnswerRepository) Update(answer *domain.Answer) (*domain.Answer, error) {
	args := m.Called(answer)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Answer), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAnswerRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAnswerRepository) GetByID(id uuid.UUID) (*domain.Answer, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Answer), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAnswerRepository) GetAll() ([]*domain.Answer, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.Answer), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAnswerRepository) GetByQuestionID(questionID uuid.UUID) ([]*domain.Answer, error) {
	args := m.Called(questionID)
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.Answer), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockUserScoreRepository implementa repository.IUserScoreRepository.
type MockUserScoreRepository struct {
	mock.Mock
}

func (m *MockUserScoreRepository) Create(score *domain.UserScore) (*domain.UserScore, error) {
	args := m.Called(score)
	if err := args.Error(0); err != nil {
		return nil, err
	}
	return score, nil
}

func (m *MockUserScoreRepository) Update(score *domain.UserScore) (*domain.UserScore, error) {
	args := m.Called(score)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.UserScore), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserScoreRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserScoreRepository) GetByID(id uuid.UUID) (*domain.UserScore, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.UserScore), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserScoreRepository) GetAll() ([]*domain.UserScore, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.UserScore), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserScoreRepository) IncrementScore(userID uuid.UUID, points int) error {
	args := m.Called(userID, points)
	return args.Error(0)
}

func (m *MockUserScoreRepository) GetUserScore(userID uuid.UUID) (*domain.UserScore, error) {
	args := m.Called(userID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.UserScore), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockUserAnswerRepository implementa repository.IUserAnswerRepository.
type MockUserAnswerRepository struct {
	mock.Mock
}

func (m *MockUserAnswerRepository) Create(userAnswer *domain.UserAnswer) (*domain.UserAnswer, error) {
	args := m.Called(userAnswer)
	if err := args.Error(0); err != nil {
		return nil, err
	}
	return userAnswer, nil
}

func (m *MockUserAnswerRepository) Update(userAnswer *domain.UserAnswer) (*domain.UserAnswer, error) {
	args := m.Called(userAnswer)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.UserAnswer), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserAnswerRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserAnswerRepository) GetByID(id uuid.UUID) (*domain.UserAnswer, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.UserAnswer), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserAnswerRepository) GetAll() ([]*domain.UserAnswer, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.UserAnswer), args.Error(1)
	}
	return nil, args.Error(1)
}

func newGameUseCaseWithMocks() (*usecase.GameUseCase, *MockAnswerRepository, *MockUserRepository, *MockUserScoreRepository, *MockUserAnswerRepository) {
	mockAnswerRepo := new(MockAnswerRepository)
	mockUserRepo := new(MockUserRepository)
	mockUserScoreRepo := new(MockUserScoreRepository)
	mockUserAnswerRepo := new(MockUserAnswerRepository)

	uc := usecase.NewGameUseCase(mockAnswerRepo, mockUserRepo, mockUserScoreRepo, mockUserAnswerRepo)
	return uc, mockAnswerRepo, mockUserRepo, mockUserScoreRepo, mockUserAnswerRepo
}

// TestGameUseCase_ProcessAnswer_Correta_IncrementaScoreEGravaHistorico e a
// regressao da issue #14: antes, uma resposta correta de um usuario sem
// linha previa em user_scores causava 500 (UPDATE sem match), e nenhuma
// linha era gravada em user_answers independente do resultado.
func TestGameUseCase_ProcessAnswer_Correta_IncrementaScoreEGravaHistorico(t *testing.T) {
	uc, mockAnswerRepo, _, mockUserScoreRepo, mockUserAnswerRepo := newGameUseCaseWithMocks()

	questionID := uuid.New()
	answerID := uuid.New()
	userID := uuid.New()

	answer := &domain.Answer{Id: answerID, QuestionId: questionID, IsCorrect: true}
	mockAnswerRepo.On("GetByID", answerID).Return(answer, nil)

	mockUserAnswerRepo.On("Create", mock.AnythingOfType("*domain.UserAnswer")).Return(nil)
	mockUserScoreRepo.On("IncrementScore", userID, 10).Return(nil)
	mockUserScoreRepo.On("GetUserScore", userID).Return(&domain.UserScore{UserId: userID, Score: 30}, nil)

	result, err := uc.ProcessAnswer(&dto.SubmitAnswerDTO{
		UserID:     userID,
		QuestionID: questionID,
		AnswerID:   answerID,
	})

	assert.NoError(t, err)
	assert.True(t, result.IsCorrect)
	assert.Equal(t, 30, result.TotalScore)

	mockAnswerRepo.AssertExpectations(t)
	mockUserScoreRepo.AssertExpectations(t)
	mockUserAnswerRepo.AssertExpectations(t)

	// A tentativa gravada precisa carregar isCorrect=true.
	createdArg := mockUserAnswerRepo.Calls[0].Arguments.Get(0).(*domain.UserAnswer)
	assert.NotNil(t, createdArg.IsCorrect)
	assert.True(t, *createdArg.IsCorrect)
	assert.Equal(t, userID, createdArg.UserId)
	assert.Equal(t, questionID, createdArg.QuestionId)
	assert.Equal(t, answerID, createdArg.AnswerId)
}

func TestGameUseCase_ProcessAnswer_Incorreta_NaoIncrementaScoreMasGravaHistorico(t *testing.T) {
	uc, mockAnswerRepo, _, mockUserScoreRepo, mockUserAnswerRepo := newGameUseCaseWithMocks()

	questionID := uuid.New()
	answerID := uuid.New()
	userID := uuid.New()

	answer := &domain.Answer{Id: answerID, QuestionId: questionID, IsCorrect: false}
	mockAnswerRepo.On("GetByID", answerID).Return(answer, nil)
	mockUserAnswerRepo.On("Create", mock.AnythingOfType("*domain.UserAnswer")).Return(nil)
	mockUserScoreRepo.On("GetUserScore", userID).Return(&domain.UserScore{UserId: userID, Score: 0}, nil)

	result, err := uc.ProcessAnswer(&dto.SubmitAnswerDTO{
		UserID:     userID,
		QuestionID: questionID,
		AnswerID:   answerID,
	})

	assert.NoError(t, err)
	assert.False(t, result.IsCorrect)

	mockUserScoreRepo.AssertNotCalled(t, "IncrementScore", mock.Anything, mock.Anything)
	mockUserAnswerRepo.AssertExpectations(t)

	createdArg := mockUserAnswerRepo.Calls[0].Arguments.Get(0).(*domain.UserAnswer)
	assert.NotNil(t, createdArg.IsCorrect)
	assert.False(t, *createdArg.IsCorrect)
}

func TestGameUseCase_ProcessAnswer_RespostaNaoPertenceAQuestao(t *testing.T) {
	uc, mockAnswerRepo, _, mockUserScoreRepo, mockUserAnswerRepo := newGameUseCaseWithMocks()

	answerID := uuid.New()
	answer := &domain.Answer{Id: answerID, QuestionId: uuid.New(), IsCorrect: true}
	mockAnswerRepo.On("GetByID", answerID).Return(answer, nil)

	result, err := uc.ProcessAnswer(&dto.SubmitAnswerDTO{
		UserID:     uuid.New(),
		QuestionID: uuid.New(), // diferente de answer.QuestionId
		AnswerID:   answerID,
	})

	assert.Nil(t, result)
	assert.EqualError(t, err, "answer does not belong to the given question")
	mockUserAnswerRepo.AssertNotCalled(t, "Create", mock.Anything)
	mockUserScoreRepo.AssertNotCalled(t, "IncrementScore", mock.Anything, mock.Anything)
}

// TestUserScoreRepository_IncrementScore_documenta o contrato exigido pelo
// usecase: IncrementScore nao pode falhar so porque o usuario ainda nao
// tem nenhum evento de pontuacao. Isso e coberto no nivel de usecase
// acima (via mock) porque o repositorio real depende de um Postgres —
// o teste de integracao com banco fica para a issue #31 (CI).
func TestGameUseCase_ProcessAnswer_PrimeiroAcertoDoUsuario(t *testing.T) {
	uc, mockAnswerRepo, _, mockUserScoreRepo, mockUserAnswerRepo := newGameUseCaseWithMocks()

	questionID := uuid.New()
	answerID := uuid.New()
	userID := uuid.New() // usuario sem nenhum evento previo em user_scores

	answer := &domain.Answer{Id: answerID, QuestionId: questionID, IsCorrect: true}
	mockAnswerRepo.On("GetByID", answerID).Return(answer, nil)
	mockUserAnswerRepo.On("Create", mock.AnythingOfType("*domain.UserAnswer")).Return(nil)
	// IncrementScore agora faz INSERT (nao UPDATE), entao nao ha
	// gorm.ErrRecordNotFound possivel so por o usuario ser novo.
	mockUserScoreRepo.On("IncrementScore", userID, 10).Return(nil)
	mockUserScoreRepo.On("GetUserScore", userID).Return(&domain.UserScore{UserId: userID, Score: 10}, nil)

	result, err := uc.ProcessAnswer(&dto.SubmitAnswerDTO{
		UserID:     userID,
		QuestionID: questionID,
		AnswerID:   answerID,
	})

	assert.NoError(t, err)
	assert.Equal(t, 10, result.TotalScore)
}
