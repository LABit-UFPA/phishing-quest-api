package dto

import "github.com/google/uuid"

// RankingEntryDTO e uma posicao no ranking global (ou por coorte).
// Substitui domain.UserScore como shape de resposta: a query anterior
// selecionava SUM(score) AS total_score mas mapeava direto em
// UserScore (cujo campo e Score), entao total_score nunca era lido —
// a resposta sempre saia com score:0.
type RankingEntryDTO struct {
	Position   int       `json:"position"`
	UserId     uuid.UUID `json:"userId"`
	Username   string    `json:"username"`
	TotalScore int       `json:"totalScore"`
}
