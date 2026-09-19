//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestIntegration_ItemReview_PipelineCompleto cobre a issue #30 de ponta
// a ponta contra banco real: rascunho -> revisado -> publicado, e so
// entao o item se torna jogavel.
func TestIntegration_ItemReview_PipelineCompleto(t *testing.T) {
	resetDB(t)

	_, _ = registerAndLogin(t, "curador", "curador@example.com")
	promoteRole(t, "curador@example.com", "admin")
	adminToken := login(t, "curador@example.com")

	// 1. Cria rascunho tentando forcar status publicado e revisor falso.
	revisorFalso := uuid.NewString()
	w := doJSON(t, http.MethodPost, "/api/v1/admin/items", adminToken, map[string]interface{}{
		"channel":     "email",
		"isMalicious": true,
		"contentJson": map[string]string{"subject": "Sua conta sera bloqueada"},
		"status":      "published",
		"reviewedBy":  revisorFalso,
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	var draft struct {
		Id         string  `json:"id"`
		Status     string  `json:"status"`
		ReviewedBy *string `json:"reviewedBy"`
	}
	decode(t, w, &draft)
	assert.Equal(t, "draft", draft.Status, "cliente nao pode nascer com item publicado")
	assert.Nil(t, draft.ReviewedBy, "cliente nao pode se declarar revisor")

	// 2. Publicar sem revisao e recusado (criterio de aceite da issue).
	w = doJSON(t, http.MethodPost, "/api/v1/admin/items/"+draft.Id+"/publish", adminToken, nil)
	assert.Equal(t, http.StatusConflict, w.Code)

	// 3. Rascunho e invisivel nas rotas publicas.
	w = doJSON(t, http.MethodGet, "/api/v1/items/"+draft.Id, "", nil)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 4. Revisao humana: o revisor gravado e o do JWT.
	w = doJSON(t, http.MethodPost, "/api/v1/admin/items/"+draft.Id+"/review", adminToken, map[string]string{
		"reviewedBy": revisorFalso,
	})
	assert.Equal(t, http.StatusOK, w.Code)

	var reviewed struct {
		Status     string  `json:"status"`
		ReviewedBy *string `json:"reviewedBy"`
	}
	decode(t, w, &reviewed)
	assert.Equal(t, "reviewed", reviewed.Status)
	assert.NotNil(t, reviewed.ReviewedBy)
	assert.NotEqual(t, revisorFalso, *reviewed.ReviewedBy, "revisor deve vir do JWT, nao do corpo")

	// 5. Publicacao agora e aceita.
	w = doJSON(t, http.MethodPost, "/api/v1/admin/items/"+draft.Id+"/publish", adminToken, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var published struct {
		Status      string  `json:"status"`
		PublishedAt *string `json:"publishedAt"`
	}
	decode(t, w, &published)
	assert.Equal(t, "published", published.Status)
	assert.NotNil(t, published.PublishedAt)

	// 6. Agora o item aparece publicamente.
	w = doJSON(t, http.MethodGet, "/api/v1/items/"+draft.Id, "", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestIntegration_ItemReview_RascunhoNaoEntraNoJogo garante que o portao
// vale tambem para a selecao de itens: com um rascunho e nenhum item
// publicado, /items/next responde 404 em vez de servir o rascunho.
func TestIntegration_ItemReview_RascunhoNaoEntraNoJogo(t *testing.T) {
	resetDB(t)

	_, _ = registerAndLogin(t, "curador2", "curador2@example.com")
	promoteRole(t, "curador2@example.com", "admin")
	adminToken := login(t, "curador2@example.com")

	w := doJSON(t, http.MethodPost, "/api/v1/admin/items", adminToken, map[string]interface{}{
		"channel":     "sms",
		"isMalicious": true,
		"contentJson": map[string]string{"text": "rascunho nao revisado"},
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	participantID, participantToken := registerAndLogin(t, "jogador", "jogador@example.com")
	consent(t, participantID)

	w = doJSON(t, http.MethodGet, "/api/v1/items/next?sessionId="+uuid.NewString(), participantToken, nil)
	assert.Equal(t, http.StatusNotFound, w.Code, "rascunho nunca pode ser servido ao participante")
}

// TestIntegration_ItemReview_ParticipantNaoCria confirma o fail-closed
// de autorizacao no pipeline de conteudo.
func TestIntegration_ItemReview_ParticipantNaoCria(t *testing.T) {
	resetDB(t)

	_, participantToken := registerAndLogin(t, "intruso", "intruso@example.com")

	w := doJSON(t, http.MethodPost, "/api/v1/admin/items", participantToken, map[string]interface{}{
		"channel":     "email",
		"isMalicious": true,
		"contentJson": map[string]string{"subject": "conteudo injetado"},
	})

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// TestIntegration_ItemReview_RotaAnonimaAntigaNaoExiste e a regressao do
// bug de seguranca corrigido na #30: POST /api/v1/items era anonimo e
// escrevia direto na tabela servida ao jogo.
func TestIntegration_ItemReview_RotaAnonimaAntigaNaoExiste(t *testing.T) {
	resetDB(t)

	w := doJSON(t, http.MethodPost, "/api/v1/items", "", map[string]interface{}{
		"channel":     "email",
		"isMalicious": true,
		"contentJson": map[string]string{"subject": "injetado sem autenticacao"},
	})

	assert.Equal(t, http.StatusNotFound, w.Code)

	var count int64
	testDB.Raw("SELECT COUNT(*) FROM phishing_quest.items").Scan(&count)
	assert.Zero(t, count, "nenhum item pode ser criado anonimamente")
}

// TestIntegration_ItemReview_GeracaoEntraComoRascunho cobre a integracao
// opcional com LLM: o gerador nao encurta o pipeline.
func TestIntegration_ItemReview_GeracaoEntraComoRascunho(t *testing.T) {
	resetDB(t)

	_, _ = registerAndLogin(t, "curador3", "curador3@example.com")
	promoteRole(t, "curador3@example.com", "researcher")
	token := login(t, "curador3@example.com")

	w := doJSON(t, http.MethodPost, "/api/v1/admin/items/generate", token, map[string]interface{}{
		"channel":     "whatsapp",
		"isMalicious": true,
		"context":     "cobranca de fatura de energia",
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	var generated struct {
		Id     string `json:"id"`
		Status string `json:"status"`
		Source string `json:"source"`
	}
	decode(t, w, &generated)
	assert.Equal(t, "draft", generated.Status)
	assert.Equal(t, "llm_draft", generated.Source)

	// Aparece na fila de revisao.
	w = doJSON(t, http.MethodGet, "/api/v1/admin/items?status=draft", token, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var pendentes []struct {
		Id string `json:"id"`
	}
	decode(t, w, &pendentes)
	assert.Len(t, pendentes, 1)
	assert.Equal(t, generated.Id, pendentes[0].Id)
}

// TestIntegration_ItemReview_EstimadaECalibradaPersistemSeparadas cobre
// a issue #67: difficulty_estimated (a priori, do gerador) e
// difficulty_calibrated (a posteriori, medida a partir de attempts
// reais -- issue #66) sao colunas distintas e precisam sobreviver ao
// round-trip HTTP -> Postgres -> HTTP sem se confundirem.
func TestIntegration_ItemReview_EstimadaECalibradaPersistemSeparadas(t *testing.T) {
	resetDB(t)

	_, _ = registerAndLogin(t, "curador4", "curador4@example.com")
	promoteRole(t, "curador4@example.com", "admin")
	adminToken := login(t, "curador4@example.com")

	w := doJSON(t, http.MethodPost, "/api/v1/admin/items", adminToken, map[string]interface{}{
		"channel":                    "email",
		"isMalicious":                true,
		"contentJson":                map[string]string{"subject": "x"},
		"difficultyEstimated":        "hard",
		"difficultyCalibrated":       "medium",
		"phishScalePremiseAlignment": "high",
	})
	assert.Equal(t, http.StatusCreated, w.Code)

	var created struct {
		Id string `json:"id"`
	}
	decode(t, w, &created)

	w = doJSON(t, http.MethodGet, "/api/v1/admin/items/"+created.Id, adminToken, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	var fetched struct {
		DifficultyEstimated        *string `json:"difficultyEstimated"`
		DifficultyCalibrated       *string `json:"difficultyCalibrated"`
		PhishScalePremiseAlignment *string `json:"phishScalePremiseAlignment"`
	}
	decode(t, w, &fetched)

	if assert.NotNil(t, fetched.DifficultyEstimated) {
		assert.Equal(t, "hard", *fetched.DifficultyEstimated)
	}
	if assert.NotNil(t, fetched.DifficultyCalibrated) {
		assert.Equal(t, "medium", *fetched.DifficultyCalibrated)
	}
	assert.NotEqual(t, fetched.DifficultyEstimated, fetched.DifficultyCalibrated,
		"estimativa a priori e calibracao a posteriori nao podem se confundir na mesma coluna")
}

// TestIntegration_Migrations_ConstraintRejeitaDificuldadeForaDoVocabulario
// prova que o vocabulario fechado (easy|medium|hard e low|medium|high)
// e reforcado pelo banco, nao so por convencao em COMMENT (issue #67).
func TestIntegration_Migrations_ConstraintRejeitaDificuldadeForaDoVocabulario(t *testing.T) {
	resetDB(t)

	err := testDB.Exec(`
		INSERT INTO phishing_quest.items (id, channel, is_malicious, locale, content_json, difficulty_estimated)
		VALUES (?, 'email', TRUE, 'pt-BR', '{}'::jsonb, 'dificil')`, uuid.NewString()).Error

	assert.Error(t, err, "vocabulario em portugues (dificil) nao deveria passar no CHECK ingles (hard)")
	assert.Contains(t, err.Error(), "chk_items_difficulty_estimated")
}

// TestIntegration_Migrations_ConstraintImpedePublicarSemRevisor prova que
// a garantia nao depende so da aplicacao: o banco recusa a publicacao
// sem revisor mesmo por SQL direto.
func TestIntegration_Migrations_ConstraintImpedePublicarSemRevisor(t *testing.T) {
	resetDB(t)

	itemID := uuid.NewString()
	err := testDB.Exec(`
		INSERT INTO phishing_quest.items (id, channel, is_malicious, locale, content_json, status)
		VALUES (?, 'email', TRUE, 'pt-BR', '{}'::jsonb, 'draft')`, itemID).Error
	assert.NoError(t, err)

	err = testDB.Exec("UPDATE phishing_quest.items SET status = 'published' WHERE id = ?", itemID).Error

	assert.Error(t, err, "o banco deve recusar item publicado sem reviewed_by")
	assert.Contains(t, err.Error(), "chk_items_published_requires_reviewer")
}
