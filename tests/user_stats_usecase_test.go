package tests

import (
	"math"
	"testing"

	"phishing-quest/adapter/repository"
	"phishing-quest/core/usecase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserStatsRepository implementa repository.IUserStatsRepository.
type MockUserStatsRepository struct {
	mock.Mock
}

func (m *MockUserStatsRepository) GetSignalDetectionAttempts(userID uuid.UUID) ([]repository.SignalDetectionAttempt, error) {
	args := m.Called(userID)
	if args.Get(0) != nil {
		return args.Get(0).([]repository.SignalDetectionAttempt), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserStatsRepository) GetCueMastery(userID uuid.UUID) ([]repository.CueMastery, error) {
	args := m.Called(userID)
	if args.Get(0) != nil {
		return args.Get(0).([]repository.CueMastery), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserStatsRepository) GetCueOutcomes(userID uuid.UUID) ([]repository.CueAttemptOutcome, error) {
	args := m.Called(userID)
	if args.Get(0) != nil {
		return args.Get(0).([]repository.CueAttemptOutcome), args.Error(1)
	}
	return nil, args.Error(1)
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) < tolerance
}

// TestUserStatsUseCase_CalculaDPrimeECriterion e a regressao central
// da issue #23: fixa um cenario calculado manualmente e validado
// contra Postgres real (8 tentativas: 3 hits + 1 miss em maliciosos,
// 3 rejeicoes corretas + 1 falso alarme em legitimos).
//
// hitRate=0.75, falseAlarmRate=0.25 (taxas brutas, sem correcao).
// Com a correcao log-linear de Hautus: hitAdj=(3+0.5)/(4+1)=0.7,
// faAdj=(1+0.5)/(4+1)=0.3. z(0.7)=+0.5244, z(0.3)=-0.5244.
// d' = z(hit)-z(fa) = 1.0488; c = -0.5*(z(hit)+z(fa)) = 0 (simetrico).
func TestUserStatsUseCase_CalculaDPrimeECriterion(t *testing.T) {
	mockRepo := new(MockUserStatsRepository)
	uc := usecase.NewUserStatsUseCase(mockRepo)

	userID := uuid.New()
	attempts := []repository.SignalDetectionAttempt{
		{ItemIsMalicious: true, Verdict: true},   // hit
		{ItemIsMalicious: true, Verdict: true},   // hit
		{ItemIsMalicious: true, Verdict: true},   // hit
		{ItemIsMalicious: true, Verdict: false},  // miss
		{ItemIsMalicious: false, Verdict: false}, // rejeicao correta
		{ItemIsMalicious: false, Verdict: false}, // rejeicao correta
		{ItemIsMalicious: false, Verdict: false}, // rejeicao correta
		{ItemIsMalicious: false, Verdict: true},  // falso alarme
	}
	mockRepo.On("GetSignalDetectionAttempts", userID).Return(attempts, nil)
	mockRepo.On("GetCueOutcomes", userID).Return([]repository.CueAttemptOutcome{}, nil)

	stats, err := uc.GetStats(userID)

	assert.NoError(t, err)
	assert.Equal(t, 8, stats.TotalAttempts)
	assert.InDelta(t, 0.75, stats.HitRate, 0.0001)
	assert.InDelta(t, 0.25, stats.FalseAlarmRate, 0.0001)
	assert.InDelta(t, 0.75, stats.Accuracy, 0.0001)
	assert.True(t, almostEqual(stats.DPrime, 1.0488, 0.001),
		"d' esperado ~1.0488, obtido %v", stats.DPrime)
	assert.True(t, almostEqual(stats.Criterion, 0.0, 0.001),
		"criterion esperado ~0 (simetrico), obtido %v", stats.Criterion)
}

// TestUserStatsUseCase_SemTentativas garante que um usuario sem
// nenhuma tentativa recebe zeros (nao NaN/Inf/panic).
func TestUserStatsUseCase_SemTentativas(t *testing.T) {
	mockRepo := new(MockUserStatsRepository)
	uc := usecase.NewUserStatsUseCase(mockRepo)

	userID := uuid.New()
	mockRepo.On("GetSignalDetectionAttempts", userID).Return([]repository.SignalDetectionAttempt{}, nil)
	mockRepo.On("GetCueOutcomes", userID).Return([]repository.CueAttemptOutcome{}, nil)

	stats, err := uc.GetStats(userID)

	assert.NoError(t, err)
	assert.Equal(t, 0, stats.TotalAttempts)
	assert.Equal(t, 0.0, stats.HitRate)
	assert.Equal(t, 0.0, stats.DPrime)
	assert.False(t, math.IsNaN(stats.DPrime))
	assert.False(t, math.IsInf(stats.DPrime, 0))
}

// TestUserStatsUseCase_AcertoPerfeitoNaoGeraInfinito e a regressao do
// caso extremo: sem a correcao de Hautus, hitRate=100% produziria
// z(1.0)=+Inf. Com a correcao, o resultado e finito.
func TestUserStatsUseCase_AcertoPerfeitoNaoGeraInfinito(t *testing.T) {
	mockRepo := new(MockUserStatsRepository)
	uc := usecase.NewUserStatsUseCase(mockRepo)

	userID := uuid.New()
	attempts := []repository.SignalDetectionAttempt{
		{ItemIsMalicious: true, Verdict: true},
		{ItemIsMalicious: true, Verdict: true},
	}
	mockRepo.On("GetSignalDetectionAttempts", userID).Return(attempts, nil)
	mockRepo.On("GetCueOutcomes", userID).Return([]repository.CueAttemptOutcome{}, nil)

	stats, err := uc.GetStats(userID)

	assert.NoError(t, err)
	assert.Equal(t, 1.0, stats.HitRate)
	assert.False(t, math.IsInf(stats.DPrime, 0))
	assert.False(t, math.IsNaN(stats.DPrime))
	// Validado contra Postgres real: d' ~0.9674 para este cenario exato.
	assert.True(t, almostEqual(stats.DPrime, 0.9674, 0.001))
}

func TestUserStatsUseCase_BreakdownPorPista(t *testing.T) {
	mockRepo := new(MockUserStatsRepository)
	uc := usecase.NewUserStatsUseCase(mockRepo)

	userID := uuid.New()
	mockRepo.On("GetSignalDetectionAttempts", userID).Return([]repository.SignalDetectionAttempt{}, nil)
	mockRepo.On("GetCueOutcomes", userID).Return([]repository.CueAttemptOutcome{
		{CueCode: "urgency", LabelPt: "Urgência", IsCorrect: true},
		{CueCode: "urgency", LabelPt: "Urgência", IsCorrect: true},
		{CueCode: "urgency", LabelPt: "Urgência", IsCorrect: false},
		{CueCode: "typosquat", LabelPt: "Typosquat", IsCorrect: false},
	}, nil)

	stats, err := uc.GetStats(userID)

	assert.NoError(t, err)
	assert.Len(t, stats.ByCue, 2)

	byCode := map[string]float64{}
	for _, cue := range stats.ByCue {
		byCode[cue.CueCode] = cue.Accuracy
	}
	assert.InDelta(t, 2.0/3.0, byCode["urgency"], 0.0001)
	assert.InDelta(t, 0.0, byCode["typosquat"], 0.0001)
}

func TestUserStatsUseCase_PropagaErroDoRepo(t *testing.T) {
	mockRepo := new(MockUserStatsRepository)
	uc := usecase.NewUserStatsUseCase(mockRepo)

	userID := uuid.New()
	mockRepo.On("GetSignalDetectionAttempts", userID).Return(nil, assert.AnError)

	stats, err := uc.GetStats(userID)

	assert.Nil(t, stats)
	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "GetCueOutcomes", mock.Anything)
}
