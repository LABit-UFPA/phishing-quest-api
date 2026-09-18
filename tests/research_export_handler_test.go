package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"phishing-quest/adapter/http/handler"
	"phishing-quest/core/usecase"
	"phishing-quest/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestResearchExportHandler_ExportCSV_Sucesso(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("RESEARCH_EXPORT_SALT", "salt-de-teste")

	mockRepo := new(MockResearchExportRepository)
	uc := usecase.NewResearchExportUseCase(mockRepo)
	h := handler.NewResearchExportHandler(uc)

	itemID := uuid.New()
	sessionID := uuid.New()
	verdict := true
	confidence := 4
	isCorrect := true
	latency := 2500

	mockRepo.On("ListAttemptsWithItemInfo").Return([]*domain.ResearchExportRow{
		{
			PseudoUserId: uuid.New().String(),
			ItemId:       itemID,
			Channel:      "whatsapp",
			IsMalicious:  true,
			SessionId:    sessionID,
			Condition:    "feedback_formativo",
			Verdict:      &verdict,
			Action:       domain.ActionReport,
			Confidence:   &confidence,
			IsCorrect:    &isCorrect,
			LatencyMs:    &latency,
			ClickedLink:  false,
		},
	}, nil)

	r := gin.New()
	r.GET("/api/v1/research/export", h.ExportCSV)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/research/export?format=csv", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))

	body := w.Body.String()
	lines := strings.Split(strings.TrimSpace(body), "\n")
	assert.Len(t, lines, 2) // cabecalho + 1 linha de dados
	assert.Contains(t, lines[0], "pseudo_user_id")
	assert.Contains(t, lines[1], "whatsapp")
	assert.Contains(t, lines[1], "feedback_formativo")
	// itemId e sessionId sao esperados no CSV (nao sao PII do usuario);
	// so o UserId real e substituido por hash.
	assert.Contains(t, body, itemID.String())
	assert.Contains(t, body, sessionID.String())
}

func TestResearchExportHandler_ExportCSV_FormatoInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockResearchExportRepository)
	uc := usecase.NewResearchExportUseCase(mockRepo)
	h := handler.NewResearchExportHandler(uc)

	r := gin.New()
	r.GET("/api/v1/research/export", h.ExportCSV)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/research/export?format=xml", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockRepo.AssertNotCalled(t, "ListAttemptsWithItemInfo")
}
