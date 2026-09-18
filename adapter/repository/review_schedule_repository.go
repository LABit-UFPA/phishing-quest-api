package repository

import (
	"phishing-quest/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IReviewScheduleRepository interface {
	// GetOrCreate busca o agendamento existente para user+item+cue,
	// ou cria um novo na caixa 1 (due imediatamente) se nao existir.
	GetOrCreate(userID, itemID, cueID uuid.UUID) (*domain.ReviewSchedule, error)
	Update(schedule *domain.ReviewSchedule) (*domain.ReviewSchedule, error)
	// GetDue retorna os agendamentos do usuario cujo due_at ja
	// passou (ou e agora), ordenados pelos mais atrasados primeiro.
	GetDue(userID uuid.UUID, now time.Time, limit int) ([]*domain.ReviewSchedule, error)
}

type ReviewScheduleRepository struct {
	db *gorm.DB
}

func NewReviewScheduleRepository(db *gorm.DB) IReviewScheduleRepository {
	return &ReviewScheduleRepository{db: db}
}

func (rsr *ReviewScheduleRepository) GetOrCreate(userID, itemID, cueID uuid.UUID) (*domain.ReviewSchedule, error) {
	var existing domain.ReviewSchedule
	err := rsr.db.Where("user_id = ? AND item_id = ? AND cue_id = ?", userID, itemID, cueID).
		First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	now := time.Now()
	created := &domain.ReviewSchedule{
		Id:        uuid.New(),
		UserId:    userID,
		ItemId:    itemID,
		CueId:     cueID,
		Box:       domain.MinLeitnerBox,
		DueAt:     now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := rsr.db.Create(created).Error; err != nil {
		return nil, err
	}
	return created, nil
}

func (rsr *ReviewScheduleRepository) Update(schedule *domain.ReviewSchedule) (*domain.ReviewSchedule, error) {
	if err := rsr.db.Save(schedule).Error; err != nil {
		return nil, err
	}
	return schedule, nil
}

func (rsr *ReviewScheduleRepository) GetDue(userID uuid.UUID, now time.Time, limit int) ([]*domain.ReviewSchedule, error) {
	var schedules []*domain.ReviewSchedule
	err := rsr.db.
		Where("user_id = ? AND due_at <= ?", userID, now).
		Order("due_at ASC").
		Limit(limit).
		Find(&schedules).Error
	if err != nil {
		return nil, err
	}
	return schedules, nil
}
