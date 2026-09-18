package repository

import (
	"phishing-quest/domain"

	"gorm.io/gorm"
)

type ICueRepository interface {
	IRepository[domain.Cue]
	GetByCode(code string) (*domain.Cue, error)
}

type CueRepository struct {
	IRepository[domain.Cue]
	db *gorm.DB
}

func NewCueRepository(db *gorm.DB) ICueRepository {
	return &CueRepository{
		IRepository: NewRepository[domain.Cue](db),
		db:          db,
	}
}

func (cr *CueRepository) GetByCode(code string) (*domain.Cue, error) {
	var cue domain.Cue
	if err := cr.db.Where("code = ?", code).First(&cue).Error; err != nil {
		return nil, err
	}
	return &cue, nil
}
