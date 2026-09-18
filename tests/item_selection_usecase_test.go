package tests

import (
	"testing"

	"phishing-quest/adapter/repository"
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
	uc := usecase.NewItemSelectionUseCase(mockItemRepo, new(MockUserStatsRepository))

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
	uc := usecase.NewItemSelectionUseCase(mockItemRepo, new(MockUserStatsRepository))

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
	uc := usecase.NewItemSelectionUseCase(mockItemRepo, new(MockUserStatsRepository))

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
	uc := usecase.NewItemSelectionUseCase(mockItemRepo, new(MockUserStatsRepository))

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
	uc := usecase.NewItemSelectionUseCase(mockItemRepo, new(MockUserStatsRepository))

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
	uc := usecase.NewItemSelectionUseCase(mockItemRepo, new(MockUserStatsRepository))

	userID := uuid.New()
	sessionID := uuid.New()

	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(0), int64(0), assert.AnError)

	item, err := uc.NextItem(userID, sessionID, usecase.ModeBalanced)

	assert.Nil(t, item)
	assert.Error(t, err)
	mockItemRepo.AssertNotCalled(t, "GetRandomUnseen", mock.Anything, mock.Anything, mock.Anything)
}

// TestItemSelectionUseCase_ModoAdaptativoSemPistaFracaCaiNoBalanceado
// garante que, sem historico suficiente para diagnosticar uma pista
// fraca, o modo adaptativo nao trava nem falha: cai no comportamento
// balanceado (issue #28).
func TestItemSelectionUseCase_ModoAdaptativoSemPistaFracaCaiNoBalanceado(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	mockStatsRepo := new(MockUserStatsRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo, mockStatsRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	expectedItem := &domain.Item{Id: uuid.New()}

	// Usuario sem tentativas: nenhuma pista diagnosticavel.
	mockStatsRepo.On("GetCueMastery", userID).Return([]repository.CueMastery{}, nil)
	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(1), int64(1), nil)
	mockItemRepo.On("GetRandomUnseen", userID, sessionID, (*bool)(nil)).Return(expectedItem, nil)

	item, err := uc.NextItem(userID, sessionID, usecase.ModeAdaptive)

	assert.NoError(t, err)
	assert.Equal(t, expectedItem.Id, item.Id)
	mockItemRepo.AssertNotCalled(t, "GetRandomUnseenByCues", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// TestItemSelectionUseCase_ModoAdaptativoFocaPistasComPiorDesempenho e
// a regressao central da issue #28: a selecao deve concentrar nas
// pistas de pior desempenho do usuario, ignorando as ja dominadas.
func TestItemSelectionUseCase_ModoAdaptativoFocaPistasComPiorDesempenho(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	mockStatsRepo := new(MockUserStatsRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo, mockStatsRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	cueDominada := uuid.New()
	cueFraca := uuid.New()
	expectedItem := &domain.Item{Id: uuid.New(), IsMalicious: true}

	mockStatsRepo.On("GetCueMastery", userID).Return([]repository.CueMastery{
		// 90% de acerto: dominada (acima do threshold default 0.75).
		{CueId: cueDominada, CueCode: "urgency", Answered: 10, Correct: 9},
		// 30% de acerto: fraca, deve entrar no foco.
		{CueId: cueFraca, CueCode: "typosquat", Answered: 10, Correct: 3},
	}, nil)
	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(1), int64(1), nil)
	// Lado empatado -> preferMalicious nil -> vai direto na busca por pista.
	mockItemRepo.On("GetRandomUnseenByCues", userID, sessionID, []uuid.UUID{cueFraca}, (*bool)(nil)).
		Return(expectedItem, nil)

	item, err := uc.NextItem(userID, sessionID, usecase.ModeAdaptive)

	assert.NoError(t, err)
	assert.Equal(t, expectedItem.Id, item.Id)
	mockItemRepo.AssertExpectations(t)
	// A pista dominada nunca entra no filtro.
	mockItemRepo.AssertNotCalled(t, "GetRandomUnseen", mock.Anything, mock.Anything, mock.Anything)
}

// TestItemSelectionUseCase_ModoAdaptativoIgnoraPistaComPoucaExposicao
// garante o parametro MinExposures: errar 1 de 1 nao caracteriza
// dificuldade sistematica, entao a pista nao entra no foco (e a selecao
// cai no balanceado por falta de candidato adaptativo).
func TestItemSelectionUseCase_ModoAdaptativoIgnoraPistaComPoucaExposicao(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	mockStatsRepo := new(MockUserStatsRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo, mockStatsRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	fallbackItem := &domain.Item{Id: uuid.New()}

	mockStatsRepo.On("GetCueMastery", userID).Return([]repository.CueMastery{
		// 0% de acerto, mas so 1 exposicao (default MinExposures = 3).
		{CueId: uuid.New(), CueCode: "homoglyph", Answered: 1, Correct: 0},
	}, nil)
	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(1), int64(1), nil)
	mockItemRepo.On("GetRandomUnseen", userID, sessionID, (*bool)(nil)).Return(fallbackItem, nil)

	item, err := uc.NextItem(userID, sessionID, usecase.ModeAdaptive)

	assert.NoError(t, err)
	assert.Equal(t, fallbackItem.Id, item.Id)
	mockItemRepo.AssertNotCalled(t, "GetRandomUnseenByCues", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// TestItemSelectionUseCase_ModoAdaptativoLimitaQuantidadeDePistasFracas
// garante o parametro MaxWeakCues (default 3): com 5 pistas fracas, so
// as 3 piores entram no foco — senao a selecao viraria "qualquer item"
// e o efeito adaptativo desapareceria.
func TestItemSelectionUseCase_ModoAdaptativoLimitaQuantidadeDePistasFracas(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	mockStatsRepo := new(MockUserStatsRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo, mockStatsRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	expectedItem := &domain.Item{Id: uuid.New()}

	pior1 := uuid.New()
	pior2 := uuid.New()
	pior3 := uuid.New()

	mockStatsRepo.On("GetCueMastery", userID).Return([]repository.CueMastery{
		{CueId: uuid.New(), CueCode: "c4", Answered: 10, Correct: 6}, // 60%
		{CueId: pior1, CueCode: "c1", Answered: 10, Correct: 1},      // 10% (pior)
		{CueId: pior3, CueCode: "c3", Answered: 10, Correct: 3},      // 30%
		{CueId: uuid.New(), CueCode: "c5", Answered: 10, Correct: 7}, // 70%
		{CueId: pior2, CueCode: "c2", Answered: 10, Correct: 2},      // 20%
	}, nil)
	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(1), int64(1), nil)
	// Espera exatamente as 3 piores, na ordem crescente de acuracia.
	mockItemRepo.On("GetRandomUnseenByCues", userID, sessionID,
		[]uuid.UUID{pior1, pior2, pior3}, (*bool)(nil)).Return(expectedItem, nil)

	item, err := uc.NextItem(userID, sessionID, usecase.ModeAdaptive)

	assert.NoError(t, err)
	assert.Equal(t, expectedItem.Id, item.Id)
	mockItemRepo.AssertExpectations(t)
}

// TestItemSelectionUseCase_ModoAdaptativoRespeitaBalanceamentoDeLado
// garante que o modo adaptativo nao abandona o balanceamento: se o
// lado legitimo esta sub-representado na sessao, tenta primeiro um
// item legitimo que contenha a pista fraca. Sem isso, focar em pistas
// (mais comuns em itens maliciosos) enviesaria a sessao para "tudo e
// phishing", inflando acerto sem ganho de d'.
func TestItemSelectionUseCase_ModoAdaptativoRespeitaBalanceamentoDeLado(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	mockStatsRepo := new(MockUserStatsRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo, mockStatsRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	cueFraca := uuid.New()
	expectedItem := &domain.Item{Id: uuid.New(), IsMalicious: false}

	mockStatsRepo.On("GetCueMastery", userID).Return([]repository.CueMastery{
		{CueId: cueFraca, CueCode: "typosquat", Answered: 8, Correct: 2},
	}, nil)
	// 3 maliciosos vistos, 0 legitimos -> prefere legitimo.
	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(3), int64(0), nil)
	mockItemRepo.On("GetRandomUnseenByCues", userID, sessionID, []uuid.UUID{cueFraca},
		mock.MatchedBy(func(v *bool) bool { return v != nil && *v == false })).
		Return(expectedItem, nil)

	item, err := uc.NextItem(userID, sessionID, usecase.ModeAdaptive)

	assert.NoError(t, err)
	assert.Equal(t, expectedItem.Id, item.Id)
	mockItemRepo.AssertExpectations(t)
}

// TestItemSelectionUseCase_ModoAdaptativoIgnoraLadoQuandoEsgotado
// garante que, se nao existe item do lado preferido com a pista fraca,
// a pista tem prioridade sobre o balanceamento (tenta qualquer lado)
// antes de desistir do modo adaptativo.
func TestItemSelectionUseCase_ModoAdaptativoIgnoraLadoQuandoEsgotado(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	mockStatsRepo := new(MockUserStatsRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo, mockStatsRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	cueFraca := uuid.New()
	expectedItem := &domain.Item{Id: uuid.New(), IsMalicious: true}

	mockStatsRepo.On("GetCueMastery", userID).Return([]repository.CueMastery{
		{CueId: cueFraca, CueCode: "typosquat", Answered: 8, Correct: 2},
	}, nil)
	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(3), int64(0), nil)
	// Nao ha item legitimo com a pista fraca...
	mockItemRepo.On("GetRandomUnseenByCues", userID, sessionID, []uuid.UUID{cueFraca},
		mock.MatchedBy(func(v *bool) bool { return v != nil && *v == false })).
		Return(nil, gorm.ErrRecordNotFound)
	// ...entao aceita qualquer lado, desde que contenha a pista.
	mockItemRepo.On("GetRandomUnseenByCues", userID, sessionID, []uuid.UUID{cueFraca}, (*bool)(nil)).
		Return(expectedItem, nil)

	item, err := uc.NextItem(userID, sessionID, usecase.ModeAdaptive)

	assert.NoError(t, err)
	assert.Equal(t, expectedItem.Id, item.Id)
	mockItemRepo.AssertNotCalled(t, "GetRandomUnseen", mock.Anything, mock.Anything, mock.Anything)
}

// TestItemSelectionUseCase_ModoAdaptativoSemItemComPistaFracaCaiNoBalanceado
// garante que esgotar os itens com a pista fraca nao trava a sessao.
func TestItemSelectionUseCase_ModoAdaptativoSemItemComPistaFracaCaiNoBalanceado(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	mockStatsRepo := new(MockUserStatsRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo, mockStatsRepo)

	userID := uuid.New()
	sessionID := uuid.New()
	cueFraca := uuid.New()
	fallbackItem := &domain.Item{Id: uuid.New()}

	mockStatsRepo.On("GetCueMastery", userID).Return([]repository.CueMastery{
		{CueId: cueFraca, CueCode: "typosquat", Answered: 8, Correct: 2},
	}, nil)
	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(1), int64(1), nil)
	mockItemRepo.On("GetRandomUnseenByCues", userID, sessionID, []uuid.UUID{cueFraca}, (*bool)(nil)).
		Return(nil, gorm.ErrRecordNotFound)
	mockItemRepo.On("GetRandomUnseen", userID, sessionID, (*bool)(nil)).Return(fallbackItem, nil)

	item, err := uc.NextItem(userID, sessionID, usecase.ModeAdaptive)

	assert.NoError(t, err)
	assert.Equal(t, fallbackItem.Id, item.Id)
}

// TestItemSelectionUseCase_ModoAdaptativoPropagaErroRealDeBanco garante
// que uma falha real na leitura de mastery nao e mascarada como
// "usuario nao tem pista fraca" (o que silenciaria um problema de
// infraestrutura).
func TestItemSelectionUseCase_ModoAdaptativoPropagaErroRealDeBanco(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	mockStatsRepo := new(MockUserStatsRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo, mockStatsRepo)

	userID := uuid.New()
	sessionID := uuid.New()

	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(1), int64(1), nil)
	mockStatsRepo.On("GetCueMastery", userID).Return(nil, assert.AnError)

	item, err := uc.NextItem(userID, sessionID, usecase.ModeAdaptive)

	assert.Nil(t, item)
	assert.ErrorIs(t, err, assert.AnError)
	mockItemRepo.AssertNotCalled(t, "GetRandomUnseen", mock.Anything, mock.Anything, mock.Anything)
}

// TestItemSelectionUseCase_ModoBalanceadoNaoConsultaMastery garante que
// o modo default nao paga o custo da query de mastery.
func TestItemSelectionUseCase_ModoBalanceadoNaoConsultaMastery(t *testing.T) {
	mockItemRepo := new(MockItemRepository)
	mockStatsRepo := new(MockUserStatsRepository)
	uc := usecase.NewItemSelectionUseCase(mockItemRepo, mockStatsRepo)

	userID := uuid.New()
	sessionID := uuid.New()

	mockItemRepo.On("CountSeenInSession", userID, sessionID).Return(int64(1), int64(1), nil)
	mockItemRepo.On("GetRandomUnseen", userID, sessionID, (*bool)(nil)).
		Return(&domain.Item{Id: uuid.New()}, nil)

	_, err := uc.NextItem(userID, sessionID, usecase.ModeBalanced)

	assert.NoError(t, err)
	mockStatsRepo.AssertNotCalled(t, "GetCueMastery", mock.Anything)
}
