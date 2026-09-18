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
	GetByUserID(userID uuid.UUID) (*domain.StudyParticipant, error)
	Withdraw(userID uuid.UUID) error
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

// HasConsented verifica se existe consentimento ATIVO (nao retirado)
// para o usuario. Um participante que exerceu o direito de exclusao
// (withdrawn_at preenchido) e tratado como sem consentimento.
func (spr *StudyParticipantRepository) HasConsented(userID uuid.UUID) (bool, error) {
	var count int64
	err := spr.db.Model(&domain.StudyParticipant{}).
		Where("user_id = ? AND withdrawn_at IS NULL", userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (spr *StudyParticipantRepository) GetByUserID(userID uuid.UUID) (*domain.StudyParticipant, error) {
	var participant domain.StudyParticipant
	if err := spr.db.Where("user_id = ?", userID).First(&participant).Error; err != nil {
		return nil, err
	}
	return &participant, nil
}

// Withdraw registra a retirada de consentimento (direito de exclusao,
// LGPD). Nao apaga o registro nem as tentativas ja coletadas — apenas
// marca o participante como inativo, impedindo novas tentativas.
func (spr *StudyParticipantRepository) Withdraw(userID uuid.UUID) error {
	result := spr.db.Model(&domain.StudyParticipant{}).
		Where("user_id = ?", userID).
		Update("withdrawn_at", gorm.Expr("NOW()"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
