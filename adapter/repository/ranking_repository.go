package repository

import (
	"phishing-quest/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IRankingRepository define os métodos do repositório de ranking
type IRankingRepository interface {
	GetGlobalRanking(limit int, offset int) ([]dto.RankingEntryDTO, error)
	GetRankingByCohort(cohortID uuid.UUID, limit int, offset int) ([]dto.RankingEntryDTO, error)
}

// RankingRepository implementa IRankingRepository
type RankingRepository struct {
	db *gorm.DB
}

// NewRankingRepository cria uma nova instância de RankingRepository
func NewRankingRepository(db *gorm.DB) IRankingRepository {
	return &RankingRepository{
		db: db,
	}
}

// rankingQuery monta a base compartilhada por ranking global e por
// coorte: soma o score de cada usuario (via user_scores, que e um
// ledger de eventos — ver adapter/repository/user_score_repository.go),
// junta com users para trazer o username, e calcula a posicao via
// ROW_NUMBER() ordenado por score desc. filterSQL/filterArgs permitem
// escopar por coorte sem duplicar a query inteira.
func (rr *RankingRepository) rankingQuery(filterSQL string, filterArgs []interface{}, limit, offset int) ([]dto.RankingEntryDTO, error) {
	query := `
		SELECT
			ROW_NUMBER() OVER (ORDER BY COALESCE(SUM(us.score), 0) DESC) AS position,
			u.id AS user_id,
			u.username AS username,
			COALESCE(SUM(us.score), 0) AS total_score
		FROM phishing_quest.users u
		LEFT JOIN phishing_quest.user_scores us ON us.user_id = u.id
	`
	if filterSQL != "" {
		query += " JOIN phishing_quest.study_participants sp ON sp.user_id = u.id AND " + filterSQL
	}
	query += `
		GROUP BY u.id, u.username
		ORDER BY total_score DESC
		LIMIT ? OFFSET ?
	`

	args := append(filterArgs, limit, offset)

	var entries []dto.RankingEntryDTO
	if err := rr.db.Raw(query, args...).Scan(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

// GetGlobalRanking retorna o ranking global ordenado pela pontuação total
func (rr *RankingRepository) GetGlobalRanking(limit int, offset int) ([]dto.RankingEntryDTO, error) {
	return rr.rankingQuery("", nil, limit, offset)
}

// GetRankingByCohort retorna o ranking restrito aos usuarios de uma
// coorte/turma especifica (study_participants.cohort_id). Leaderboard
// global desmotiva quem esta embaixo; comparacao por coorte funciona
// melhor (ROADMAP_PESQUISA_2027.md, mecanicas pedagogicas).
func (rr *RankingRepository) GetRankingByCohort(cohortID uuid.UUID, limit int, offset int) ([]dto.RankingEntryDTO, error) {
	return rr.rankingQuery("sp.cohort_id = ?", []interface{}{cohortID}, limit, offset)
}
