package usecase

import (
	"errors"
	"phishing-quest/adapter/repository"
	"phishing-quest/domain"
	"time"

	"github.com/google/uuid"
)

// ErrConsentRequired e retornado quando o usuario tenta registrar uma
// tentativa sem ter consentido em participar da coleta de dados.
var ErrConsentRequired = errors.New("usuario precisa registrar consentimento antes de participar da coleta de dados")

type AttemptUseCase struct {
	attemptRepo repository.IAttemptRepository
	itemRepo    repository.IItemRepository
	consentRepo repository.IStudyParticipantRepository
}

func NewAttemptUseCase(attemptRepo repository.IAttemptRepository, itemRepo repository.IItemRepository, consentRepo repository.IStudyParticipantRepository) *AttemptUseCase {
	return &AttemptUseCase{attemptRepo: attemptRepo, itemRepo: itemRepo, consentRepo: consentRepo}
}

// RegisterAttempt valida consentimento e a existencia do item antes de
// persistir a tentativa. Calcula IsCorrect a partir de Item.IsMalicious
// e Verdict quando o veredito binario foi informado — a decisao NUNCA
// e confiada ao cliente, sempre recalculada aqui a partir da fonte de
// verdade (Item.IsMalicious).
func (auc *AttemptUseCase) RegisterAttempt(attemptRequest *domain.Attempt) (*domain.Attempt, error) {
	consented, err := auc.consentRepo.HasConsented(attemptRequest.UserId)
	if err != nil {
		return nil, err
	}
	if !consented {
		return nil, ErrConsentRequired
	}

	item, err := auc.itemRepo.GetByID(attemptRequest.ItemId)
	if err != nil {
		return nil, err
	}

	attempt := &domain.Attempt{
		Id:              uuid.New(),
		UserId:          attemptRequest.UserId,
		ItemId:          attemptRequest.ItemId,
		SessionId:       attemptRequest.SessionId,
		Condition:       attemptRequest.Condition,
		Verdict:         attemptRequest.Verdict,
		Action:          attemptRequest.Action,
		Confidence:      attemptRequest.Confidence,
		Justification:   attemptRequest.Justification,
		LlmRating:       attemptRequest.LlmRating,
		LlmFeedbackJSON: attemptRequest.LlmFeedbackJSON,
		LatencyMs:       attemptRequest.LatencyMs,
		ClickedLink:     attemptRequest.ClickedLink,
		CreatedAt:       time.Now(),
	}

	if attempt.Verdict != nil {
		isCorrect := *attempt.Verdict == item.IsMalicious
		attempt.IsCorrect = &isCorrect
	}

	if err := attempt.Validate(); err != nil {
		return nil, err
	}

	return auc.attemptRepo.Create(attempt)
}

func (auc *AttemptUseCase) ListAttemptsByUser(userID uuid.UUID) ([]*domain.Attempt, error) {
	return auc.attemptRepo.GetByUserID(userID)
}

// RegisterConsent registra o consentimento do usuario para participar
// da coleta de dados. Idempotente: se o usuario ja consentiu, retorna
// o registro existente em vez de tentar inserir de novo (o que
// violaria a PK user_id) ou de tratar qualquer erro do Create como
// "ja consentiu" — essa checagem explicita evita mascarar falhas
// reais (ex.: erro de conexao com o banco) como sucesso.
func (auc *AttemptUseCase) RegisterConsent(userID uuid.UUID) (participant *domain.StudyParticipant, alreadyConsented bool, err error) {
	consented, err := auc.consentRepo.HasConsented(userID)
	if err != nil {
		return nil, false, err
	}
	if consented {
		return &domain.StudyParticipant{UserId: userID}, true, nil
	}

	participant, err = auc.consentRepo.Create(&domain.StudyParticipant{
		UserId:      userID,
		ConsentedAt: time.Now(),
	})
	return participant, false, err
}
