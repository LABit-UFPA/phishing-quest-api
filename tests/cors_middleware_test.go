package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"phishing-quest/adapter/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestCORS_LiberaTudoSemVariavelDeAmbiente e a regressao da issue #18:
// antes desta mudanca a API nao tinha middleware de CORS, entao um
// front web (Flutter web, por exemplo) nunca conseguiria chamar a API
// diretamente do navegador. Sem CORS_ALLOWED_ORIGINS, o comportamento
// padrao deve liberar qualquer origem (adequado para dev local).
func TestCORS_LiberaTudoSemVariavelDeAmbiente(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.CORS())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "https://qualquer-origem-de-teste.example")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_RestringeOrigemViaVariavelDeAmbiente(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://phishingquest.labsc.com.br")

	r := gin.New()
	r.Use(middleware.CORS())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Origem permitida.
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", "https://phishingquest.labsc.com.br")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, "https://phishingquest.labsc.com.br", w.Header().Get("Access-Control-Allow-Origin"))

	// Origem nao permitida: o header nao deve ecoar essa origem.
	req2 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req2.Header.Set("Origin", "https://site-nao-autorizado.example")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.NotEqual(t, "https://site-nao-autorizado.example", w2.Header().Get("Access-Control-Allow-Origin"))
}
