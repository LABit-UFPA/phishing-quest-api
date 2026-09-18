# Plano de execução — Backend 100% funcional

> Ordem de trabalho para fechar as 24 issues abertas do `phishing-quest-api`.
> Escrito sobre os padrões reais do código: DI manual em `container/container.go`, CRUD genérico `IRepository[T]` por embedding, migrations Flyway versionadas, erro em `gin.H{"error": ...}`.
> Docs irmãos: `BACKEND_TODO.md` (diagnóstico), `../ROADMAP_PESQUISA_2027.md` (visão macro), `../INTEGRACAO_FRONT_BACK.md` (contrato).

---

## Sprint 0 — Ambiente (meio dia, pré-requisito de tudo)

O ambiente não roda hoje: **Go não está instalado** na máquina e os defaults de conexão do `postgres/connector.go` (`default_user`/`default_password`) não batem com o `docker-compose.yml` (`labsc`/`phishingquest`).

```bash
brew install go                    # Go 1.22+ (go.mod pede 1.22.5)
docker --version                   # necessário para postgres + flyway
```

- [ ] Instalar Go e Docker
- [ ] Criar `.env` na raiz do backend:
  ```
  DB_HOST=localhost
  DB_USER=labsc
  DB_PASSWORD=phishingquest
  DB_NAME=phishing_quest
  DB_PORT=5432
  DB_SSLMODE=disable
  JWT_SECRET=dev-only-trocar-em-producao
  JWT_EXPIRES_IN=24h
  ```
