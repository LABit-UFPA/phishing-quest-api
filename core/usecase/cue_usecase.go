package usecase

import (
	"phishing-quest/adapter/repository"
	"phishing-quest/domain"

	"github.com/google/uuid"
)

type CueUseCase struct {
	cueRepo     repository.ICueRepository
	itemCueRepo repository.IItemCueRepository
	itemRepo    repository.IItemRepository
}

func NewCueUseCase(cueRepo repository.ICueRepository, itemCueRepo repository.IItemCueRepository, itemRepo repository.IItemRepository) *CueUseCase {
	return &CueUseCase{cueRepo: cueRepo, itemCueRepo: itemCueRepo, itemRepo: itemRepo}
}

func (cuc *CueUseCase) ListCues() ([]*domain.Cue, error) {
	return cuc.cueRepo.GetAll()
}

func (cuc *CueUseCase) CreateCue(cueRequest *domain.Cue) (*domain.Cue, error) {
	cue := &domain.Cue{
		Id:       uuid.New(),
		Code:     cueRequest.Code,
		LabelPt:  cueRequest.LabelPt,
		Category: cueRequest.Category,
	}

	if err := cue.Validate(); err != nil {
		return nil, err
	}

	return cuc.cueRepo.Create(cue)
}

// AssociateItemCue liga um item a uma pista, validando que ambos
// existem antes de criar a associacao (as FKs do banco tambem
// garantem isso, mas validar aqui devolve um erro mais claro que a
// mensagem crua de violacao de foreign key).
func (cuc *CueUseCase) AssociateItemCue(itemCueRequest *domain.ItemCue) (*domain.ItemCue, error) {
	if _, err := cuc.itemRepo.GetByID(itemCueRequest.ItemId); err != nil {
		return nil, err
	}
	if _, err := cuc.cueRepo.GetByID(itemCueRequest.CueId); err != nil {
		return nil, err
	}

	itemCue := &domain.ItemCue{
		Id:        uuid.New(),
		ItemId:    itemCueRequest.ItemId,
		CueId:     itemCueRequest.CueId,
		SpanStart: itemCueRequest.SpanStart,
		SpanEnd:   itemCueRequest.SpanEnd,
	}

	if err := itemCue.Validate(); err != nil {
		return nil, err
	}

	return cuc.itemCueRepo.Create(itemCue)
}

func (cuc *CueUseCase) GetCuesByItemID(itemID uuid.UUID) ([]*domain.ItemCue, error) {
	return cuc.itemCueRepo.GetByItemID(itemID)
}
