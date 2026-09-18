package repository

import (
	"phishing-quest/domain"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type IUserScoreRepository interface {
	IRepository[domain.UserScore]
	IncrementScore(userID uuid.UUID, points int) error
	GetUserScore(userID uuid.UUID) (*domain.UserScore, error)
}

type UserScoreRepository struct {
	IRepository[domain.UserScore]
	db *gorm.DB
}

func NewUserScoreRepository(db *gorm.DB) IUserScoreRepository {
	return &UserScoreRepository{
		IRepository: NewRepository[domain.UserScore](db),
		db:          db,
	}
}

// IncrementScore registra um novo evento de pontuacao para o usuario.
//
// phishing_quest.user_scores e um ledger (uma linha por evento), nao um
// contador por usuario — o ranking (ranking_repository.go) agrega com
// "SUM(score) GROUP BY user_id", e a propria migration documenta uma
// VIEW global_ranking baseada nesse SUM (domain/user_score.go). Por
// isso a implementacao correta e INSERT de uma nova linha, nao UPDATE:
// a versao anterior fazia "UPDATE ... WHERE user_id = ?" e retornava
// gorm.ErrRecordNotFound (500) sempre que o usuario ainda nao tinha
// nenhuma linha — ou seja, no primeiro acerto de qualquer usuario.
func (r *UserScoreRepository) IncrementScore(userID uuid.UUID, points int) error {
	logrus.Infof("Registrando evento de pontuação para o usuário com ID: %s, pontos: %d", userID, points)
	event := &domain.UserScore{
		Id:     uuid.New(),
		UserId: userID,
		Score:  points,
	}
	if err := r.db.Create(event).Error; err != nil {
		logrus.Errorf("Erro ao registrar pontuação para o usuário com ID %s: %v", userID, err)
		return err
	}
	logrus.Infof("Pontuação registrada com sucesso para o usuário com ID: %s", userID)
	return nil
}

// GetUserScore retorna o total acumulado do usuario (SUM de todos os
// eventos em user_scores), nao um unico evento.
func (r *UserScoreRepository) GetUserScore(userID uuid.UUID) (*domain.UserScore, error) {
	logrus.Infof("Buscando pontuação total para o usuário com ID: %s", userID)
	var total int
	err := r.db.Model(&domain.UserScore{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(score), 0)").
		Scan(&total).Error
	if err != nil {
		logrus.Errorf("Erro ao buscar pontuação para o usuário com ID %s: %v", userID, err)
		return nil, err
	}
	logrus.Infof("Pontuação total encontrada para o usuário com ID: %s: %d", userID, total)
	return &domain.UserScore{UserId: userID, Score: total}, nil
}
