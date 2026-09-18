//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	cueTyposquat = "00000000-0000-0000-0000-000000000002"
	cueUrgency   = "00000000-0000-0000-0000-000000000004"
)

// TestIntegration_Attempts_ExigeConsentimento e a regressao central da
// issue #21 contra banco real: nenhuma tentativa e coletada de quem nao
// consentiu em participar da pesquisa.
func TestIntegration_Attempts_ExigeConsentimento(t *testing.T) {
	resetDB(t)

	userID, token := registerAndLogin(t, "semconsent", "semconsent@example.com")
	itemID := uuid.NewString()
	publishItem(t, itemID, userID, true, `{"subject":"isca"}`)

	w := doJSON(t, http.MethodPost, "/api/v1/attempts", token, map[string]interface{}{
		"userId":    userID,
		"itemId":    itemID,
		"sessionId": uuid.NewString(),
		"action":    "report",
		"verdict":   true,
	})

	assert.Equal(t, http.StatusForbidden, w.Code)

	var count int64
	testDB.Raw("SELECT COUNT(*) FROM phishing_quest.attempts").Scan(&count)
	assert.Zero(t, count, "nenhuma tentativa deve ser gravada sem consentimento")
}

// TestIntegration_Attempts_IsCorrectCalculadoNoServidor garante que o
// veredito nao e confiado ao cliente: o servidor recalcula a partir de
// items.is_malicious.
func TestIntegration_Attempts_IsCorrectCalculadoNoServidor(t *testing.T) {
	resetDB(t)

	userID, token := registerAndLogin(t, "consentido", "consentido@example.com")
	consent(t, userID)

	itemMalicioso := uuid.NewString()
	publishItem(t, itemMalicioso, userID, true, `{"subject":"sua conta sera bloqueada"}`)

	sessionID := uuid.NewString()

	// verdict=true num item malicioso -> acerto.
	w := doJSON(t, http.MethodPost, "/api/v1/attempts", token, map[string]interface{}{
		"userId":    userID,
		"itemId":    itemMalicioso,
		"sessionId": sessionID,
		"action":    "report",
		"verdict":   true,
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	var created struct {
		IsCorrect *bool `json:"isCorrect"`
	}
	decode(t, w, &created)
	assert.NotNil(t, created.IsCorrect)
	assert.True(t, *created.IsCorrect)

	// verdict=false num item legitimo -> tambem acerto (rejeicao correta).
	itemLegitimo := uuid.NewString()
	publishItem(t, itemLegitimo, userID, false, `{"subject":"boleto legitimo"}`)

	w = doJSON(t, http.MethodPost, "/api/v1/attempts", token, map[string]interface{}{
		"userId":    userID,
		"itemId":    itemLegitimo,
		"sessionId": sessionID,
		"action":    "ignore",
		"verdict":   false,
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	decode(t, w, &created)
	assert.NotNil(t, created.IsCorrect)
	assert.True(t, *created.IsCorrect)
}

// TestIntegration_Attempts_AtualizaFilaLeitner cobre a integracao entre
// tentativas e revisao espacada (issue #27) de ponta a ponta: errar uma
// pista deve deixa-la devida para revisao imediata.
func TestIntegration_Attempts_AtualizaFilaLeitner(t *testing.T) {
	resetDB(t)

	userID, token := registerAndLogin(t, "leitner", "leitner@example.com")
	consent(t, userID)

	itemID := uuid.NewString()
	publishItem(t, itemID, userID, true, `{"subject":"clique urgente"}`)
	associateCue(t, itemID, cueUrgency)

	// Acerto: a pista avanca de caixa e sai da fila de vencidos.
	w := doJSON(t, http.MethodPost, "/api/v1/attempts", token, map[string]interface{}{
		"userId":    userID,
		"itemId":    itemID,
		"sessionId": uuid.NewString(),
		"action":    "report",
		"verdict":   true,
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	var box int
	testDB.Raw("SELECT box FROM phishing_quest.review_schedule WHERE user_id = ? AND cue_id = ?", userID, cueUrgency).Scan(&box)
	assert.Equal(t, 2, box, "acerto deve avancar a caixa do Leitner")

	w = doJSON(t, http.MethodGet, "/api/v1/review/due", token, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	var dueResponse struct {
		Due []struct {
			Box int `json:"box"`
		} `json:"due"`
	}
	decode(t, w, &dueResponse)
	assert.Empty(t, dueResponse.Due, "nada deve estar vencido logo apos um acerto")

	// Erro no mesmo item: volta para a caixa 1 e fica devido agora.
	w = doJSON(t, http.MethodPost, "/api/v1/attempts", token, map[string]interface{}{
		"userId":    userID,
		"itemId":    itemID,
		"sessionId": uuid.NewString(),
		"action":    "ignore",
		"verdict":   false,
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	testDB.Raw("SELECT box FROM phishing_quest.review_schedule WHERE user_id = ? AND cue_id = ?", userID, cueUrgency).Scan(&box)
	assert.Equal(t, 1, box, "erro deve voltar a pista para a caixa 1")

	w = doJSON(t, http.MethodGet, "/api/v1/review/due", token, nil)
	decode(t, w, &dueResponse)
	assert.Len(t, dueResponse.Due, 1, "pista errada deve reaparecer imediatamente")
}

// TestIntegration_Attempts_NaoRepeteItemNaSessao cobre GET /items/next
// (issue #22) com banco real.
func TestIntegration_Attempts_NaoRepeteItemNaSessao(t *testing.T) {
	resetDB(t)

	userID, token := registerAndLogin(t, "selecao", "selecao@example.com")
	consent(t, userID)

	primeiro := uuid.NewString()
	segundo := uuid.NewString()
	publishItem(t, primeiro, userID, true, `{"subject":"a"}`)
	publishItem(t, segundo, userID, false, `{"subject":"b"}`)

	sessionID := uuid.NewString()
	vistos := map[string]bool{}

	for i := 0; i < 2; i++ {
		w := doJSON(t, http.MethodGet, "/api/v1/items/next?sessionId="+sessionID, token, nil)
		assert.Equal(t, http.StatusOK, w.Code)

		var item struct {
			Id string `json:"id"`
		}
		decode(t, w, &item)
		assert.False(t, vistos[item.Id], "item repetido na mesma sessao: %s", item.Id)
		vistos[item.Id] = true

		// Responde para o item contar como visto na sessao.
		w = doJSON(t, http.MethodPost, "/api/v1/attempts", token, map[string]interface{}{
			"userId":    userID,
			"itemId":    item.Id,
			"sessionId": sessionID,
			"action":    "report",
			"verdict":   true,
		})
		assert.Equal(t, http.StatusCreated, w.Code)
	}

	// Esgotados os itens publicados, responde 404 em vez de repetir.
	w := doJSON(t, http.MethodGet, "/api/v1/items/next?sessionId="+sessionID, token, nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// consent registra o consentimento do participante via endpoint real.
func consent(t *testing.T, userID string) {
	t.Helper()

	w := doJSON(t, http.MethodPost, "/api/v1/auth/consent", "", map[string]interface{}{
		"userId":         userID,
		"consentVersion": "v1",
	})
	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Fatalf("consentimento falhou (status %d): %s", w.Code, w.Body.String())
	}
}
