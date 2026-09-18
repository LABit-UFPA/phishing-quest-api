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

// TestItemSelectionUseCase_PrefereLadoComMenosExposicao e a regressao
// central da issue #22: se o usuario ja viu mais itens maliciosos que
// legitimos na sessao, o proximo item pedido deve priorizar o lado
// legitimo (para nao desbalancear o instrumento so pelo acaso).
func TestItemSelectionUseCase_PrefereLadoComMenosExposicao(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	expectedItem := &domain.Item{Id: uuid.New(), IsMalicious: false}

	// 2 maliciosos vistos, 0 legitimos -> deve pedir isMalicious=false
	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(2), int64(0), nil)
	mockItemRepo.On("GetRandomUnseen", userID, sessionID, mock.MatchedBy(func(v *bool) bool {
		return v != nil && *v == false
	})).Return(expectedItem, nil)

	item, err := uc.NextItem(userID, sessionID, usecase.ModeBalanced)

	assert.NoError(t, err)
	assert.Equal(t, expectedItem.Id, item.Id)
	mockItemRepo.AssertExpectations(t)
}

func TestItemSelectionUseCase_PrefereMaliciosoQuandoLegitimoPredomina(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	expectedItem := &domain.Item{Id: uuid.New(), IsMalicious: true}

	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(0), int64(3), nil)
	mockItemRepo.On("GetRandomUnseen", userID, sessionID, mock.MatchedBy(func(v *bool) bool {
		return v != nil && *v == true
	})).Return(expectedItem, nil)

	item, err := uc.NextItem(userID, sessionID, usecase.ModeBalanced)

	assert.NoError(t, err)
	assert.Equal(t, expectedItem.Id, item.Id)
}

func TestItemSelectionUseCase_QualquerLadoQuandoEmpatado(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	expectedItem := &domain.Item{Id: uuid.New()}

	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(1), int64(1), nil)
	mockItemRepo.On("GetRandomUnseen", userID, sessionID, (*bool)(nil)).Return(expectedItem, nil)

	item, err := uc.NextItem(userID, sessionID, usecase.ModeBalanced)

	assert.NoError(t, err)
	assert.Equal(t, expectedItem.Id, item.Id)
}

// TestItemSelectionUseCase_FallbackQuandoLadoPreferidoEsgotado garante
// que, se o lado preferido pelo balanceamento nao tem mais itens
// disponiveis, o usecase tenta "qualquer item nao visto" em vez de
// falhar — melhor entregar um item desbalanceado do que travar a sessao.
func TestItemSelectionUseCase_FallbackQuandoLadoPreferidoEsgotado(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	fallbackItem := &domain.Item{Id: uuid.New(), IsMalicious: true}

	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(0), int64(3), nil)
	// Lado preferido (malicioso=true) esgotado.
	mockItemRepo.On("GetRandomUnseen", userID, sessionID, mock.MatchedBy(func(v *bool) bool {
		return v != nil && *v == true
	})).Return(nil, gorm.ErrRecordNotFound)
	// Fallback: qualquer item nao visto.
	mockItemRepo.On("GetRandomUnseen", userID, sessionID, (*bool)(nil)).Return(fallbackItem, nil)

	item, err := uc.NextItem(userID, sessionID, usecase.ModeBalanced)

	assert.NoError(t, err)
	assert.Equal(t, fallbackItem.Id, item.Id)
}

// TestItemSelectionUseCase_SemItensDisponiveis garante o criterio de
// aceite "nao repetir itens ja vistos": quando todos os itens da
// sessao ja foram respondidos, retorna ErrNoUnseenItems (404 no
// handler) em vez de um item repetido.
func TestItemSelectionUseCase_SemItensDisponiveis(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo)

	userID := uuid.New()
	sessionID := uuid.New()

	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(3), int64(3), nil)
	mockItemRepo.On("GetRandomUnseen", userID, sessionID, (*bool)(nil)).Return(nil, gorm.ErrRecordNotFound)

	item, err := uc.NextItem(userID, sessionID, usecase.ModeBalanced)

	assert.Nil(t, item)
	assert.ErrorIs(t, err, usecase.ErrNoUnseenItems)
}

func TestItemSelectionUseCase_PropagaErroDeContagem(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo)

	userID := uuid.New()
	sessionID := uuid.New()

	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(0), int64(0), assert.AnError)

	item, err := uc.NextItem(userID, sessionID, usecase.ModeBalanced)

	assert.Nil(t, item)
	assert.Error(t, err)
	mockItemRepo.AssertNotCalled(t, "GetRandomUnseen", mock.Anything, mock.Anything, mock.Anything)
}

// TestItemSelectionUseCase_ModoAdaptativoAindaFuncionaComoBalanceado
// documenta o gancho da issue #28: enquanto a selecao adaptativa por
// pista nao existe, ModeAdaptive se comporta como ModeBalanced.
func TestItemSelectionUseCase_ModoAdaptativoAindaFuncionaComoBalanceado(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	expectedItem := &domain.Item{Id: uuid.New()}

	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(1), int64(1), nil)
	mockItemRepo.On("GetRandomUnseen", userID, sessionID, (*bool)(nil)).Return(expectedItem, nil)

	item, err := uc.NextItem(userID, sessionID, usecase.ModeAdaptive)

	assert.NoError(t, err)
	assert.Equal(t, expectedItem.Id, item.Id)
}
