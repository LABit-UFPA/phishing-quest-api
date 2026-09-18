package tests

import (
	"testing"

	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTelemetryEventRepository implementa repository.ITelemetryEventRepository.
type MockTelemetryEventRepository struct {
	mock.Mock
}

func (m *MockTelemetryEventRepository) CreateBatch(events []*domain.TelemetryEvent) error {
	args := m.Called(events)
	return args.Error(0)
}

func TestTelemetryUseCase_IngestBatch_RejeitaLoteVazio(t *testing.T) {
	mockRepo := new(MockTelemetryEventRepository)
	uc := usecase.NewTelemetryUseCase(mockRepo)

	err := uc.IngestBatch([]*domain.TelemetryEvent{})

	assert.ErrorIs(t, err, usecase.ErrEmptyBatch)
	mockRepo.AssertNotCalled(t, "CreateBatch", mock.Anything)
}

func TestTelemetryUseCase_IngestBatch_RejeitaEventoSemEventType(t *testing.T) {
	mockRepo := new(MockTelemetryEventRepository)
	uc := usecase.NewTelemetryUseCase(mockRepo)

	err := uc.IngestBatch([]*domain.TelemetryEvent{
		{Id: uuid.New(), EventType: "screen_opened"},
		{Id: uuid.New()}, // sem EventType
	})

	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "CreateBatch", mock.Anything)
}

func TestTelemetryUseCase_IngestBatch_Sucesso(t *testing.T) {
	mockRepo := new(MockTelemetryEventRepository)
	uc := usecase.NewTelemetryUseCase(mockRepo)

	events := []*domain.TelemetryEvent{
		{Id: uuid.New(), EventType: "screen_opened"},
		{Id: uuid.New(), EventType: "review_notification_answered"},
	}
	mockRepo.On("CreateBatch", events).Return(nil)

	err := uc.IngestBatch(events)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestTelemetryUseCase_IngestBatch_PropagaErroDoRepo garante que uma
// falha real do repositorio (ex.: erro de conexao) nao e engolida.
func TestTelemetryUseCase_IngestBatch_PropagaErroDoRepo(t *testing.T) {
	mockRepo := new(MockTelemetryEventRepository)
	uc := usecase.NewTelemetryUseCase(mockRepo)

	events := []*domain.TelemetryEvent{{Id: uuid.New(), EventType: "screen_opened"}}
	mockRepo.On("CreateBatch", events).Return(assert.AnError)

	err := uc.IngestBatch(events)

	assert.Error(t, err)
}
