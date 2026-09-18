package repository

import (
	"phishing-quest/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IStudyParticipantRepository nao embute IRepository[T]: a chave
// primaria de StudyParticipant e UserId (nao um Id proprio), e a unica
// operacao necessaria hoje e "existe consentimento para este
// usuario?" + "registrar consentimento". Segue o mesmo padrao de
// IRankingRepository, que tambem define a interface do zero quando o
// CRUD generico nao se aplica bem.
type IStudyParticipantRepository interface {
	Create(participant *domain.StudyParticipant) (*domain.StudyParticipant, error)
	HasConsented(userID uuid.UUID) (bool, error)
}

type StudyParticipantRepository struct {
	db *gorm.DB
}

func NewStudyParticipantRepository(db *gorm.DB) IStudyParticipantRepository {
	return &StudyParticipantRepository{db: db}
}

func (spr *StudyParticipantRepository) Create(participant *domain.StudyParticipant) (*domain.StudyParticipant, error) {
	if err := spr.db.Create(participant).Error; err != nil {
		return nil, err
	}
	return participant, nil
}

func (spr *StudyParticipantRepository) HasConsented(userID uuid.UUID) (bool, error) {
	var count int64
	err := spr.db.Model(&domain.StudyParticipant{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
