//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestIntegration_Auth_RegistroLoginEAcessoProtegido cobre o fluxo
// completo de autenticacao (issue #10) contra banco real: cadastro,
// login com emissao de token e uso do token numa rota protegida.
func TestIntegration_Auth_RegistroLoginEAcessoProtegido(t *testing.T) {
	resetDB(t)

	userID, token := registerAndLogin(t, "authuser", "authuser@example.com")
	assert.NotEmpty(t, userID)
	assert.NotEmpty(t, token)

	// Rota protegida COM token: passa pelo AuthRequired.
	w := doJSON(t, http.MethodGet, "/api/v1/me/stats", token, nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestIntegration_Auth_RotaProtegidaSemTokenRecusa garante que o
// middleware real esta na cadeia da rota (o teste unitario de handler
// nao cobre isso, porque monta o router sem middleware).
func TestIntegration_Auth_RotaProtegidaSemTokenRecusa(t *testing.T) {
	resetDB(t)

	w := doJSON(t, http.MethodGet, "/api/v1/me/stats", "", nil)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIntegration_Auth_TokenInvalidoRecusa(t *testing.T) {
	resetDB(t)

	w := doJSON(t, http.MethodGet, "/api/v1/me/stats", "token.completamente.invalido", nil)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestIntegration_Auth_NovoUsuarioNasceParticipant garante que nao ha
// auto-promocao de role no cadastro (issue #26): quem se registra nunca
// nasce researcher/admin.
func TestIntegration_Auth_NovoUsuarioNasceParticipant(t *testing.T) {
	resetDB(t)

	_, token := registerAndLogin(t, "roleuser", "roleuser@example.com")

	// A role viaja no token e e usada pelo RequireRole; a prova pratica
	// e que a rota de pesquisa recusa este usuario.
	w := doJSON(t, http.MethodGet, "/api/v1/research/export?format=csv", token, nil)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestIntegration_Auth_LoginComSenhaErradaRecusa(t *testing.T) {
	resetDB(t)

	registerAndLogin(t, "senhauser", "senhauser@example.com")

	w := doJSON(t, http.MethodPost, "/api/v1/users/login", "", map[string]string{
		"email":    "senhauser@example.com",
		"password": "senha-errada",
	})

	assert.NotEqual(t, http.StatusOK, w.Code)
}
