package repository

import (
	"phishing-quest/domain"

	"gorm.io/gorm"
)

type IItemRepository interface {
	IRepository[domain.Item]
	GetByChannel(channel domain.Channel) ([]*domain.Item, error)
}

type ItemRepository struct {
	IRepository[domain.Item]
	db *gorm.DB
}

func NewItemRepository(db *gorm.DB) IItemRepository {
	return &ItemRepository{
		IRepository: NewRepository[domain.Item](db),
		db:          db,
	}
}

func (ir *ItemRepository) GetByChannel(channel domain.Channel) ([]*domain.Item, error) {
	var items []*domain.Item
	if err := ir.db.Where("channel = ?", channel).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
