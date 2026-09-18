//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestIntegration_Assessments_SubmeteERecupera cobre os questionarios
// pre/pos/delayed da issue #25 contra banco real.
func TestIntegration_Assessments_SubmeteERecupera(t *testing.T) {
	resetDB(t)

	userID, token := registerAndLogin(t, "assessment", "assessment@example.com")
	consent(t, userID)

	w := doJSON(t, http.MethodPost, "/api/v1/assessments/pre", token, map[string]interface{}{
		"userId":            userID,
		"instrumentVersion": "v1",
		"responsesJson":     map[string]interface{}{"q1": "a", "q2": "b"},
	})
	assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, w.Code, "corpo: %s", w.Body.String())

	w = doJSON(t, http.MethodGet, "/api/v1/assessments/pre/users/"+userID, token, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var assessment struct {
		UserId string `json:"userId"`
		Phase  string `json:"phase"`
	}
	decode(t, w, &assessment)
	assert.Equal(t, userID, assessment.UserId)
	assert.Equal(t, "pre", assessment.Phase)
}

// TestIntegration_Assessments_FaseInvalidaRecusada garante que so as
// fases previstas no desenho do estudo sao aceitas.
func TestIntegration_Assessments_FaseInvalidaRecusada(t *testing.T) {
	resetDB(t)

	userID, token := registerAndLogin(t, "fasefake", "fasefake@example.com")
	consent(t, userID)

	w := doJSON(t, http.MethodPost, "/api/v1/assessments/depois_do_cafe", token, map[string]interface{}{
		"userId":            userID,
		"instrumentVersion": "v1",
	})

	assert.NotContains(t, []int{http.StatusOK, http.StatusCreated}, w.Code)
}

