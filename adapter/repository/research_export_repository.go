package repository

import (
	"phishing-quest/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IResearchExportRepository fornece os dados brutos para o export de
// pesquisa. Nao embute IRepository[T]: a unica operacao e uma
// consulta agregada (JOIN attempts + items), sem CRUD.
type IResearchExportRepository interface {
	ListAttemptsWithItemInfo() ([]*domain.ResearchExportRow, error)
}

type ResearchExportRepository struct {
	db *gorm.DB
}

func NewResearchExportRepository(db *gorm.DB) IResearchExportRepository {
	return &ResearchExportRepository{db: db}
}

// exportScanRow espelha exatamente as colunas selecionadas no JOIN,
// com os tipos reais (uuid.UUID, time.Time) em vez de string — evita
// perder informacao ao converter para domain.ResearchExportRow.
type exportScanRow struct {
	UserId      uuid.UUID
	ItemId      uuid.UUID
	Channel     string
	IsMalicious bool
	SessionId   uuid.UUID
	Condition   string
	Verdict     *bool
	Action      string
	Confidence  *int
	IsCorrect   *bool
	LatencyMs   *int
	ClickedLink bool
	CreatedAt   time.Time
}

// ListAttemptsWithItemInfo retorna todas as tentativas com o canal e
// se o item era malicioso — JOIN necessario porque essas informacoes
// vivem em items, nao em attempts. PseudoUserId (hash) e calculado no
// usecase, nao aqui: o repository devolve o UserId real (em UserId,
// campo interno de exportScanRow), e a pseudonimizacao e
// responsabilidade explicita da camada de negocio.
func (rer *ResearchExportRepository) ListAttemptsWithItemInfo() ([]*domain.ResearchExportRow, error) {
	var rows []exportScanRow
	err := rer.db.Table("phishing_quest.attempts AS a").
		Select(`a.user_id, a.item_id, i.channel, i.is_malicious, a.session_id,
		        a.condition, a.verdict, a.action, a.confidence, a.is_correct,
		        a.latency_ms, a.clicked_link, a.created_at`).
		Joins("JOIN phishing_quest.items i ON i.id = a.item_id").
		Order("a.created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	exportRows := make([]*domain.ResearchExportRow, 0, len(rows))
	for _, r := range rows {
		exportRows = append(exportRows, &domain.ResearchExportRow{
			PseudoUserId: r.UserId.String(), // hash aplicado no usecase, sobrescrevendo este valor
			ItemId:       r.ItemId,
			Channel:      r.Channel,
			IsMalicious:  r.IsMalicious,
			SessionId:    r.SessionId,
			Condition:    r.Condition,
			Verdict:      r.Verdict,
			Action:       domain.AttemptAction(r.Action),
			Confidence:   r.Confidence,
			IsCorrect:    r.IsCorrect,
			LatencyMs:    r.LatencyMs,
			ClickedLink:  r.ClickedLink,
			CreatedAt:    r.CreatedAt,
		})
	}
	return exportRows, nil
}
