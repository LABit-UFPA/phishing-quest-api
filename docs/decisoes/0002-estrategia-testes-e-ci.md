# ADR 0002 — Estratégia de testes de integração e CI

- **Status:** aceita
- **Data:** 18/09/2026
- **Issue:** [phishing-quest-api#31](https://github.com/LABit-UFPA/phishing-quest-api/issues/31)

## Contexto

O repositório já tinha `.github/workflows/workflow_ci.yml`, mas ele dispara
**somente em tag** (`v*.*.*`) e roda em runner `self-hosted`. Ou seja: nunca
executou em pull request e nunca funcionou como portão de merge. Na prática,
não havia CI verificando PR nenhum.

Além disso, toda a suíte vivia em `tests/` com mocks. Isso pega erro de lógica,
mas não pega o que mais quebrou durante o desenvolvimento das issues de
pesquisa: SQL inválido, migration que falha, middleware que não está na cadeia
da rota, contrato JSON divergente. A migration da issue #30, por exemplo,
passou nos testes unitários e **falhou no banco real**.

## Decisão

### 1. Novo workflow em vez de alterar o existente

Criado `.github/workflows/workflow_pr_ci.yml`, que dispara em `pull_request`
para `main` e em `push` para `main`, rodando em `ubuntu-latest`.

O `workflow_ci.yml` foi deixado **intacto**: ele é o pipeline de release
(build + push da imagem + deploy no Kubernetes) e depende de runner
self-hosted e de segredos. Misturar o portão de PR com o pipeline de deploy
faria o gate de merge depender da disponibilidade do runner self-hosted e do
acesso ao cluster — dois motivos de falha que não têm nada a ver com a
qualidade do código sob revisão.

### 2. Testes de integração atrás da build tag `integration`

Os testes em `tests/integration/` exigem Postgres migrado. Ficam atrás de
`//go:build integration` para que `go test ./...` continue rodando sem banco
(no dia a dia e no job de testes unitários).

Quando são executados (`-tags=integration`), a ausência do banco **falha** a
suíte em vez de pular. Um teste que se auto-desativa quando o ambiente não
está pronto deixaria o CI verde sem ter testado nada — o oposto do critério
de aceite da issue.

### 3. Os testes de integração sobem o app real

`TestMain` monta `container.NewContainer()` + `apphttp.SetupRouter(...)`, o
mesmo grafo de dependências e o mesmo router de produção, e exercita tudo via
`httptest` contra o Postgres de verdade. Nenhum middleware fica de fora — foi
assim que passou a existir cobertura para "a rota protegida realmente tem
`AuthRequired` na cadeia", que os testes de handler isolados não cobrem porque
montam o próprio `gin.New()`.

### 4. Migrations aplicadas pelo flyway no CI, não por código Go

O job de integração roda a **mesma imagem** `flyway/flyway:7.0.2` e os
**mesmos changelogs** do `docker-compose.yml`. Consequência desejada:
migration quebrada reprova o CI.

Alternativa rejeitada: aplicar os `.sql` por código no `TestMain`. Além de
duplicar o que o flyway já faz, o driver pgx (protocolo estendido) não aceita
múltiplas instruções por `Exec`, o que exigiria fatiar os arquivos por `;` —
frágil e, pior, testaria um caminho de migração **diferente** do que roda em
produção.

### 5. Dados de referência preservados entre testes

`resetDB` trunca as tabelas voláteis mas **não** `cues`: as 10 pistas da
taxonomia são dados de referência inseridos pela migration
`V20260917110000`. Apagá-las quebraria os testes de pista e divergiria do
estado real do banco.

## Consequências

- Todo PR passa por: `gofmt`, `go vet` (inclusive com a tag `integration`),
  `go build`, testes unitários e testes de integração com banco efêmero.
- O critério "CI verde obrigatório para merge" depende também de **branch
  protection** no GitHub exigindo estes checks — configuração de repositório,
  não de arquivo.
- Rodar localmente:

```bash
# unitários (sem banco)
go test ./...

# integração (precisa do banco migrado)
docker-compose up -d phishing-quest-postgresql phishing-quest-flyway-phishing-quest
go test -tags=integration ./tests/integration/ -v
```
