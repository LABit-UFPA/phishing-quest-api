package tests

import (
	"testing"
	"time"

	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockReviewScheduleRepository implementa repository.IReviewScheduleRepository.
type MockReviewScheduleRepository struct {
	mock.Mock
}

func (m *MockReviewScheduleRepository) GetOrCreate(userID, itemID, cueID uuid.UUID) (*domain.ReviewSchedule, error) {
	args := m.Called(userID, itemID, cueID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.ReviewSchedule), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockReviewScheduleRepository) Update(schedule *domain.ReviewSchedule) (*domain.ReviewSchedule, error) {
	args := m.Called(schedule)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.ReviewSchedule), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockReviewScheduleRepository) GetDue(userID uuid.UUID, now time.Time, limit int) ([]*domain.ReviewSchedule, error) {
	args := m.Called(userID, now, limit)
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.ReviewSchedule), args.Error(1)
	}
	return nil, args.Error(1)
}

func newReviewScheduleUseCaseWithMocks() (*usecase.ReviewScheduleUseCase, *MockReviewScheduleRepository, *MockItemCueRepository) {
	mockScheduleRepo := new(MockReviewScheduleRepository)
	mockItemCueRepo := new(MockItemCueRepository)
	uc := usecase.NewReviewScheduleUseCase(mockScheduleRepo, mockItemCueRepo)
	return uc, mockScheduleRepo, mockItemCueRepo
}

// TestReviewScheduleUseCase_RecordOutcome_AtualizaAgendamentoDeCadaPista
// garante o comportamento central da issue #27: uma tentativa correta
// atualiza o agendamento Leitner de TODAS as pistas presentes no item
// (nao so uma), permitindo diagnosticar cada pista independentemente.
func TestReviewScheduleUseCase_RecordOutcome_AtualizaAgendamentoDeCadaPista(t *testing.T) {
	uc, mockScheduleRepo, mockItemCueRepo := newReviewScheduleUseCaseWithMocks()

	userID := uuid.New()
	itemID := uuid.New()
	cue1 := uuid.New()
	cue2 := uuid.New()

	mockItemCueRepo.On("GetByItemID", itemID).Return([]*domain.ItemCue{
		{Id: uuid.New(), ItemId: itemID, CueId: cue1},
		{Id: uuid.New(), ItemId: itemID, CueId: cue2},
	}, nil)

	schedule1 := &domain.ReviewSchedule{Id: uuid.New(), UserId: userID, ItemId: itemID, CueId: cue1, Box: domain.MinLeitnerBox}
	schedule2 := &domain.ReviewSchedule{Id: uuid.New(), UserId: userID, ItemId: itemID, CueId: cue2, Box: domain.MinLeitnerBox}
	mockScheduleRepo.On("GetOrCreate", userID, itemID, cue1).Return(schedule1, nil)
	mockScheduleRepo.On("GetOrCreate", userID, itemID, cue2).Return(schedule2, nil)
	mockScheduleRepo.On("Update", mock.AnythingOfType("*domain.ReviewSchedule")).Return(&domain.ReviewSchedule{}, nil)

	err := uc.RecordOutcome(userID, itemID, true)

	assert.NoError(t, err)
	mockScheduleRepo.AssertNumberOfCalls(t, "Update", 2)
	// Ambas as pistas avancaram de caixa apos o acerto.
	assert.Equal(t, 2, schedule1.Box)
	assert.Equal(t, 2, schedule2.Box)
}

// TestReviewScheduleUseCase_RecordOutcome_ErroVoltaCaixaParaUm garante
// que uma tentativa incorreta reseta a(s) pista(s) do item para a
// caixa 1, mesmo que ja estivessem avancadas.
func TestReviewScheduleUseCase_RecordOutcome_ErroVoltaCaixaParaUm(t *testing.T) {
	uc, mockScheduleRepo, mockItemCueRepo := newReviewScheduleUseCaseWithMocks()

	userID := uuid.New()
	itemID := uuid.New()
	cueID := uuid.New()

	mockItemCueRepo.On("GetByItemID", itemID).Return([]*domain.ItemCue{
		{Id: uuid.New(), ItemId: itemID, CueId: cueID},
	}, nil)

	schedule := &domain.ReviewSchedule{Id: uuid.New(), UserId: userID, ItemId: itemID, CueId: cueID, Box: 4}
	mockScheduleRepo.On("GetOrCreate", userID, itemID, cueID).Return(schedule, nil)
	mockScheduleRepo.On("Update", mock.AnythingOfType("*domain.ReviewSchedule")).Return(&domain.ReviewSchedule{}, nil)

	err := uc.RecordOutcome(userID, itemID, false)

	assert.NoError(t, err)
	assert.Equal(t, domain.MinLeitnerBox, schedule.Box)
}

// TestReviewScheduleUseCase_RecordOutcome_ItemSemPistasNaoFalha garante
// que um item sem pistas anotadas (ainda) simplesmente nao cria
// agendamento nenhum, em vez de retornar erro — cenario esperado
// durante a curadoria incremental do banco de itens.
func TestReviewScheduleUseCase_RecordOutcome_ItemSemPistasNaoFalha(t *testing.T) {
	uc, mockScheduleRepo, mockItemCueRepo := newReviewScheduleUseCaseWithMocks()

	userID := uuid.New()
	itemID := uuid.New()
	mockItemCueRepo.On("GetByItemID", itemID).Return([]*domain.ItemCue{}, nil)

	err := uc.RecordOutcome(userID, itemID, true)

	assert.NoError(t, err)
	mockScheduleRepo.AssertNotCalled(t, "GetOrCreate", mock.Anything, mock.Anything, mock.Anything)
}

// TestReviewScheduleUseCase_RecordOutcome_PropagaErroDoRepositorio
// garante que falhas reais de banco nao sao silenciadas dentro do
// proprio usecase — quem decide ignorar o erro (deliberadamente) e o
// AttemptUseCase, nao o ReviewScheduleUseCase.
func TestReviewScheduleUseCase_RecordOutcome_PropagaErroDoRepositorio(t *testing.T) {
	uc, _, mockItemCueRepo := newReviewScheduleUseCaseWithMocks()

	userID := uuid.New()
	itemID := uuid.New()
	mockItemCueRepo.On("GetByItemID", itemID).Return(nil, assert.AnError)

	err := uc.RecordOutcome(userID, itemID, true)

	assert.Error(t, err)
}

// TestReviewScheduleUseCase_GetDueReviews_DelegaAoRepositorioComAgora
// garante que o usecase passa o "agora" do servidor (nao confia em
// timestamp de cliente) e o limite solicitado ao repositorio.
func TestReviewScheduleUseCase_GetDueReviews_DelegaAoRepositorioComAgora(t *testing.T) {
	uc, mockScheduleRepo, _ := newReviewScheduleUseCaseWithMocks()

	userID := uuid.New()
	expected := []*domain.ReviewSchedule{{Id: uuid.New(), UserId: userID}}
	mockScheduleRepo.On("GetDue", userID, mock.AnythingOfType("time.Time"), 10).Return(expected, nil)

	due, err := uc.GetDueReviews(userID, 10)

	assert.NoError(t, err)
	assert.Len(t, due, 1)
	mockScheduleRepo.AssertExpectations(t)
}