// TestIntegration_Export_ParticipantNaoAcessa e a regressao de
// autorizacao da issue #26: dado de pesquisa nao vaza para participante.
func TestIntegration_Export_ParticipantNaoAcessa(t *testing.T) {
	resetDB(t)

	_, token := registerAndLogin(t, "curioso", "curioso@example.com")

	w := doJSON(t, http.MethodGet, "/api/v1/research/export?format=csv", token, nil)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestIntegration_Export_SemTokenNaoAcessa(t *testing.T) {
	resetDB(t)

	w := doJSON(t, http.MethodGet, "/api/v1/research/export?format=csv", "", nil)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestIntegration_Export_ResearcherRecebeCSVPseudonimizado cobre o
// caminho autorizado E o requisito de privacidade: o CSV nao pode
// conter o userId real nem o e-mail do participante.
func TestIntegration_Export_ResearcherRecebeCSVPseudonimizado(t *testing.T) {
	resetDB(t)

	participantID, participantToken := registerAndLogin(t, "sujeito", "sujeito@example.com")
	consent(t, participantID)

	itemID := uuid.NewString()
	publishItem(t, itemID, participantID, true, `{"subject":"isca para export"}`)
	postAttempt(t, participantToken, participantID, itemID, uuid.NewString(), true)

	// Promove um segundo usuario a researcher.
	_, _ = registerAndLogin(t, "pesquisador", "pesquisador@example.com")
	promoteRole(t, "pesquisador@example.com", "researcher")
	researcherToken := login(t, "pesquisador@example.com") // token novo, com a role atualizada

	w := doJSON(t, http.MethodGet, "/api/v1/research/export?format=csv", researcherToken, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	csv := w.Body.String()
	assert.NotEmpty(t, csv)
	// O dado precisa estar la...
	assert.Contains(t, csv, itemID, "o export deve conter a tentativa coletada")
	// ...mas nunca identificando o participante.
	assert.NotContains(t, csv, participantID, "o export nao pode expor o userId real")
	assert.NotContains(t, csv, "sujeito@example.com", "o export nao pode expor o e-mail")
}

// TestIntegration_Export_PseudonimoEstavelEntreChamadas garante que o
// hash e deterministico (mesmo participante -> mesmo pseudo-id), senao
// nao daria para acompanhar o mesmo sujeito ao longo do estudo.
func TestIntegration_Export_PseudonimoEstavelEntreChamadas(t *testing.T) {
	resetDB(t)

	participantID, participantToken := registerAndLogin(t, "estavel", "estavel@example.com")
	consent(t, participantID)

	itemID := uuid.NewString()
	publishItem(t, itemID, participantID, true, `{"subject":"x"}`)
	postAttempt(t, participantToken, participantID, itemID, uuid.NewString(), true)

	_, _ = registerAndLogin(t, "pesquisador2", "pesquisador2@example.com")
	promoteRole(t, "pesquisador2@example.com", "researcher")
	researcherToken := login(t, "pesquisador2@example.com")

	primeiro := doJSON(t, http.MethodGet, "/api/v1/research/export?format=csv", researcherToken, nil)
	segundo := doJSON(t, http.MethodGet, "/api/v1/research/export?format=csv", researcherToken, nil)

	assert.Equal(t, http.StatusOK, primeiro.Code)
	assert.Equal(t, http.StatusOK, segundo.Code)
	assert.Equal(t, primeiro.Body.String(), segundo.Body.String(),
		"duas chamadas seguidas devem produzir o mesmo pseudo-id (bug corrigido na #26: mutacao in-place)")
}

// TestIntegration_Consent_Withdraw cobre o direito de exclusao (LGPD):
// depois de retirar o consentimento, novas tentativas sao recusadas.
func TestIntegration_Consent_Withdraw(t *testing.T) {
	resetDB(t)

	userID, token := registerAndLogin(t, "saiu", "saiu@example.com")
	consent(t, userID)

	itemID := uuid.NewString()
	publishItem(t, itemID, userID, true, `{"subject":"antes da saida"}`)
	postAttempt(t, token, userID, itemID, uuid.NewString(), true)

	w := doJSON(t, http.MethodPost, "/api/v1/auth/consent/"+userID+"/withdraw", token, nil)
	assert.Contains(t, []int{http.StatusNoContent, http.StatusOK}, w.Code)

	outroItem := uuid.NewString()
	publishItem(t, outroItem, userID, true, `{"subject":"depois da saida"}`)
	w = doJSON(t, http.MethodPost, "/api/v1/attempts", token, map[string]interface{}{
		"userId":    userID,
		"itemId":    outroItem,
		"sessionId": uuid.NewString(),
		"action":    "report",
		"verdict":   true,
	})
	assert.Equal(t, http.StatusForbidden, w.Code, "apos withdraw, novas tentativas devem ser recusadas")

	// A tentativa ja coletada NAO e apagada (dado de pesquisa ja consentido).
	var count int64
	testDB.Raw("SELECT COUNT(*) FROM phishing_quest.attempts WHERE user_id = ?", userID).Scan(&count)
	assert.Equal(t, int64(1), count)
}

// TestIntegration_Telemetry_IngestaoIdempotente cobre a issue #29: o id
// vem do cliente e reenviar o mesmo lote nao duplica evento.
func TestIntegration_Telemetry_IngestaoIdempotente(t *testing.T) {
	resetDB(t)

	userID, _ := registerAndLogin(t, "telemetria", "telemetria@example.com")

	eventID := uuid.NewString()
	batch := map[string]interface{}{
		"events": []map[string]interface{}{
			{
				"id":        eventID,
				"userId":    userID,
				"sessionId": uuid.NewString(),
				"eventType": "item_opened",
				"createdAt": "2026-09-17T12:00:00Z",
			},
		},
	}

	primeiro := doJSON(t, http.MethodPost, "/api/v1/telemetry-events", "", batch)
	assert.Contains(t, []int{http.StatusOK, http.StatusCreated, http.StatusAccepted}, primeiro.Code,
		"corpo: %s", primeiro.Body.String())

	segundo := doJSON(t, http.MethodPost, "/api/v1/telemetry-events", "", batch)
	assert.Contains(t, []int{http.StatusOK, http.StatusCreated, http.StatusAccepted}, segundo.Code)

	var count int64
	testDB.Raw("SELECT COUNT(*) FROM phishing_quest.telemetry_events WHERE id = ?", eventID).Scan(&count)
	assert.Equal(t, int64(1), count, "reenvio do mesmo evento nao deve duplicar")
}

// TestIntegration_CORS_RespondePreflight garante que o middleware de
// CORS da issue #18 esta efetivamente na cadeia do router real.
func TestIntegration_CORS_RespondePreflight(t *testing.T) {
	resetDB(t)

	req := newPreflightRequest("/api/v1/rankings")
	w := serve(req)

	assert.Contains(t, []int{http.StatusOK, http.StatusNoContent}, w.Code)
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

// TestIntegration_Rankings_PosicaoEUsername cobre a correcao da issue
// #24: o ranking traz posicao e username (nao so um score solto).
func TestIntegration_Rankings_PosicaoEUsername(t *testing.T) {
	resetDB(t)

	userID, _ := registerAndLogin(t, "rankeado", "rankeado@example.com")
	// user_scores e um LEDGER (uma linha por pontuacao ganha, somadas
	// depois) e a coluna de data se chama "timestamp", nao created_at.
	err := testDB.Exec(`INSERT INTO phishing_quest.user_scores (id, user_id, score, timestamp)
		VALUES (gen_random_uuid(), ?, 50, NOW())`, userID).Error
	if err != nil {
		t.Fatalf("falha ao inserir score: %v", err)
	}

	w := doJSON(t, http.MethodGet, "/api/v1/rankings", "", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Ranking []struct {
			Position   int    `json:"position"`
			UserId     string `json:"userId"`
			Username   string `json:"username"`
			TotalScore int    `json:"totalScore"`
		} `json:"ranking"`
	}
	decode(t, w, &response)

	if !assert.NotEmpty(t, response.Ranking, "ranking nao deveria vir vazio: %s", w.Body.String()) {
		return
	}
	assert.Equal(t, 1, response.Ranking[0].Position)
	assert.Equal(t, userID, response.Ranking[0].UserId)
	assert.Equal(t, "rankeado", response.Ranking[0].Username)
	assert.Equal(t, 50, response.Ranking[0].TotalScore)
}
