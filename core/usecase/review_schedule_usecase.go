package usecase

import (
	"phishing-quest/adapter/repository"
	"phishing-quest/domain"
	"time"

	"github.com/google/uuid"
)

type ReviewScheduleUseCase struct {
	scheduleRepo repository.IReviewScheduleRepository
	itemCueRepo  repository.IItemCueRepository
}

func NewReviewScheduleUseCase(scheduleRepo repository.IReviewScheduleRepository, itemCueRepo repository.IItemCueRepository) *ReviewScheduleUseCase {
	return &ReviewScheduleUseCase{scheduleRepo: scheduleRepo, itemCueRepo: itemCueRepo}
}

// RecordOutcome atualiza o agendamento Leitner de CADA pista presente
// no item respondido, com o mesmo resultado (correct) da tentativa.
// Chamado pelo fluxo de /attempts quando isCorrect esta disponivel —
// ver AttemptUseCase.RegisterAttempt.
func (rsu *ReviewScheduleUseCase) RecordOutcome(userID, itemID uuid.UUID, correct bool) error {
	itemCues, err := rsu.itemCueRepo.GetByItemID(itemID)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, itemCue := range itemCues {
		schedule, err := rsu.scheduleRepo.GetOrCreate(userID, itemID, itemCue.CueId)
		if err != nil {
			return err
		}
		schedule.ApplyResult(correct, now)
		if _, err := rsu.scheduleRepo.Update(schedule); err != nil {
			return err
		}
	}
	return nil
}

// GetDueReviews retorna os agendamentos vencidos do usuario (itens
// cuja pista esta devida para revisao agora).
func (rsu *ReviewScheduleUseCase) GetDueReviews(userID uuid.UUID, limit int) ([]*domain.ReviewSchedule, error) {
	return rsu.scheduleRepo.GetDue(userID, time.Now(), limit)
}
