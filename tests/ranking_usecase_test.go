package tests

import (
	"testing"

	"phishing-quest/core/usecase"
	"phishing-quest/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRankingRepository implementa repository.IRankingRepository.
type MockRankingRepository struct {
	mock.Mock
}

func (m *MockRankingRepository) GetGlobalRanking(limit, offset int) ([]dto.RankingEntryDTO, error) {
	args := m.Called(limit, offset)
	if args.Get(0) != nil {
		return args.Get(0).([]dto.RankingEntryDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRankingRepository) GetRankingByCohort(cohortID uuid.UUID, limit, offset int) ([]dto.RankingEntryDTO, error) {
	args := m.Called(cohortID, limit, offset)
	if args.Get(0) != nil {
		return args.Get(0).([]dto.RankingEntryDTO), args.Error(1)
	}
	return nil, args.Error(1)
}

// TestRankingUseCase_GetGlobalRanking_RetornaTotalScoreEPosition e a
// regressao da issue #24: a query anterior selecionava SUM(score) AS
// total_score mas mapeava direto em domain.UserScore (cujo campo e
// Score), entao a resposta sempre saia com score zerado e sem
// username/position. O DTO dedicado corrige isso.
func TestRankingUseCase_GetGlobalRanking_RetornaTotalScoreEPosition(t *testing.T) {
	mockRepo := new(MockRankingRepository)
	uc := usecase.NewRankingUseCase(mockRepo)

	expected := []dto.RankingEntryDTO{
		{Position: 1, UserId: uuid.New(), Username: "top_user", TotalScore: 100},
		{Position: 2, UserId: uuid.New(), Username: "second_user", TotalScore: 50},
	}
	mockRepo.On("GetGlobalRanking", 10, 0).Return(expected, nil)

	ranking, err := uc.GetGlobalRanking(10, 0)

	assert.NoError(t, err)
	assert.Len(t, ranking, 2)
	assert.Equal(t, 100, ranking[0].TotalScore)
	assert.Equal(t, "top_user", ranking[0].Username)
	assert.Equal(t, 1, ranking[0].Position)
	mockRepo.AssertExpectations(t)
}

func TestRankingUseCase_GetCohortRanking(t *testing.T) {
	mockRepo := new(MockRankingRepository)
	uc := usecase.NewRankingUseCase(mockRepo)

	cohortID := uuid.New()
	expected := []dto.RankingEntryDTO{
		{Position: 1, UserId: uuid.New(), Username: "in_cohort", TotalScore: 30},
	}
	mockRepo.On("GetRankingByCohort", cohortID, 10, 0).Return(expected, nil)

	ranking, err := uc.GetCohortRanking(cohortID, 10, 0)

	assert.NoError(t, err)
	assert.Len(t, ranking, 1)
	assert.Equal(t, "in_cohort", ranking[0].Username)
	mockRepo.AssertExpectations(t)
}

func TestRankingUseCase_PropagaErroDoRepo(t *testing.T) {
	mockRepo := new(MockRankingRepository)
	uc := usecase.NewRankingUseCase(mockRepo)

	mockRepo.On("GetGlobalRanking", 10, 0).Return(nil, assert.AnError)

	ranking, err := uc.GetGlobalRanking(10, 0)

	assert.Nil(t, ranking)
	assert.Error(t, err)
}
