//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestIntegration_Stats_DPrimeBateComCalculoManual e a regressao da
// issue #23 com o caminho completo (SQL de agregacao + calculo de
// deteccao de sinal), usando o mesmo cenario conferido a mao:
// 3 hits + 1 miss em 4 maliciosos, 1 falso alarme + 3 rejeicoes
// corretas em 4 legitimos.
//
// Com a correcao log-linear de Hautus:
//
//	hitAdj = (3+0.5)/(4+1) = 0.7   -> z = +0.5244
//	faAdj  = (1+0.5)/(4+1) = 0.3   -> z = -0.5244
//	d' = 0.5244 - (-0.5244) = 1.0488 ; c = 0
func TestIntegration_Stats_DPrimeBateComCalculoManual(t *testing.T) {
	resetDB(t)

	userID, token := registerAndLogin(t, "dprime", "dprime@example.com")
	consent(t, userID)

	sessionID := uuid.NewString()

	// 4 itens maliciosos: 3 identificados (hit), 1 perdido (miss).
	submitJudgement(t, token, userID, sessionID, true, true)
	submitJudgement(t, token, userID, sessionID, true, true)
	submitJudgement(t, token, userID, sessionID, true, true)
	submitJudgement(t, token, userID, sessionID, true, false)

	// 4 itens legitimos: 1 falso alarme, 3 rejeicoes corretas.
	submitJudgement(t, token, userID, sessionID, false, true)
	submitJudgement(t, token, userID, sessionID, false, false)
	submitJudgement(t, token, userID, sessionID, false, false)
	submitJudgement(t, token, userID, sessionID, false, false)

	w := doJSON(t, http.MethodGet, "/api/v1/me/stats", token, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var stats struct {
		TotalAttempts  int     `json:"totalAttempts"`
		HitRate        float64 `json:"hitRate"`
		FalseAlarmRate float64 `json:"falseAlarmRate"`
		Accuracy       float64 `json:"accuracy"`
		DPrime         float64 `json:"dPrime"`
		Criterion      float64 `json:"criterion"`
	}
	decode(t, w, &stats)

	assert.Equal(t, 8, stats.TotalAttempts)
	assert.InDelta(t, 0.75, stats.HitRate, 0.0001)        // 3/4
	assert.InDelta(t, 0.25, stats.FalseAlarmRate, 0.0001) // 1/4
	assert.InDelta(t, 0.75, stats.Accuracy, 0.0001)       // (3+3)/8
	assert.InDelta(t, 1.0488, stats.DPrime, 0.001)
	assert.InDelta(t, 0.0, stats.Criterion, 0.001)
}

// TestIntegration_Stats_AcertoPerfeitoNaoGeraInfinito garante que o
// caso extremo (100% de acerto) devolve d' finito — sem a correcao de
// Hautus, z(1.0) seria +Inf e o JSON nem seria valido.
func TestIntegration_Stats_AcertoPerfeitoNaoGeraInfinito(t *testing.T) {
	resetDB(t)

	userID, token := registerAndLogin(t, "perfeito", "perfeito@example.com")
	consent(t, userID)

	sessionID := uuid.NewString()
	submitJudgement(t, token, userID, sessionID, true, true)
	submitJudgement(t, token, userID, sessionID, false, false)

	w := doJSON(t, http.MethodGet, "/api/v1/me/stats", token, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var stats struct {
		DPrime float64 `json:"dPrime"`
	}
	decode(t, w, &stats)

	assert.False(t, isInfOrNaN(stats.DPrime), "d' deve ser finito, recebeu %v", stats.DPrime)
	assert.Greater(t, stats.DPrime, 0.0)
}

// TestIntegration_Stats_BreakdownPorPista cobre o diagnostico por pista
// (byCue), que alimenta a selecao adaptativa da issue #28.
func TestIntegration_Stats_BreakdownPorPista(t *testing.T) {
	resetDB(t)

	userID, token := registerAndLogin(t, "porpista", "porpista@example.com")
	consent(t, userID)

	sessionID := uuid.NewString()

	// Erra duas vezes um item com typosquat.
	itemTyposquat := uuid.NewString()
	publishItem(t, itemTyposquat, userID, true, `{"subject":"banco-do-brasi1.com"}`)
	associateCue(t, itemTyposquat, cueTyposquat)
	postAttempt(t, token, userID, itemTyposquat, sessionID, false)

	// Acerta um item com urgency.
	itemUrgency := uuid.NewString()
	publishItem(t, itemUrgency, userID, true, `{"subject":"ultimas 2 horas"}`)
	associateCue(t, itemUrgency, cueUrgency)
	postAttempt(t, token, userID, itemUrgency, sessionID, true)

	w := doJSON(t, http.MethodGet, "/api/v1/me/stats", token, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var stats struct {
		ByCue []struct {
			CueCode  string  `json:"cueCode"`
			Answered int     `json:"answered"`
			Correct  int     `json:"correct"`
			Accuracy float64 `json:"accuracy"`
		} `json:"byCue"`
	}
	decode(t, w, &stats)

	accuracyPorPista := map[string]float64{}
	for _, cue := range stats.ByCue {
		accuracyPorPista[cue.CueCode] = cue.Accuracy
	}

	assert.Equal(t, 0.0, accuracyPorPista["typosquat"], "pista errada deve ter acuracia 0")
	assert.Equal(t, 1.0, accuracyPorPista["urgency"], "pista acertada deve ter acuracia 1")
}

// TestIntegration_Stats_SemTentativasNaoQuebra garante que um usuario
// novo recebe zeros em vez de erro/NaN.
func TestIntegration_Stats_SemTentativasNaoQuebra(t *testing.T) {
	resetDB(t)

	_, token := registerAndLogin(t, "zerado", "zerado@example.com")

	w := doJSON(t, http.MethodGet, "/api/v1/me/stats", token, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var stats struct {
		TotalAttempts int     `json:"totalAttempts"`
		DPrime        float64 `json:"dPrime"`
	}
	decode(t, w, &stats)

	assert.Zero(t, stats.TotalAttempts)
	assert.Zero(t, stats.DPrime)
}

// submitJudgement cria um item publicado com a malicia informada e
// registra uma tentativa com o veredito dado.
func submitJudgement(t *testing.T, token, userID, sessionID string, itemMalicioso, verdict bool) {
	t.Helper()

	itemID := uuid.NewString()
	publishItem(t, itemID, userID, itemMalicioso, `{"subject":"cenario de deteccao de sinal"}`)
	postAttempt(t, token, userID, itemID, sessionID, verdict)
}

func postAttempt(t *testing.T, token, userID, itemID, sessionID string, verdict bool) {
	t.Helper()

	action := "report"
	if !verdict {
		action = "ignore"
	}

	w := doJSON(t, http.MethodPost, "/api/v1/attempts", token, map[string]interface{}{
		"userId":    userID,
		"itemId":    itemID,
		"sessionId": sessionID,
		"action":    action,
		"verdict":   verdict,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("tentativa falhou (status %d): %s", w.Code, w.Body.String())
	}
}

func isInfOrNaN(value float64) bool {
	return value != value || value > 1e308 || value < -1e308
}
