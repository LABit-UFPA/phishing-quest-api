package tests

import (
	"testing"

	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockAttemptRepository implementa repository.IAttemptRepository.
type MockAttemptRepository struct {
	mock.Mock
}

func (m *MockAttemptRepository) Create(attempt *domain.Attempt) (*domain.Attempt, error) {
	args := m.Called(attempt)
	if err := args.Error(0); err != nil {
		return nil, err
	}
	return attempt, nil
}

func (m *MockAttemptRepository) Update(attempt *domain.Attempt) (*domain.Attempt, error) {
	args := m.Called(attempt)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Attempt), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAttemptRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAttemptRepository) GetByID(id uuid.UUID) (*domain.Attempt, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Attempt), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAttemptRepository) GetAll() ([]*domain.Attempt, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.Attempt), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAttemptRepository) GetByUserID(userID uuid.UUID) ([]*domain.Attempt, error) {
	args := m.Called(userID)
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.Attempt), args.Error(1)
	}
	return nil, args.Error(1)
}

// MockStudyParticipantRepository implementa repository.IStudyParticipantRepository.
type MockStudyParticipantRepository struct {
	mock.Mock
}

func (m *MockStudyParticipantRepository) Create(participant *domain.StudyParticipant) (*domain.StudyParticipant, error) {
	args := m.Called(participant)
	if err := args.Error(0); err != nil {
		return nil, err
	}
	return participant, nil
}

