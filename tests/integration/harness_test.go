//go:build integration

// Package integration reune os testes de integracao dos endpoints de
// pesquisa (issue #31). Diferente do pacote tests/ (unitario, com
// mocks), aqui sobe o app REAL: container de dependencias, router com
// todos os middlewares e um Postgres de verdade ja migrado.
//
// Ficam atras da build tag "integration" para nao entrarem no
// `go test ./...` do dia a dia, que precisa rodar sem banco. No CI sao
// executados explicitamente com -tags=integration, depois do flyway
// aplicar as migrations — se o banco nao estiver acessivel, a suite
// FALHA em vez de pular, senao o CI ficaria verde sem ter testado nada.
package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	apphttp "phishing-quest/adapter/http"
	"phishing-quest/container"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	testRouter *gin.Engine
	testDB     *gorm.DB
)

// volatileTables sao truncadas entre testes. cues NAO entra na lista:
// as 10 pistas da taxonomia sao dados de referencia inseridos pela
// migration V20260917110000, e apagá-las quebraria os testes de pista.
var volatileTables = []string{
	"phishing_quest.review_schedule",
	"phishing_quest.telemetry_events",
	"phishing_quest.attempts",
	"phishing_quest.assessments",
	"phishing_quest.study_participants",
	"phishing_quest.item_cues",
	"phishing_quest.items",
	"phishing_quest.user_answers",
	"phishing_quest.user_scores",
	"phishing_quest.answers",
	"phishing_quest.questions",
	"phishing_quest.categories",
	"phishing_quest.users",
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	// Defaults iguais ao docker-compose, para rodar localmente sem
	// exportar nada: docker-compose up postgres flyway && go test -tags=integration ./tests/integration/
	setEnvDefault("DB_HOST", "localhost")
	setEnvDefault("DB_USER", "labsc")
	setEnvDefault("DB_PASSWORD", "phishingquest")
	setEnvDefault("DB_NAME", "phishing_quest")
	setEnvDefault("DB_PORT", "5432")
	setEnvDefault("DB_SSLMODE", "disable")
	setEnvDefault("JWT_SECRET", "segredo-de-teste-de-integracao")
	setEnvDefault("RESEARCH_EXPORT_SALT", "salt-de-teste-de-integracao")

	db, err := openTestDB()
	if err != nil {
		log.Fatalf("integracao: nao foi possivel conectar ao Postgres de teste: %v\n"+
			"Suba o banco e aplique as migrations antes de rodar:\n"+
			"  docker-compose up -d phishing-quest-postgresql phishing-quest-flyway-phishing-quest", err)
	}
	testDB = db

	if err := assertSchemaMigrated(db); err != nil {
		log.Fatalf("integracao: schema nao esta migrado: %v", err)
	}

	// container.NewContainer abre a propria conexao lendo as mesmas
	// variaveis de ambiente, e monta o grafo completo de dependencias.
	// Usar o router real (com CORS, AuthRequired e RequireRole de
	// verdade) e o ponto do teste de integracao: nenhum middleware fica
	// de fora, ao contrario dos testes de handler com httptest isolado.
	testRouter = apphttp.SetupRouter(container.NewContainer())

	os.Exit(m.Run())
}

func setEnvDefault(key, value string) {
	if _, exists := os.LookupEnv(key); !exists {
		_ = os.Setenv(key, value)
	}
}

func openTestDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC connect_timeout=10",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"), os.Getenv("DB_SSLMODE"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// assertSchemaMigrated falha rapido e com mensagem clara se as
// migrations nao foram aplicadas, em vez de deixar cada teste explodir
// com "relation does not exist".
func assertSchemaMigrated(db *gorm.DB) error {
	required := []string{"users", "items", "cues", "attempts", "study_participants", "assessments", "review_schedule"}
	for _, table := range required {
		var exists bool
		err := db.Raw(`SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'phishing_quest' AND table_name = ?
		)`, table).Scan(&exists).Error
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("tabela phishing_quest.%s nao existe (rode o flyway)", table)
		}
	}
	return nil
}