- [ ] Subir infra: `docker-compose up -d phishing-quest-postgresql phishing-quest-flyway-phishing-quest`
- [ ] Confirmar migrations aplicadas: `docker exec -it <pg> psql -U labsc -d phishing_quest -c '\dt phishing_quest.*'`
- [ ] `go build ./...` (vai falhar nos testes — é a issue #32)
- [ ] Rodar API: `go run .` → `curl localhost:8080/api/v1/categories`

**DoD do sprint:** `curl` retorna resposta da API rodando local contra o Postgres do Docker.

---

## Sprint 1 — Desbloqueio de autenticação (issues #10, #11, #12, #32)

É o sprint que destrava o front. Fazer nesta ordem porque #10 define o contrato que o front consome.

### #32 — Consertar a suíte de testes primeiro
Sem isso não há rede de segurança para o resto. `tests/user_usecase_test.go` tem mock com assinaturas antigas (`Create(*User) error`, `GetByID(int)`); a interface atual embute `IRepository[domain.User]`.
- [ ] Reescrever `MockUserRepository` implementando `Create(*T) (*T, error)`, `Update`, `Delete`, `GetByID(uuid.UUID)`, `GetAll`, `GetByEmail`
- [ ] `go test ./...` verde

### #12 — Senha em texto puro (5 min, maior risco/esforço do sprint)
- [ ] `domain/user.go`: campo `Password` → tag `json:"-"` (hoje volta na resposta de cadastro)
- [ ] Criar `dto.UserResponseDTO{id, username, email, totalScore}` e usar em `CreateUser`
- [ ] Verificar: `curl -X POST .../users/register` não retorna senha nem hash

### #11 — Path de registro
- [ ] Decidir canônico: **manter `POST /api/v1/users/register`** e alinhar o front (issue #7 do repo do app)
- [ ] Registrar a decisão em `../INTEGRACAO_FRONT_BACK.md`

### #10 — JWT
Arquivos novos/alterados:
```
core/service/jwt_service.go        (novo)  Generate(userID, role) / Parse(token)
adapter/http/middleware/auth.go    (novo)  AuthRequired() gin.HandlerFunc
dto/user.go                        (alt)   UserLoginResponseDTO + campo Token
core/usecase/user_usecase.go       (alt)   Login() retorna token
adapter/http/router/*.go           (alt)   aplicar middleware nos grupos protegidos
container/container.go             (alt)   registrar JWTService
```
- [ ] `go get github.com/golang-jwt/jwt/v5`
- [ ] `JWT_SECRET` e `JWT_EXPIRES_IN` via env (padrão do `getEnv` em `postgres/connector.go`)
- [ ] Middleware valida `Authorization: Bearer <token>`, injeta `userId` no `gin.Context` (`c.Set("userId", ...)`)
- [ ] Aplicar em `game`, `user-answers` (e nos endpoints novos dos sprints 3–5)
- [ ] Manter `login`/`register`/`categories` públicos

**DoD do sprint:** login retorna token; rota protegida sem token responde 401; front consegue autenticar de verdade.

---

## Sprint 2 — Correção dos endpoints existentes (issues #13, #14, #15, #16, #17, #18, #33)

### #13 — `/categories/:id/questions` (dois bugs, não um)
- [ ] `category_handler.go:48`: `c.Param("category_id")` → `c.Param("id")`
- [ ] `NewCategoryUseCase(categoryRepo, questionRepo)` — hoje `questionRepo` é declarado e nunca injetado (nil panic garantido)
- [ ] `container.go`: mover criação de `categoryUseCase` para depois de `questionRepo` e passar os dois
- [ ] Verificar: endpoint retorna `CategoryQuestionsDTO` com as questões

### #14 — Score no `/game/answer`
- [ ] `user_score_repository.go`: `IncrementScore` faz UPDATE e retorna `ErrRecordNotFound` se `RowsAffected == 0`. Trocar por **upsert** (criar linha com `uuid.New()` se não existir)
- [ ] `game_usecase.go`: gravar `user_answers` no fluxo (hoje o histórico não é persistido)
- [ ] Retornar `totalScore` atualizado no `AnswerResultDTO`
- [ ] Verificar: submeter resposta de usuário novo não dá 500

### #17 — Validações de domínio
- [ ] `domain/answer.go`: remover `validate:"required"` de `IsCorrect bool` (validator trata `false` como zero e falha)
- [ ] `domain/question.go`: resolver `CorrectAnswer` — a coluna foi dropada na migration `V20250104153937`. Remover campo e validação
- [ ] Verificar: criar answer com `isCorrect:false` funciona

### #16 — `GET /users/:id` e stub
- [ ] Implementar busca real via `userRepo.GetByID` retornando DTO seguro
- [ ] Remover `GetTeste` (`GET /users` retorna a string `"olhaaaaaa"`)

### #15 — Convenção JSON
- [ ] Decidir: **camelCase** (é a maioria) e migrar `dto/answer.go` (`SubmitAnswerDTO`, `AnswerResultDTO` usam snake_case)
- [ ] Atualizar `../INTEGRACAO_FRONT_BACK.md` e avisar o front (issue #10 do app)

### #18 / #33 — Infra
- [ ] Middleware CORS em `adapter/http/router.go` (`SetupRouter`)
- [ ] Helper de erro padronizado (`adapter/http/response/`) mantendo `gin.H{"error": ...}`
- [ ] `.env.example` versionado
- [ ] Remover `config.yaml` (vestigial: diz porta 8088 e schema `pessoas`, não é lido por ninguém)
- [ ] `flyway-prod.config`: corrigir `flyway.locations` para `migrate/changelogs`

**DoD do sprint:** todos os endpoints existentes respondem corretamente; nenhum 500/400 espúrio. Fase 1 fechada, front pode remover todos os mocks.

---

## Sprint 3 — Fundação de pesquisa: schema (issues #19, #20, #21, #29)

A partir daqui é feature nova. Cada entidade segue o mesmo molde de 6 arquivos + 2 edições (ver "Checklist de feature nova" ao final).

Uma migration por issue, timestamps crescentes acima de `V20250104153937`:

| Issue | Migration | Tabelas |
|-------|-----------|---------|
| #19 | `V2026____add_cues_tables.sql` | `cues`, `item_cues` |
| #20 | `V2026____add_items_table.sql` | `items` (multicanal) |
| #21 | `V2026____add_attempts_table.sql` | `attempts` |
| #29 | `V2026____add_telemetry_events.sql` | `telemetry_events` |

- [ ] **#19** `cues` + seed das 10 pistas (`sender_domain_mismatch`, `typosquat`, `homoglyph`, `urgency`, `authority`, `generic_greeting`, `credential_request`, `link_text_mismatch`, `unexpected_attachment`, `scarcity`); `item_cues` com `span_start`/`span_end`
- [ ] **#20** `items` com `channel` (email|sms|whatsapp|website|phone_call|pix_qr), `is_malicious`, `content_json JSONB`, campos do Phish Scale; plano de migração dos dados hoje mockados no app
- [ ] **#21** `attempts` + `POST /attempts` — **é o dado central do artigo**: `verdict`, `action`, `confidence`, `justification`, `latency_ms`, `clicked_link`, `condition`, `session_id`
- [ ] **#29** `telemetry_events` + ingestão em lote idempotente (suporta a fila offline do app)

Padrões obrigatórios: `CREATE TABLE IF NOT EXISTS phishing_quest.<nome>`, PK `UUID PRIMARY KEY` (UUID gerado no usecase com `uuid.New()`, não no banco), FK `REFERENCES phishing_quest.x(id)`, colunas snake_case, `TIMESTAMPTZ NOT NULL DEFAULT NOW()`, `TableName()` retornando `phishing_quest.<tabela>`.

**DoD:** `docker-compose up` aplica as migrations; `POST /attempts` persiste tentativa completa e recusa usuário sem consentimento.

---

## Sprint 4 — Estudo: consentimento, papéis e métricas (issues #22, #23, #25, #26)

- [ ] **#25** `study_participants` + `assessments`; `POST /auth/consent` (TCLE versionado + atribuição de condição); `POST /assessments/:phase` (pre|post|delayed_4w)
- [ ] **#26** coluna `role` em `users` (participant|researcher|admin) + middleware de autorização (reutiliza o de #10); `GET /research/export?format=csv` pseudonimizado
- [ ] **#22** `GET /items/next` — seleção balanceada malicioso/legítimo, sem repetir na mesma fase, com gancho para adaptativo
- [ ] **#23** `GET /me/stats` — **d′ = z(hit) − z(falso alarme)**, critério c, taxas, breakdown por pista. Cálculo server-side (fonte única para app e análise), com correção log-linear para taxas 0/1

**DoD:** um participante consegue percorrer consentimento → itens → tentativas → stats com d′ calculado corretamente.

---

## Sprint 5 — Mecânicas e fechamento (issues #24, #27, #28, #30, #31)

- [ ] **#24** `GET /rankings`: DTO próprio (`userId`, `username`, `totalScore`, `position`), JOIN com `users`, escopo por coorte. Hoje a query faz `SUM(score) AS total_score` mas mapeia em `UserScore` cujo campo é `score` → retorna 0
- [ ] **#27** `review_schedule` + Leitner por pista + `GET /review/due`
- [ ] **#28** seleção adaptativa por domínio de pista (integra em `/items/next`)
- [ ] **#30** `POST /admin/items` + fluxo rascunho → revisado → publicado, com `reviewed_by` obrigatório
- [ ] **#31** testes de integração (auth, attempts, stats/d′, assessments, export) + GitHub Actions com Postgres efêmero

**DoD:** CI verde obrigatório para merge; backend pronto para o piloto.

---

## Dependências (o que não pode inverter)

```
Sprint 0 (ambiente)
   └─> #32 (testes compilam)
         └─> #10 (JWT) ──> #26 (roles) ──> #23 (/me/stats)  [precisa de userId no contexto]
   └─> #12, #11 (contrato de auth do front)
   └─> #13, #14, #17 (bugs independentes, podem ir em paralelo)
        #15 (convenção JSON) antes do front alinhar serialização
   └─> #19, #20 ──> #21 (attempts referencia items e cues)
                      └─> #23 (d′ precisa de attempts)
                      └─> #27, #28 (agendamento precisa de pistas e histórico)
   └─> #25 ──> #21 (attempts valida consentimento)
   └─> #31 (CI por último, cobre o que já existe)
```

Regra prática: **#10 e #21 são os dois gargalos**. Tudo de auth depende do primeiro; tudo de pesquisa depende do segundo.

---

## Estimativa

| Sprint | Issues | Esforço |
|--------|--------|---------|
| 0 — Ambiente | — | 0,5 dia |
| 1 — Auth | #32, #12, #11, #10 | 3–4 dias |
| 2 — Correções | #13, #14, #15, #16, #17, #18, #33 | 4–5 dias |
| 3 — Schema de pesquisa | #19, #20, #21, #29 | 5–7 dias |
| 4 — Estudo | #22, #23, #25, #26 | 6–8 dias |
| 5 — Mecânicas + CI | #24, #27, #28, #30, #31 | 6–8 dias |

Total: **~5 a 6 semanas** de trabalho focado — casa com a Fase 1 (out/2026) e Fase 2 (nov/2026) do roadmap.

---

## Checklist de feature nova (molde do projeto)

Para cada entidade `Foo`, replicar 6 arquivos + 2 edições + 1 migration:

1. `domain/foo.go` — struct com tags `json`/`gorm`/`validate`, `TableName() "phishing_quest.foos"`, `Validate()`, opcional `ToDTO()`
2. `adapter/repository/foo_repository.go` — `IFooRepository` embutindo `IRepository[domain.Foo]` + métodos custom; struct embutindo `IRepository[domain.Foo]` + `db *gorm.DB`; construtor retorna a **interface**
3. `core/usecase/foo_usecase.go` — recebe interfaces de repo; gera `uuid.New()`, chama `Validate()`, delega
4. `adapter/http/handler/foo_handler.go` — `func (h *FooHandler) M(c *gin.Context)`, erro em `gin.H{"error": ...}`
5. `adapter/http/router/foo_router.go` — `SetupFooRoutes(r, handler)` com `r.Group("api/v1/foos")`
6. `dto/foo.go` — se houver request/response especializado
7. Editar `container/container.go`: 3 campos no struct + 3 instanciações + 3 no literal de retorno
8. Editar `adapter/http/router.go`: `router.SetupFooRoutes(r, cont.FooHandler)`
9. Migration `migrate/changelogs/V<timestamp>__add_foo_table.sql`

**Nunca** usar `AutoMigrate` do GORM — o schema é 100% Flyway. **Nunca** `validate:"required"` em `bool`.

---

## Fluxo de trabalho por issue

```bash
git checkout -b fix/10-jwt-auth
# implementar
go build ./... && go test ./...
git commit -m "feat(auth): emitir JWT no login e validar via middleware

Closes #10"
git push -u origin fix/10-jwt-auth
gh pr create --fill
```

Uma branch e um PR por issue, com `Closes #N` na mensagem para fechar automaticamente no merge.

---

## Log

**Status: plano concluído.** Todas as 22 issues de backend previstas nos sprints 0–5
foram implementadas e mergeadas. Épico [#34](https://github.com/LABit-UFPA/phishing-quest-api/issues/34) fechado.

| Data | Sprint | Issue | PR | Observação |
|------|--------|-------|----|------------|
| 17/09 | 1 | #32 suíte de testes | [#35](https://github.com/LABit-UFPA/phishing-quest-api/pull/35) | `MockUserRepository` desatualizado impedia compilar qualquer teste |
| 17/09 | 1 | #12 senha em texto puro | [#36](https://github.com/LABit-UFPA/phishing-quest-api/pull/36) | Maior risco/menor esforço do sprint |
| 17/09 | 1 | #11 path de registro | [#37](https://github.com/LABit-UFPA/phishing-quest-api/pull/37) | Nenhuma mudança no back; ADR 0001 registrou o path canônico |
| 18/09 | 1 | #10 JWT + middleware | [#38](https://github.com/LABit-UFPA/phishing-quest-api/pull/38) | `IJWTService` HS256, `AuthRequired`; desbloqueou o front |
| 18/09 | 2 | #13 `/categories/:id/questions` | [#39](https://github.com/LABit-UFPA/phishing-quest-api/pull/39) | Eram **dois** bugs: param errado e `questionRepo` nunca injetado |
| 18/09 | 2 | #14 score no `/game/answer` | [#40](https://github.com/LABit-UFPA/phishing-quest-api/pull/40) | Descoberto que `user_scores` é **ledger**, não contador: `INSERT` + `SUM`, não upsert |
| 18/09 | 2 | #17 validações de domínio | [#41](https://github.com/LABit-UFPA/phishing-quest-api/pull/41) | `validate:"required"` em `bool` rejeitava `false`; `Question.CorrectAnswer` já não existia no banco |
| 18/09 | 2 | #16 `GET /users/:id` | [#42](https://github.com/LABit-UFPA/phishing-quest-api/pull/42) | Stub `GetTeste` removido |
| 18/09 | 2 | #15 convenção JSON | [#43](https://github.com/LABit-UFPA/phishing-quest-api/pull/43) | `dto/answer.go` migrado para camelCase; issue #38 aberta no front |
| 18/09 | 2 | #18 CORS / erro / env | [#44](https://github.com/LABit-UFPA/phishing-quest-api/pull/44) | `response.Error` centralizado nos 7 handlers |
| 18/09 | 2 | #33 `flyway-prod.config` | [#45](https://github.com/LABit-UFPA/phishing-quest-api/pull/45) | Apontava para diretório inexistente; descoberto que o pipeline real usa ConfigMap |
| 18/09 | 3 | #20 items multicanal | [#46](https://github.com/LABit-UFPA/phishing-quest-api/pull/46) | `content_json` JSONB porque o shape varia por canal. Fechada manualmente: o PR não trazia `Closes #20` |
| 18/09 | 3 | #19 cues + item_cues | [#47](https://github.com/LABit-UFPA/phishing-quest-api/pull/47) | Seed de 10 pistas com UUIDs fixos |
| 18/09 | 3 | #21 attempts | [#48](https://github.com/LABit-UFPA/phishing-quest-api/pull/48) | `isCorrect` recalculado no servidor; consentimento obrigatório |
| 18/09 | 3 | #29 telemetry_events | [#49](https://github.com/LABit-UFPA/phishing-quest-api/pull/49) | Id gerado pelo cliente + `ON CONFLICT DO NOTHING` = ingestão idempotente |
| 18/09 | 4 | #25 consentimento + assessments | [#50](https://github.com/LABit-UFPA/phishing-quest-api/pull/50) | Condição experimental sorteada pelo servidor; `withdraw` (LGPD) |
| 18/09 | 4 | #26 roles + export | [#51](https://github.com/LABit-UFPA/phishing-quest-api/pull/51) | Bug corrigido: o usecase mutava structs in-place, o que quebrava a estabilidade do pseudo-id |
| 18/09 | 4 | #22 `GET /items/next` | [#52](https://github.com/LABit-UFPA/phishing-quest-api/pull/52) | Balanceia malicioso/legítimo na sessão, sem repetir item |
| 18/09 | 4 | #23 `GET /me/stats` | [#53](https://github.com/LABit-UFPA/phishing-quest-api/pull/53) | d' e critério c com correção log-linear de Hautus (evita `Inf`) |
| 18/09 | 5 | #24 `GET /rankings` | [#54](https://github.com/LABit-UFPA/phishing-quest-api/pull/54) | `ROW_NUMBER()` para `position` + JOIN em users; escopo por coorte |
| 18/09 | 5 | #27 revisão espaçada | [#55](https://github.com/LABit-UFPA/phishing-quest-api/pull/55) | Leitner **por pista**, não por item: cada pista evolui independente |
| 18/09 | 5 | #28 seleção adaptativa | [#56](https://github.com/LABit-UFPA/phishing-quest-api/pull/56) | Foca pistas fracas **mantendo** o balanceamento de lado, para não inflar acerto sem ganho de d' |
| 18/09 | 5 | #30 pipeline de itens | [#57](https://github.com/LABit-UFPA/phishing-quest-api/pull/57) | **Bug de segurança:** `POST /api/v1/items` era anônimo e escrevia direto na tabela servida ao jogo |
| 18/09 | 5 | #31 integração + CI | [#58](https://github.com/LABit-UFPA/phishing-quest-api/pull/58) | **Descoberta:** o CI existente só disparava em tag — nunca houve portão de merge |

### Números finais

- **171 testes**: 142 unitários (`tests/`) + 29 de integração (`tests/integration/`, banco real).
- **11 migrations** Flyway, aplicadas do zero a cada execução do CI.
- **CI obrigatório em `main`**: `gofmt` + `go vet` + `go build`, testes unitários e testes de
  integração com Postgres efêmero. Branch protection exige os 3 checks (`strict`, ou seja, a
  branch precisa estar atualizada). `enforce_admins` ficou desligado de propósito, para não
  travar uma correção emergencial.

### Achados que não estavam no plano

Coisas que só apareceram porque cada issue foi validada contra Postgres real antes do merge:

1. **`user_scores` é ledger** (#14) — a issue pedia "upsert"; o schema pede `INSERT` + `SUM`.
   Implementar o upsert teria destruído o histórico.
2. **`POST /api/v1/items` sem autenticação** (#30) — qualquer pessoa com a URL podia injetar
   conteúdo que seria ensinado como verdade ao participante. Não era o objetivo da issue,
   apareceu ao montar o pipeline de curadoria.
3. **O CI nunca rodou em PR** (#31) — `workflow_ci.yml` dispara só em tag `v*.*.*` e em runner
   self-hosted. O repositório passou o projeto inteiro sem verificação automática de PR.
4. **Migration reprovada pelo banco** (#30) — a constraint que exige revisor em item publicado
   falhou contra as linhas legadas (`reviewed_by NULL`). Resolvido com `NOT VALID`, que
   grandfathera o passado e exige a regra em todo `INSERT`/`UPDATE` futuro. Nenhum teste
   unitário pegaria isso.
5. **Contratos divergentes** (#31) — `assessments` usa `instrumentVersion`/`responsesJson`;
   `user_scores` tem coluna `timestamp`, não `created_at`; `GET /rankings` devolve
   `{"ranking": [...]}` e não um array solto.

### Pendências conhecidas (fora do escopo destas issues)

- **Senha em texto puro nos logs**: os logs de `INFO` do repositório genérico imprimem o struct
  `domain.User` inteiro, incluindo o campo `Password` preenchido. A resposta HTTP já foi
  corrigida na #12, mas o log ainda vaza. Registrado como issue separada.
- **Front desatualizado**: issue #38 no repositório `phishing_quest` cobre o ajuste do
  `game_repository.dart` para o camelCase da #15. Nenhum endpoint novo do backend
  (`/attempts`, `/items/next`, `/me/stats`, `/review/due`, `/admin/items`) tem consumidor no
  front ainda.
