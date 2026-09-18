package repository

import (
	"phishing-quest/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IAttemptRepository interface {
	IRepository[domain.Attempt]
	GetByUserID(userID uuid.UUID) ([]*domain.Attempt, error)
}

type AttemptRepository struct {
	IRepository[domain.Attempt]
	db *gorm.DB
}

func NewAttemptRepository(db *gorm.DB) IAttemptRepository {
	return &AttemptRepository{
		IRepository: NewRepository[domain.Attempt](db),
		db:          db,
	}
}

func (ar *AttemptRepository) GetByUserID(userID uuid.UUID) ([]*domain.Attempt, error) {
	var attempts []*domain.Attempt
	if err := ar.db.Where("user_id = ?", userID).Find(&attempts).Error; err != nil {
		return nil, err
	}
	return attempts, nil
}
