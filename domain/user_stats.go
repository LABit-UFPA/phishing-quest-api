package domain

// CuePerformance resume o desempenho do usuario numa pista especifica
// (ex.: "usuario acerta 90% quando a pista e urgency mas so 40% quando
// e typosquat"), permitindo diagnostico por tipo de sinal em vez de
// so uma taxa agregada.
type CuePerformance struct {
	CueCode  string  `json:"cueCode"`
	LabelPt  string  `json:"labelPt"`
	Answered int     `json:"answered"`
	Correct  int     `json:"correct"`
	Accuracy float64 `json:"accuracy"`
}

// UserStats resume o desempenho do usuario via teoria de deteccao de
// sinal (ROADMAP_PESQUISA_2027.md, Fase 2 e 3): d' mede a capacidade
// de DISCRIMINAR itens maliciosos de legitimos (independente de
// vies); o criterio c mede o VIES de resposta (tendencia a dizer
// "e phishing" para tudo, o que infla a taxa de acerto bruta sem
// refletir discriminacao real). Substitui a agregacao simples por
// dificuldade/categoria que o front usava antes (so mock).
type UserStats struct {
	TotalAttempts  int              `json:"totalAttempts"`
	HitRate        float64          `json:"hitRate"`        // P(disse malicioso | era malicioso)
	FalseAlarmRate float64          `json:"falseAlarmRate"` // P(disse malicioso | era legitimo)
	Accuracy       float64          `json:"accuracy"`       // (hits + rejeicoes corretas) / total
	DPrime         float64          `json:"dPrime"`         // discriminacao (z(hit) - z(falseAlarm))
	Criterion      float64          `json:"criterion"`      // vies de resposta (-0.5*(z(hit)+z(falseAlarm)))
	ByCue          []CuePerformance `json:"byCue"`
}
