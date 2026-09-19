package tests

import (
	"testing"

	"phishing-quest/core/service"
	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func newItemReviewUseCaseWithMocks() (*usecase.ItemReviewUseCase, *MockItemRepository) {
	mockRepo := new(MockItemRepository)
	uc := usecase.NewItemReviewUseCase(mockRepo, service.NewTemplateDraftService())
	return uc, mockRepo
}

// TestItemReviewUseCase_CreateDraft_NasceComoRascunho e a regressao
// central da issue #30: todo item criado pela API entra como rascunho,
// nunca publicado.
func TestItemReviewUseCase_CreateDraft_NasceComoRascunho(t *testing.T) {
	uc, mockRepo := newItemReviewUseCaseWithMocks()

	mockRepo.On("Create", mock.AnythingOfType("*domain.Item")).Return(nil)

	created, err := uc.CreateDraft(&domain.Item{
		Channel:     domain.ChannelWhatsApp,
		IsMalicious: true,
		ContentJSON: datatypes.JSON(`{"messages":[]}`),
	})

	assert.NoError(t, err)
	assert.Equal(t, domain.StatusDraft, created.Status)
	assert.Nil(t, created.ReviewedBy)
	assert.Nil(t, created.PublishedAt)
	assert.Equal(t, "pt-BR", created.Locale) // default de locale preservado
	mockRepo.AssertExpectations(t)
}

// TestItemReviewUseCase_CreateDraft_IgnoraStatusERevisorDoCliente e a
// protecao contra burlar o portao: mesmo mandando status=published e um
// reviewedBy no corpo, o item nasce rascunho e sem revisor.
func TestItemReviewUseCase_CreateDraft_IgnoraStatusERevisorDoCliente(t *testing.T) {
	uc, mockRepo := newItemReviewUseCaseWithMocks()

	mockRepo.On("Create", mock.AnythingOfType("*domain.Item")).Return(nil)

	falsoRevisor := uuid.New()
	created, err := uc.CreateDraft(&domain.Item{
		Channel:     domain.ChannelEmail,
		ContentJSON: datatypes.JSON(`{"subject":"x"}`),
		Status:      domain.StatusPublished,
		ReviewedBy:  &falsoRevisor,
	})

	assert.NoError(t, err)
	assert.Equal(t, domain.StatusDraft, created.Status)
	assert.Nil(t, created.ReviewedBy)
}

func TestItemReviewUseCase_CreateDraft_RejeitaCanalInvalido(t *testing.T) {
	uc, mockRepo := newItemReviewUseCaseWithMocks()

	created, err := uc.CreateDraft(&domain.Item{
		Channel:     domain.Channel("carta_pombo"),
		ContentJSON: datatypes.JSON(`{}`),
	})

	assert.Nil(t, created)
	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything)
}

// TestItemReviewUseCase_CreateDraft_MantemEstimadaECalibradaSeparadas e
// a regressao central da issue #67: DifficultyEstimated (a priori, do
// gerador) e DifficultyCalibrated (a posteriori, medida a partir de
// attempts reais -- issue #66) sao colunas distintas, e CreateDraft nao
// pode conflar as duas ao copiar do request para o dominio.
func TestItemReviewUseCase_CreateDraft_MantemEstimadaECalibradaSeparadas(t *testing.T) {
	uc, mockRepo := newItemReviewUseCaseWithMocks()

	mockRepo.On("Create", mock.AnythingOfType("*domain.Item")).Return(nil)

	estimada := "hard"
	calibrada := "medium"
	created, err := uc.CreateDraft(&domain.Item{
		Channel:              domain.ChannelEmail,
		ContentJSON:          datatypes.JSON(`{"subject":"x"}`),
		DifficultyEstimated:  &estimada,
		DifficultyCalibrated: &calibrada,
	})

	assert.NoError(t, err)
	assert.Equal(t, &estimada, created.DifficultyEstimated)
	assert.Equal(t, &calibrada, created.DifficultyCalibrated)
	assert.NotEqual(t, created.DifficultyEstimated, created.DifficultyCalibrated)
}

// TestItemReviewUseCase_FluxoCompleto_RascunhoRevisadoPublicado cobre o
// caminho felizes das 3 etapas do pipeline.
func TestItemReviewUseCase_FluxoCompleto_RascunhoRevisadoPublicado(t *testing.T) {
	uc, mockRepo := newItemReviewUseCaseWithMocks()

	itemID := uuid.New()
	reviewerID := uuid.New()
	draft := &domain.Item{Id: itemID, Status: domain.StatusDraft}

	mockRepo.On("GetByID", itemID).Return(draft, nil)
	mockRepo.On("Update", mock.AnythingOfType("*domain.Item")).Return(draft, nil)

	reviewed, err := uc.MarkReviewed(itemID, reviewerID)
	assert.NoError(t, err)
	assert.Equal(t, domain.StatusReviewed, reviewed.Status)
	assert.NotNil(t, reviewed.ReviewedBy)
	assert.Equal(t, reviewerID, *reviewed.ReviewedBy)
	assert.NotNil(t, reviewed.ReviewedAt)

	published, err := uc.Publish(itemID)
	assert.NoError(t, err)
	assert.Equal(t, domain.StatusPublished, published.Status)
	assert.NotNil(t, published.PublishedAt)
	// A assinatura de quem revisou sobrevive a publicacao.
	assert.Equal(t, reviewerID, *published.ReviewedBy)
}

