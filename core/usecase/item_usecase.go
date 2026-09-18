package usecase

import (
	"phishing-quest/adapter/repository"
	"phishing-quest/domain"

	"github.com/google/uuid"
)

type ItemUseCase struct {
	itemRepo repository.IItemRepository
}

func NewItemUseCase(itemRepo repository.IItemRepository) *ItemUseCase {
	return &ItemUseCase{itemRepo: itemRepo}
}

func (iuc *ItemUseCase) CreateItem(itemRequest *domain.Item) (*domain.Item, error) {
	item := &domain.Item{
		Id:                         uuid.New(),
		Channel:                    itemRequest.Channel,
		IsMalicious:                itemRequest.IsMalicious,
		Locale:                     itemRequest.Locale,
		PhishScaleCueCount:         itemRequest.PhishScaleCueCount,
		PhishScalePremiseAlignment: itemRequest.PhishScalePremiseAlignment,
		DifficultyCalibrated:       itemRequest.DifficultyCalibrated,
		ContentJSON:                itemRequest.ContentJSON,
		Explanation:                itemRequest.Explanation,
		Source:                     itemRequest.Source,
		ReviewedBy:                 itemRequest.ReviewedBy,
	}

	if item.Locale == "" {
		item.Locale = "pt-BR"
	}

	if err := item.Validate(); err != nil {
		return nil, err
	}

	return iuc.itemRepo.Create(item)
}

func (iuc *ItemUseCase) GetItem(id uuid.UUID) (*domain.Item, error) {
	return iuc.itemRepo.GetByID(id)
}

func (iuc *ItemUseCase) ListItems() ([]*domain.Item, error) {
	return iuc.itemRepo.GetAll()
}

func (iuc *ItemUseCase) ListItemsByChannel(channel domain.Channel) ([]*domain.Item, error) {
	return iuc.itemRepo.GetByChannel(channel)
}
