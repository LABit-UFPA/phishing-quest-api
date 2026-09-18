package tests

import (
	"testing"

	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockAssessmentRepository implementa repository.IAssessmentRepository.
type MockAssessmentRepository struct {
	mock.Mock
}

func (m *MockAssessmentRepository) Create(assessment *domain.Assessment) (*domain.Assessment, error) {
	args := m.Called(assessment)
	if err := args.Error(0); err != nil {
		return nil, err
	}
	return assessment, nil
}

func (m *MockAssessmentRepository) Update(assessment *domain.Assessment) (*domain.Assessment, error) {
	args := m.Called(assessment)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Assessment), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAssessmentRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAssessmentRepository) GetByID(id uuid.UUID) (*domain.Assessment, error) {
	args := m.Called(id)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Assessment), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAssessmentRepository) GetAll() ([]*domain.Assessment, error) {
	args := m.Called()
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.Assessment), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockAssessmentRepository) GetByUserAndPhase(userID uuid.UUID, phase domain.AssessmentPhase) (*domain.Assessment, error) {
	args := m.Called(userID, phase)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.Assessment), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestAssessmentUseCase_SubmitAssessment_Sucesso(t *testing.T) {
	mockRepo := new(MockAssessmentRepository)
	uc := usecase.NewAssessmentUseCase(mockRepo)

	userID := uuid.New()
	mockRepo.On("Create", mock.AnythingOfType("*domain.Assessment")).Return(nil)

	created, err := uc.SubmitAssessment(&domain.Assessment{
		UserId:            userID,
		Phase:             domain.PhasePre,
		InstrumentVersion: "A",
	})

	assert.NoError(t, err)
	assert.Equal(t, domain.PhasePre, created.Phase)
	assert.NotNil(t, created.FinishedAt)
	assert.NotEqual(t, uuid.Nil, created.Id)
	mockRepo.AssertExpectations(t)
}

func TestAssessmentUseCase_SubmitAssessment_RejeitaFaseInvalida(t *testing.T) {
	mockRepo := new(MockAssessmentRepository)
	uc := usecase.NewAssessmentUseCase(mockRepo)

	created, err := uc.SubmitAssessment(&domain.Assessment{
		UserId:            uuid.New(),
		Phase:             domain.AssessmentPhase("nao_existe"),
		InstrumentVersion: "A",
	})

	assert.Nil(t, created)
	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "Create", mock.Anything)
}

func TestAssessmentUseCase_GetAssessment_NaoEncontrado(t *testing.T) {
	mockRepo := new(MockAssessmentRepository)
	uc := usecase.NewAssessmentUseCase(mockRepo)

	userID := uuid.New()
	mockRepo.On("GetByUserAndPhase", userID, domain.PhaseDelayed4w).Return(nil, gorm.ErrRecordNotFound)

	assessment, err := uc.GetAssessment(userID, domain.PhaseDelayed4w)

	assert.Nil(t, assessment)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
