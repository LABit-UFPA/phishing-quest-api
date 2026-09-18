package usecase

import (
	"phishing-quest/adapter/repository"
	"phishing-quest/domain"
	"time"

	"github.com/google/uuid"
)

type AssessmentUseCase struct {
	assessmentRepo repository.IAssessmentRepository
}

func NewAssessmentUseCase(assessmentRepo repository.IAssessmentRepository) *AssessmentUseCase {
	return &AssessmentUseCase{assessmentRepo: assessmentRepo}
}

// SubmitAssessment registra as respostas de um instrumento para uma
// fase especifica (pre/post/delayed_4w). Phase e InstrumentVersion vem
// do path/request; CreatedAt e FinishedAt sao definidos pelo servidor.
func (auc *AssessmentUseCase) SubmitAssessment(assessmentRequest *domain.Assessment) (*domain.Assessment, error) {
	now := time.Now()
	assessment := &domain.Assessment{
		Id:                uuid.New(),
		UserId:            assessmentRequest.UserId,
		Phase:             assessmentRequest.Phase,
		InstrumentVersion: assessmentRequest.InstrumentVersion,
		StartedAt:         assessmentRequest.StartedAt,
		FinishedAt:        &now,
		ResponsesJSON:     assessmentRequest.ResponsesJSON,
		CreatedAt:         now,
	}

	if err := assessment.Validate(); err != nil {
		return nil, err
	}

	return auc.assessmentRepo.Create(assessment)
}

func (auc *AssessmentUseCase) GetAssessment(userID uuid.UUID, phase domain.AssessmentPhase) (*domain.Assessment, error) {
	return auc.assessmentRepo.GetByUserAndPhase(userID, phase)
}