func (m *MockStudyParticipantRepository) HasConsented(userID uuid.UUID) (bool, error) {
	args := m.Called(userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockStudyParticipantRepository) GetByUserID(userID uuid.UUID) (*domain.StudyParticipant, error) {
	args := m.Called(userID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.StudyParticipant), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStudyParticipantRepository) Withdraw(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func newAttemptUseCaseWithMocks() (*usecase.AttemptUseCase, *MockAttemptRepository, *MockItemRepository, *MockStudyParticipantRepository) {
	mockAttemptRepo := new(MockAttemptRepository)
	mockItemRepo := new(MockItemRepository)
	mockConsentRepo := new(MockStudyParticipantRepository)
	// reviewScheduleUC = nil: o guard em AttemptUseCase.RegisterAttempt
	// (auc.reviewScheduleUC != nil) torna a integracao com a fila de
	// revisao Leitner opcional, permitindo testar RegisterAttempt de
	// forma isolada sem precisar mockar o repositorio de item_cues.
	uc := usecase.NewAttemptUseCase(mockAttemptRepo, mockItemRepo, mockConsentRepo, nil)
	return uc, mockAttemptRepo, mockItemRepo, mockConsentRepo
}

// TestAttemptUseCase_RegisterAttempt_RecusaSemConsentimento e a
// regressao central da issue #21: POST /attempts nao pode aceitar
// tentativas de usuarios que nao consentiram em participar da coleta.
func TestAttemptUseCase_RegisterAttempt_RecusaSemConsentimento(t *testing.T) {
	uc, mockAttemptRepo, mockItemRepo, mockConsentRepo := newAttemptUseCaseWithMocks()

	userID := uuid.New()
	mockConsentRepo.On("HasConsented", userID).Return(false, nil)

	created, err := uc.RegisterAttempt(&domain.Attempt{
		UserId:    userID,
		ItemId:    uuid.New(),
		SessionId: uuid.New(),
		Action:    domain.ActionReport,
	})

	assert.Nil(t, created)
	assert.ErrorIs(t, err, usecase.ErrConsentRequired)
	mockItemRepo.AssertNotCalled(t, "GetByID", mock.Anything)
	mockAttemptRepo.AssertNotCalled(t, "Create", mock.Anything)
}

// TestAttemptUseCase_RegisterAttempt_CalculaIsCorrectAPartirDoItem
// garante que IsCorrect nunca e confiado ao cliente: e sempre
// recalculado a partir de Item.IsMalicious e do Verdict informado.
func TestAttemptUseCase_RegisterAttempt_CalculaIsCorrectAPartirDoItem(t *testing.T) {
	uc, mockAttemptRepo, mockItemRepo, mockConsentRepo := newAttemptUseCaseWithMocks()

	userID := uuid.New()
	itemID := uuid.New()
	verdictTrue := true

	mockConsentRepo.On("HasConsented", userID).Return(true, nil)
	mockItemRepo.On("GetByID", itemID).Return(&domain.Item{Id: itemID, IsMalicious: true}, nil)
	mockAttemptRepo.On("Create", mock.AnythingOfType("*domain.Attempt")).Return(nil)

	created, err := uc.RegisterAttempt(&domain.Attempt{
		UserId:    userID,
		ItemId:    itemID,
		SessionId: uuid.New(),
		Action:    domain.ActionReport,
		Verdict:   &verdictTrue,
	})

	assert.NoError(t, err)
	assert.NotNil(t, created.IsCorrect)
	assert.True(t, *created.IsCorrect) // verdict=true e item.IsMalicious=true -> correto
	mockAttemptRepo.AssertExpectations(t)
}

func TestAttemptUseCase_RegisterAttempt_VerdictErradoResultaEmIsCorrectFalse(t *testing.T) {
	uc, mockAttemptRepo, mockItemRepo, mockConsentRepo := newAttemptUseCaseWithMocks()

	userID := uuid.New()
	itemID := uuid.New()
	verdictFalse := false // usuario disse "legitimo"

	mockConsentRepo.On("HasConsented", userID).Return(true, nil)
	mockItemRepo.On("GetByID", itemID).Return(&domain.Item{Id: itemID, IsMalicious: true}, nil) // mas e malicioso
	mockAttemptRepo.On("Create", mock.AnythingOfType("*domain.Attempt")).Return(nil)

	created, err := uc.RegisterAttempt(&domain.Attempt{
		UserId:    userID,
		ItemId:    itemID,
		SessionId: uuid.New(),
		Action:    domain.ActionIgnore,
		Verdict:   &verdictFalse,
	})

	assert.NoError(t, err)
	assert.NotNil(t, created.IsCorrect)
	assert.False(t, *created.IsCorrect)
}

func TestAttemptUseCase_RegisterAttempt_ItemInexistente(t *testing.T) {
	uc, mockAttemptRepo, mockItemRepo, mockConsentRepo := newAttemptUseCaseWithMocks()

	userID := uuid.New()
	itemID := uuid.New()

	mockConsentRepo.On("HasConsented", userID).Return(true, nil)
	mockItemRepo.On("GetByID", itemID).Return(nil, assert.AnError)

	created, err := uc.RegisterAttempt(&domain.Attempt{
		UserId:    userID,
		ItemId:    itemID,
		SessionId: uuid.New(),
		Action:    domain.ActionReport,
	})

	assert.Nil(t, created)
	assert.Error(t, err)
	mockAttemptRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestAttemptUseCase_RegisterConsent_PrimeiraVez(t *testing.T) {
	uc, _, _, mockConsentRepo := newAttemptUseCaseWithMocks()

	userID := uuid.New()
	mockConsentRepo.On("HasConsented", userID).Return(false, nil)
	mockConsentRepo.On("Create", mock.AnythingOfType("*domain.StudyParticipant")).Return(nil)

	participant, alreadyConsented, err := uc.RegisterConsent(usecase.ConsentRequest{
		UserID:         userID,
		ConsentVersion: "v1",
	})

	assert.NoError(t, err)
	assert.False(t, alreadyConsented)
	assert.Equal(t, userID, participant.UserId)
	// A condicao experimental e atribuida pelo servidor, nunca vazia.
	assert.NotEmpty(t, participant.Condition)
	mockConsentRepo.AssertExpectations(t)
}

// TestAttemptUseCase_RegisterConsent_Idempotente garante que consentir
// de novo nao tenta um INSERT que violaria a PK, e nao mascara erros
// reais de banco como "ja consentiu".
func TestAttemptUseCase_RegisterConsent_Idempotente(t *testing.T) {
	uc, _, _, mockConsentRepo := newAttemptUseCaseWithMocks()

	userID := uuid.New()
	existing := &domain.StudyParticipant{UserId: userID, Condition: "feedback_binario"}
	mockConsentRepo.On("HasConsented", userID).Return(true, nil)
	mockConsentRepo.On("GetByUserID", userID).Return(existing, nil)

	participant, alreadyConsented, err := uc.RegisterConsent(usecase.ConsentRequest{
		UserID:         userID,
		ConsentVersion: "v1",
	})

	assert.NoError(t, err)
	assert.True(t, alreadyConsented)
	assert.Equal(t, userID, participant.UserId)
	assert.Equal(t, "feedback_binario", participant.Condition)
	mockConsentRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestAttemptUseCase_RegisterConsent_PropagaErroRealDeConexao(t *testing.T) {
	uc, _, _, mockConsentRepo := newAttemptUseCaseWithMocks()

	userID := uuid.New()
	mockConsentRepo.On("HasConsented", userID).Return(false, assert.AnError)

	participant, alreadyConsented, err := uc.RegisterConsent(usecase.ConsentRequest{
		UserID:         userID,
		ConsentVersion: "v1",
	})

	assert.Nil(t, participant)
	assert.False(t, alreadyConsented)
	assert.Error(t, err)
}

// TestAttemptUseCase_RegisterAttempt_RecusaAposWithdraw simula o
// direito de exclusao (LGPD): apos WithdrawConsent, HasConsented deve
// refletir que o participante nao esta mais ativo, bloqueando novas
// tentativas.
func TestAttemptUseCase_RegisterAttempt_RecusaAposWithdraw(t *testing.T) {
	uc, mockAttemptRepo, mockItemRepo, mockConsentRepo := newAttemptUseCaseWithMocks()

	userID := uuid.New()
	mockConsentRepo.On("Withdraw", userID).Return(nil)
	// Apos o withdraw, o repositorio real (via WHERE withdrawn_at IS NULL)
	// passaria a responder false aqui — simulado explicitamente no mock.
	mockConsentRepo.On("HasConsented", userID).Return(false, nil)

	err := uc.WithdrawConsent(userID)
	assert.NoError(t, err)

	created, err := uc.RegisterAttempt(&domain.Attempt{
		UserId:    userID,
		ItemId:    uuid.New(),
		SessionId: uuid.New(),
		Action:    domain.ActionReport,
	})

	assert.Nil(t, created)
	assert.ErrorIs(t, err, usecase.ErrConsentRequired)
	mockItemRepo.AssertNotCalled(t, "GetByID", mock.Anything)
	mockAttemptRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestAttemptUseCase_WithdrawConsent_PropagaErroDoRepo(t *testing.T) {
	uc, _, _, mockConsentRepo := newAttemptUseCaseWithMocks()

	userID := uuid.New()
	mockConsentRepo.On("Withdraw", userID).Return(gorm.ErrRecordNotFound)

	err := uc.WithdrawConsent(userID)

	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestAttemptUseCase_ListAttemptsByUser(t *testing.T) {
	uc, mockAttemptRepo, _, _ := newAttemptUseCaseWithMocks()

	userID := uuid.New()
	expected := []*domain.Attempt{{Id: uuid.New(), UserId: userID}}
	mockAttemptRepo.On("GetByUserID", userID).Return(expected, nil)

	attempts, err := uc.ListAttemptsByUser(userID)

	assert.NoError(t, err)
	assert.Len(t, attempts, 1)
	mockAttemptRepo.AssertExpectations(t)
}
