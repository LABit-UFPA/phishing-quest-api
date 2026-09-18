package repository

import (
	"phishing-quest/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IItemRepository interface {
	IRepository[domain.Item]
	GetByChannel(channel domain.Channel) ([]*domain.Item, error)

	// CountSeenInSession conta, dentre os items ja respondidos (via
	// attempts) pelo usuario na sessao informada, quantos eram
	// maliciosos e quantos eram legitimos. Usado para balancear a
	// selecao de GET /items/next.
	CountSeenInSession(userID, sessionID uuid.UUID) (maliciousSeen, legitimateSeen int64, err error)

	// GetRandomUnseen retorna um item aleatorio que o usuario ainda
	// nao respondeu na sessao informada, opcionalmente filtrado por
	// IsMalicious. isMalicious == nil significa "qualquer".
	GetRandomUnseen(userID, sessionID uuid.UUID, isMalicious *bool) (*domain.Item, error)

	// GetRandomUnseenByCues e igual ao GetRandomUnseen, mas restrito a
	// items que contenham AO MENOS UMA das pistas informadas. Usado
	// pela selecao adaptativa (issue #28) para concentrar a pratica nas
	// pistas que o usuario ainda erra.
	GetRandomUnseenByCues(userID, sessionID uuid.UUID, cueIDs []uuid.UUID, isMalicious *bool) (*domain.Item, error)
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

func (ir *ItemRepository) CountSeenInSession(userID, sessionID uuid.UUID) (maliciousSeen, legitimateSeen int64, err error) {
	base := ir.db.Table("phishing_quest.attempts AS a").
		Joins("JOIN phishing_quest.items i ON i.id = a.item_id").
		Where("a.user_id = ? AND a.session_id = ?", userID, sessionID)

	if err = base.Session(&gorm.Session{}).Where("i.is_malicious = TRUE").Count(&maliciousSeen).Error; err != nil {
		return 0, 0, err
	}
	if err = base.Session(&gorm.Session{}).Where("i.is_malicious = FALSE").Count(&legitimateSeen).Error; err != nil {
		return 0, 0, err
	}
	return maliciousSeen, legitimateSeen, nil
}

func (ir *ItemRepository) GetRandomUnseen(userID, sessionID uuid.UUID, isMalicious *bool) (*domain.Item, error) {
	query := ir.db.Model(&domain.Item{}).
		Where(`id NOT IN (
			SELECT item_id FROM phishing_quest.attempts
			WHERE user_id = ? AND session_id = ?
		)`, userID, sessionID)

	if isMalicious != nil {
		query = query.Where("is_malicious = ?", *isMalicious)
	}

	var item domain.Item
	if err := query.Order("RANDOM()").First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (ir *ItemRepository) GetRandomUnseenByCues(userID, sessionID uuid.UUID, cueIDs []uuid.UUID, isMalicious *bool) (*domain.Item, error) {
	// Sem pistas para focar nao existe candidato adaptativo. Retornar
	// ErrRecordNotFound (em vez de rodar um IN vazio, que o Postgres
	// rejeita) deixa o usecase cair no fallback balanceado.
	if len(cueIDs) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	query := ir.db.Model(&domain.Item{}).
		Where(`id NOT IN (
			SELECT item_id FROM phishing_quest.attempts
			WHERE user_id = ? AND session_id = ?
		)`, userID, sessionID).
		Where(`id IN (
			SELECT item_id FROM phishing_quest.item_cues
			WHERE cue_id IN ?
		)`, cueIDs)

	if isMalicious != nil {
		query = query.Where("is_malicious = ?", *isMalicious)
	}

	var item domain.Item
	if err := query.Order("RANDOM()").First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
