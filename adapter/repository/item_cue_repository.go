package repository

import (
	"phishing-quest/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IItemCueRepository interface {
	IRepository[domain.ItemCue]
	GetByItemID(itemID uuid.UUID) ([]*domain.ItemCue, error)
	GetByCueID(cueID uuid.UUID) ([]*domain.ItemCue, error)
}

type ItemCueRepository struct {
	IRepository[domain.ItemCue]
	db *gorm.DB
}

func NewItemCueRepository(db *gorm.DB) IItemCueRepository {
	return &ItemCueRepository{
		IRepository: NewRepository[domain.ItemCue](db),
		db:          db,
	}
}

func (icr *ItemCueRepository) GetByItemID(itemID uuid.UUID) ([]*domain.ItemCue, error) {
	var itemCues []*domain.ItemCue
	if err := icr.db.Where("item_id = ?", itemID).Find(&itemCues).Error; err != nil {
		return nil, err
	}
	return itemCues, nil
}

func (icr *ItemCueRepository) GetByCueID(cueID uuid.UUID) ([]*domain.ItemCue, error) {
	var itemCues []*domain.ItemCue
	if err := icr.db.Where("cue_id = ?", cueID).Find(&itemCues).Error; err != nil {
		return nil, err
	}
	return itemCues, nil
}