// resetDB limpa as tabelas volateis para cada teste comecar de um
// estado conhecido. Chamado no inicio de cada teste (nao no fim) para
// que uma falha deixe o estado disponivel para inspecao.
func resetDB(t *testing.T) {
	t.Helper()

	for _, table := range volatileTables {
		if err := testDB.Exec("TRUNCATE TABLE " + table + " CASCADE").Error; err != nil {
			t.Fatalf("falha ao truncar %s: %v", table, err)
		}
	}
}

// doJSON executa uma requisicao contra o router real e devolve o
// recorder. token vazio = requisicao sem Authorization.
func doJSON(t *testing.T, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("falha ao serializar corpo: %v", err)
		}
		reader = bytes.NewReader(encoded)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

// newPreflightRequest monta um OPTIONS de preflight CORS.
func newPreflightRequest(path string) *http.Request {
	req := httptest.NewRequest(http.MethodOptions, path, nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	return req
}

// serve executa uma requisicao ja montada contra o router real.
func serve(req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	return w
}

// decode desserializa o corpo da resposta, falhando o teste com o
// conteudo bruto quando o JSON nao bate com o destino esperado.
func decode(t *testing.T, w *httptest.ResponseRecorder, target interface{}) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), target); err != nil {
		t.Fatalf("resposta nao e o JSON esperado (status %d): %s", w.Code, w.Body.String())
	}
}

// registerAndLogin cria um usuario e devolve (id, token). Centraliza o
// fluxo usado por praticamente todos os testes de integracao.
func registerAndLogin(t *testing.T, username, email string) (string, string) {
	t.Helper()

	w := doJSON(t, http.MethodPost, "/api/v1/users/register", "", map[string]string{
		"username": username,
		"name":     username,
		"email":    email,
		"password": "senha1234",
	})
	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		t.Fatalf("registro falhou (status %d): %s", w.Code, w.Body.String())
	}

	var registered struct {
		Id string `json:"id"`
	}
	decode(t, w, &registered)

	return registered.Id, login(t, email)
}

func login(t *testing.T, email string) string {
	t.Helper()

	w := doJSON(t, http.MethodPost, "/api/v1/users/login", "", map[string]string{
		"email":    email,
		"password": "senha1234",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("login falhou (status %d): %s", w.Code, w.Body.String())
	}

	var session struct {
		Token string `json:"token"`
	}
	decode(t, w, &session)
	if session.Token == "" {
		t.Fatal("login nao devolveu token")
	}
	return session.Token
}

// promoteRole eleva a role do usuario direto no banco. Nao existe (de
// propósito) endpoint de auto-promocao, entao o teste faz o papel do
// operador que atribui a role manualmente.
func promoteRole(t *testing.T, email, role string) {
	t.Helper()
	if err := testDB.Exec("UPDATE phishing_quest.users SET role = ? WHERE email = ?", role, email).Error; err != nil {
		t.Fatalf("falha ao promover %s para %s: %v", email, role, err)
	}
}

// publishItem cria um item ja publicado direto no banco, com revisor
// registrado para respeitar a constraint da issue #30. Usado pelos
// testes que precisam de itens jogaveis sem exercitar todo o pipeline
// de curadoria.
func publishItem(t *testing.T, itemID, reviewerID string, isMalicious bool, content string) {
	t.Helper()

	err := testDB.Exec(`
		INSERT INTO phishing_quest.items
			(id, channel, is_malicious, locale, content_json, explanation, source, status, reviewed_by, reviewed_at, published_at)
		VALUES (?, 'email', ?, 'pt-BR', ?::jsonb, '', 'teste', 'published', ?, NOW(), NOW())`,
		itemID, isMalicious, content, reviewerID).Error
	if err != nil {
		t.Fatalf("falha ao inserir item publicado: %v", err)
	}
}

// associateCue liga um item a uma pista da taxonomia semeada pela
// migration (ids fixos 00000000-...-0000000000NN).
func associateCue(t *testing.T, itemID, cueID string) {
	t.Helper()

	err := testDB.Exec(`
		INSERT INTO phishing_quest.item_cues (id, item_id, cue_id)
		VALUES (gen_random_uuid(), ?, ?)`, itemID, cueID).Error
	if err != nil {
		t.Fatalf("falha ao associar pista: %v", err)
	}
}
