package usecase

import (
	"phishing-quest/adapter/repository"
	"phishing-quest/dto"

	"github.com/google/uuid"
)

type RankingUseCase struct {
	rankingRepo repository.IRankingRepository
}

func NewRankingUseCase(rankingRepo repository.IRankingRepository) *RankingUseCase {
	return &RankingUseCase{rankingRepo: rankingRepo}
}

func (ruc *RankingUseCase) GetGlobalRanking(limit, offset int) ([]dto.RankingEntryDTO, error) {
	return ruc.rankingRepo.GetGlobalRanking(limit, offset)
}

// GetCohortRanking retorna o ranking restrito a uma coorte/turma.
func (ruc *RankingUseCase) GetCohortRanking(cohortID uuid.UUID, limit, offset int) ([]dto.RankingEntryDTO, error) {
	return ruc.rankingRepo.GetRankingByCohort(cohortID, limit, offset)
}
