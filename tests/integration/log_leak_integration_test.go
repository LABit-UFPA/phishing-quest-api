//go:build integration

package integration

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestIntegration_Log_CadastroNaoVazaSenha e a regressao da issue #60.
//
// O vazamento acontecia no caminho real: o repositorio generico logava a
// entidade com "%+v" e domain.User carregava a senha em texto puro, entao
// todo cadastro gravava a senha do usuario no log em nivel INFO. Os
// testes unitarios nunca pegariam isso porque usam mock de repositorio,
// que nao loga nada — o unico jeito de provar a correcao e exercitar o
// repositorio de verdade e inspecionar a saida de log.
func TestIntegration_Log_CadastroNaoVazaSenha(t *testing.T) {
	resetDB(t)

	const senha = "SenhaSuperSecreta#12345"

	logCapturado := capturarLog(t, func() {
		w := doJSON(t, http.MethodPost, "/api/v1/users/register", "", map[string]string{
			"username": "semvazamento",
			"email":    "semvazamento@example.com",
			"password": senha,
		})
		assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, w.Code,
			"cadastro deveria ter sido aceito: %s", w.Body.String())
	})

	assert.NotEmpty(t, logCapturado, "o cadastro deveria ter produzido log (senao o teste nao prova nada)")
	assert.NotContains(t, logCapturado, senha,
		"a senha em texto puro apareceu no log do cadastro")
	// "Password:" era o prefixo exato do dump de struct que vazava.
	assert.NotContains(t, logCapturado, "Password:",
		"o log voltou a despejar a struct do usuario")
}

// TestIntegration_Log_LoginNaoVazaSenha cobre o outro caminho onde a
// senha em texto puro trafega (comparacao com o hash).
func TestIntegration_Log_LoginNaoVazaSenha(t *testing.T) {
	resetDB(t)

	const senha = "OutraSenhaSecreta#67890"

	w := doJSON(t, http.MethodPost, "/api/v1/users/register", "", map[string]string{
		"username": "loginsemvazamento",
		"email":    "loginsemvazamento@example.com",
		"password": senha,
	})
	assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, w.Code)

	logCapturado := capturarLog(t, func() {
		w := doJSON(t, http.MethodPost, "/api/v1/users/login", "", map[string]string{
			"email":    "loginsemvazamento@example.com",
			"password": senha,
		})
		assert.Equal(t, http.StatusOK, w.Code)
	})

	assert.NotContains(t, logCapturado, senha, "a senha apareceu no log do login")
}

// TestIntegration_Log_MantemRastreabilidade garante que a correcao nao
// cegou o log: ele precisa continuar dizendo QUAL entidade foi criada.
// Sem isso, a "correcao" seria so remover observabilidade.
func TestIntegration_Log_MantemRastreabilidade(t *testing.T) {
	resetDB(t)

	logCapturado := capturarLog(t, func() {
		w := doJSON(t, http.MethodPost, "/api/v1/users/register", "", map[string]string{
			"username": "rastreavel",
			"email":    "rastreavel@example.com",
			"password": "SenhaQualquer#123",
		})
		assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, w.Code)
	})

	assert.Contains(t, logCapturado, "User", "o log deve identificar o tipo da entidade")
	assert.Contains(t, logCapturado, "id=", "o log deve identificar o id da entidade")
}

// capturarLog redireciona a saida do logrus durante a execucao de fn e
// devolve o que foi escrito, restaurando o destino original ao final.
func capturarLog(t *testing.T, fn func()) string {
	t.Helper()

	var buffer bytes.Buffer
	saidaOriginal := logrus.StandardLogger().Out
	nivelOriginal := logrus.GetLevel()

	logrus.SetOutput(&buffer)
	logrus.SetLevel(logrus.DebugLevel) // captura tudo, para o teste nao passar por filtro de nivel
	defer func() {
		logrus.SetOutput(saidaOriginal)
		logrus.SetLevel(nivelOriginal)
	}()

	fn()

	return strings.TrimSpace(buffer.String())
}
