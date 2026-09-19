package usecase

import (
	"phishing-quest/adapter/repository"
	"phishing-quest/core/service"
	"phishing-quest/domain"
	"time"

	"github.com/google/uuid"
)

// ItemReviewUseCase implementa o pipeline de curadoria de itens da
// issue #30: rascunho -> revisado -> publicado, com revisao humana
// obrigatoria antes da publicacao.
type ItemReviewUseCase struct {
	itemRepo     repository.IItemRepository
	draftService service.IItemDraftService
}

func NewItemReviewUseCase(itemRepo repository.IItemRepository, draftService service.IItemDraftService) *ItemReviewUseCase {
	return &ItemReviewUseCase{itemRepo: itemRepo, draftService: draftService}
}

// CreateDraft cria um item sempre em rascunho. Status, ReviewedBy,
// ReviewedAt e PublishedAt enviados pelo cliente sao IGNORADOS: o
// cliente nunca consegue nascer com um item ja publicado nem se
// declarar revisor de si mesmo (seria burlar o portao de revisao).
func (iruc *ItemReviewUseCase) CreateDraft(itemRequest *domain.Item) (*domain.Item, error) {
	item := &domain.Item{
		Id:                         uuid.New(),
		Channel:                    itemRequest.Channel,
		IsMalicious:                itemRequest.IsMalicious,
		Locale:                     itemRequest.Locale,
		PhishScaleCueCount:         itemRequest.PhishScaleCueCount,
		PhishScalePremiseAlignment: itemRequest.PhishScalePremiseAlignment,
		DifficultyEstimated:        itemRequest.DifficultyEstimated,
		DifficultyCalibrated:       itemRequest.DifficultyCalibrated,
		ContentJSON:                itemRequest.ContentJSON,
		Explanation:                itemRequest.Explanation,
		Source:                     itemRequest.Source,
		Status:                     domain.StatusDraft,
	}

	if item.Locale == "" {
		item.Locale = "pt-BR"
	}

	if err := item.Validate(); err != nil {
		return nil, err
	}

	return iruc.itemRepo.Create(item)
}

// GenerateDraft usa o servico de geracao (LLM ou template local) para
// montar o conteudo e persiste o resultado como rascunho. O item
// gerado NAO fica revisado nem publicado: continua sujeito ao mesmo
// portao de revisao humana de qualquer outro rascunho.
func (iruc *ItemReviewUseCase) GenerateDraft(spec service.DraftSpec) (*domain.Item, error) {
	draft, err := iruc.draftService.GenerateItemDraft(spec)
	if err != nil {
		return nil, err
	}
	return iruc.CreateDraft(draft)
}

// MarkReviewed registra a revisao humana. O reviewerID vem do JWT (o
// handler nao aceita revisor pelo corpo da requisicao), garantindo que
// a assinatura da revisao corresponde a quem esta autenticado.
func (iruc *ItemReviewUseCase) MarkReviewed(itemID, reviewerID uuid.UUID) (*domain.Item, error) {
	item, err := iruc.itemRepo.GetByID(itemID)
	if err != nil {
		return nil, err
	}

	if err := item.MarkReviewed(reviewerID, time.Now()); err != nil {
		return nil, err
	}

	return iruc.itemRepo.Update(item)
}

// Publish libera o item para o jogo. A transicao valida acontece no
// dominio (Item.Publish), que recusa publicar item nao revisado ou sem
// revisor registrado.
func (iruc *ItemReviewUseCase) Publish(itemID uuid.UUID) (*domain.Item, error) {
	item, err := iruc.itemRepo.GetByID(itemID)
	if err != nil {
		return nil, err
	}

	if err := item.Publish(time.Now()); err != nil {
		return nil, err
	}

	return iruc.itemRepo.Update(item)
}

// ListByStatus lista os itens de um estado do pipeline (ex.: fila de
// rascunhos aguardando revisao).
func (iruc *ItemReviewUseCase) ListByStatus(status domain.ItemStatus) ([]*domain.Item, error) {
	return iruc.itemRepo.GetByStatus(status)
}

// GetItem devolve o item independente do estado — uso administrativo,
// para o revisor inspecionar um rascunho.
func (iruc *ItemReviewUseCase) GetItem(itemID uuid.UUID) (*domain.Item, error) {
	return iruc.itemRepo.GetByID(itemID)
}
