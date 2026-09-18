package domain

import (
	"time"

	"github.com/google/uuid"
)

// ResearchExportRow e uma linha do export de pesquisa: uma tentativa
// (attempt) com dados suficientes para a analise estatistica, mas
// PSEUDONIMIZADA — sem username, email ou justificativa em texto
// livre (que o participante poderia usar para se identificar
// involuntariamente).
//
// PseudoUserId substitui o UserId real: e um hash estavel (nao
// reversivel sem a tabela de usuarios), suficiente para agrupar
// tentativas do mesmo participante ao longo do tempo sem expor a
// identidade real no arquivo exportado.
type ResearchExportRow struct {
	PseudoUserId string
	ItemId       uuid.UUID
	Channel      string
	IsMalicious  bool
	SessionId    uuid.UUID
	Condition    string
	Verdict      *bool
	Action       AttemptAction
	Confidence   *int
	IsCorrect    *bool
	LatencyMs    *int
	ClickedLink  bool
	CreatedAt    time.Time
}
