package usecase

import (
	"errors"
	"math/rand"
	"phishing-quest/adapter/repository"
	"phishing-quest/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ErrConsentRequired e retornado quando o usuario tenta registrar uma
// tentativa sem ter consentido em participar da coleta de dados.
var ErrConsentRequired = errors.New("usuario precisa registrar consentimento antes de participar da coleta de dados")

// experimentConditions e o desenho de 2 bracos do estudo principal
// (ROADMAP_PESQUISA_2027.md, RQ1): feedback formativo por justificativa
// vs. feedback binario simples. A atribuicao e aleatoria e feita pelo
// SERVIDOR no momento do consentimento — o participante nunca escolhe
// a propria condicao.
var experimentConditions = []string{"feedback_formativo", "feedback_binario"}

func assignCondition() string {
	return experimentConditions[rand.Intn(len(experimentConditions))]
}

type AttemptUseCase struct {
	attemptRepo      repository.IAttemptRepository
	itemRepo         repository.IItemRepository
	consentRepo      repository.IStudyParticipantRepository
	reviewScheduleUC *ReviewScheduleUseCase
}

func NewAttemptUseCase(
	attemptRepo repository.IAttemptRepository,
	itemRepo repository.IItemRepository,
	consentRepo repository.IStudyParticipantRepository,
	reviewScheduleUC *ReviewScheduleUseCase,
) *AttemptUseCase {
	return &AttemptUseCase{
		attemptRepo:      attemptRepo,
		itemRepo:         itemRepo,
		consentRepo:      consentRepo,
		reviewScheduleUC: reviewScheduleUC,
	}
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

	created, err := auc.attemptRepo.Create(attempt)
	if err != nil {
		return nil, err
	}

	// Atualiza a fila de revisao espacada (Leitner) para cada pista do
	// item respondido. Deliberadamente NAO propaga erro: a tentativa
	// ja foi gravada com sucesso (o dado de pesquisa esta seguro) e a
	// fila de revisao e uma funcionalidade secundaria — uma falha aqui
	// (ex.: item sem pistas anotadas ainda) nao deve fazer o usuario
	// perder a tentativa que acabou de submeter.
	if created.IsCorrect != nil && auc.reviewScheduleUC != nil {
		_ = auc.reviewScheduleUC.RecordOutcome(created.UserId, created.ItemId, *created.IsCorrect)
	}

	return created, nil
}

func (auc *AttemptUseCase) ListAttemptsByUser(userID uuid.UUID) ([]*domain.Attempt, error) {
	return auc.attemptRepo.GetByUserID(userID)
}

// ConsentRequest reune os campos aceitos no consentimento — definido
// no usecase (nao no dto) para o usecase nao depender do pacote dto,
// mantendo a mesma direcao de dependencia usada no resto do projeto
// (dto e consumido pelos handlers, nao pelos usecases).
type ConsentRequest struct {
	UserID           uuid.UUID
	ConsentVersion   string
	CohortID         *uuid.UUID
	DemographicsJSON datatypes.JSON
}

// RegisterConsent registra o consentimento do usuario para participar
// da coleta de dados, atribuindo uma condicao experimental aleatoria.
// Idempotente: se o usuario ja consentiu (e ainda ativo), retorna o
// registro existente em vez de tentar inserir de novo (o que violaria
// a PK user_id) ou de tratar qualquer erro do Create como "ja
// consentiu" — essa checagem explicita evita mascarar falhas reais
// (ex.: erro de conexao com o banco) como sucesso.
func (auc *AttemptUseCase) RegisterConsent(req ConsentRequest) (participant *domain.StudyParticipant, alreadyConsented bool, err error) {
	consented, err := auc.consentRepo.HasConsented(req.UserID)
	if err != nil {
		return nil, false, err
	}
	if consented {
		existing, err := auc.consentRepo.GetByUserID(req.UserID)
		if err != nil {
			return nil, false, err
		}
		return existing, true, nil
	}

	participant, err = auc.consentRepo.Create(&domain.StudyParticipant{
		UserId:           req.UserID,
		ConsentedAt:      time.Now(),
		CohortId:         req.CohortID,
		Condition:        assignCondition(),
		ConsentVersion:   req.ConsentVersion,
		DemographicsJSON: req.DemographicsJSON,
	})
	return participant, false, err
}

// WithdrawConsent registra a retirada de consentimento do participante
// (direito de exclusao, LGPD). Tentativas ja coletadas nao sao
// apagadas; apenas o participante e marcado inativo, o que bloqueia
// novas tentativas via HasConsented.
func (auc *AttemptUseCase) WithdrawConsent(userID uuid.UUID) error {
	return auc.consentRepo.Withdraw(userID)
}