// TestItemReviewUseCase_Publish_RecusaRascunhoNaoRevisado e o criterio
// de aceite da issue #30: item so entra em producao apos revisao humana
// registrada. Pular a revisao e recusado.
func TestItemReviewUseCase_Publish_RecusaRascunhoNaoRevisado(t *testing.T) {
	uc, mockRepo := newItemReviewUseCaseWithMocks()

	itemID := uuid.New()
	mockRepo.On("GetByID", itemID).Return(&domain.Item{Id: itemID, Status: domain.StatusDraft}, nil)

	published, err := uc.Publish(itemID)

	assert.Nil(t, published)
	assert.ErrorIs(t, err, domain.ErrInvalidItemTransition)
	mockRepo.AssertNotCalled(t, "Update", mock.Anything)
}

// TestItemReviewUseCase_Publish_RecusaRevisadoSemRevisor cobre a
// checagem redundante de defesa em profundidade: um item marcado como
// revisado, mas sem revisor registrado (ex.: dado corrompido no banco),
// nao pode ser publicado.
func TestItemReviewUseCase_Publish_RecusaRevisadoSemRevisor(t *testing.T) {
	uc, mockRepo := newItemReviewUseCaseWithMocks()

	itemID := uuid.New()
	mockRepo.On("GetByID", itemID).Return(&domain.Item{
		Id:         itemID,
		Status:     domain.StatusReviewed,
		ReviewedBy: nil,
	}, nil)

	published, err := uc.Publish(itemID)

	assert.Nil(t, published)
	assert.ErrorIs(t, err, domain.ErrReviewerRequired)
	mockRepo.AssertNotCalled(t, "Update", mock.Anything)
}

// TestItemReviewUseCase_MarkReviewed_RecusaItemJaPublicado evita que uma
// segunda revisao sobrescreva a assinatura da primeira.
func TestItemReviewUseCase_MarkReviewed_RecusaItemJaPublicado(t *testing.T) {
	uc, mockRepo := newItemReviewUseCaseWithMocks()

	itemID := uuid.New()
	revisorOriginal := uuid.New()
	mockRepo.On("GetByID", itemID).Return(&domain.Item{
		Id:         itemID,
		Status:     domain.StatusPublished,
		ReviewedBy: &revisorOriginal,
	}, nil)

	item, err := uc.MarkReviewed(itemID, uuid.New())

	assert.Nil(t, item)
	assert.ErrorIs(t, err, domain.ErrInvalidItemTransition)
	mockRepo.AssertNotCalled(t, "Update", mock.Anything)
}

func TestItemReviewUseCase_MarkReviewed_ItemInexistente(t *testing.T) {
	uc, mockRepo := newItemReviewUseCaseWithMocks()

	itemID := uuid.New()
	mockRepo.On("GetByID", itemID).Return(nil, gorm.ErrRecordNotFound)

	item, err := uc.MarkReviewed(itemID, uuid.New())

	assert.Nil(t, item)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// TestItemReviewUseCase_GenerateDraft_GeradoEntraComoRascunho garante
// que a geracao automatica nao encurta o pipeline: o item gerado ainda
// precisa de revisao humana.
func TestItemReviewUseCase_GenerateDraft_GeradoEntraComoRascunho(t *testing.T) {
	uc, mockRepo := newItemReviewUseCaseWithMocks()

	mockRepo.On("Create", mock.AnythingOfType("*domain.Item")).Return(nil)

	created, err := uc.GenerateDraft(service.DraftSpec{
		Channel:     domain.ChannelEmail,
		IsMalicious: true,
		Context:     "cobranca de fatura de energia",
	})

	assert.NoError(t, err)
	assert.Equal(t, domain.StatusDraft, created.Status)
	assert.Nil(t, created.ReviewedBy)
	// A procedencia fica registrada para o revisor saber que o texto nao
	// foi escrito por uma pessoa.
	assert.Equal(t, "llm_draft", created.Source)
	assert.Contains(t, string(created.ContentJSON), "cobranca de fatura de energia")
	mockRepo.AssertExpectations(t)
}

func TestItemReviewUseCase_GenerateDraft_RejeitaCanalNaoSuportado(t *testing.T) {
	uc, mockRepo := newItemReviewUseCaseWithMocks()

	created, err := uc.GenerateDraft(service.DraftSpec{Channel: domain.Channel("carta_pombo")})

	assert.Nil(t, created)
	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestItemReviewUseCase_ListByStatus(t *testing.T) {
	uc, mockRepo := newItemReviewUseCaseWithMocks()

	mockRepo.On("GetByStatus", domain.StatusDraft).Return([]*domain.Item{
		{Id: uuid.New(), Status: domain.StatusDraft},
	}, nil)

	items, err := uc.ListByStatus(domain.StatusDraft)

	assert.NoError(t, err)
	assert.Len(t, items, 1)
	mockRepo.AssertExpectations(t)
}
