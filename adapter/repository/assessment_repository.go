package repository

import (
	"phishing-quest/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IAssessmentRepository interface {
	IRepository[domain.Assessment]
	GetByUserAndPhase(userID uuid.UUID, phase domain.AssessmentPhase) (*domain.Assessment, error)
}

type AssessmentRepository struct {
	IRepository[domain.Assessment]
	db *gorm.DB
}

func NewAssessmentRepository(db *gorm.DB) IAssessmentRepository {
	return &AssessmentRepository{
		IRepository: NewRepository[domain.Assessment](db),
		db:          db,
	}
}

func (ar *AssessmentRepository) GetByUserAndPhase(userID uuid.UUID, phase domain.AssessmentPhase) (*domain.Assessment, error) {
	var assessment domain.Assessment
	if err := ar.db.Where("user_id = ? AND phase = ?", userID, phase).First(&assessment).Error; err != nil {
		return nil, err
	}
	return &assessment, nil
}
