# ADR 0001 — Path canônico de registro de usuário

- **Status:** aceita
- **Data:** 17/09/2026
- **Issue:** [phishing-quest-api#11](https://github.com/LABit-UFPA/phishing-quest-api/issues/11), [phishing_quest#7](https://github.com/LABit-UFPA/phishing_quest/issues/7)

## Contexto

O front (`register_repository.dart`) chamava `POST /users`. O backend expõe
`POST /users/register` (`adapter/http/router/user_router.go`). Os dois não
batiam e o cadastro real falhava (404).

## Decisão

Path canônico: **`POST /api/v1/users/register`**.

Nenhuma mudança de código no backend — o path já era esse. O ajuste foi
inteiramente no front, que passou a chamar `/users/register`.

## Alternativas consideradas

- Adicionar um alias `POST /api/v1/users` no backend apontando para o mesmo
  handler (`CreateUser`). Rejeitada: mantém duas formas de fazer a mesma
  coisa sem necessidade, e `/users/register` já segue a mesma convenção de
  `/users/login` usada no restante do grupo de rotas.

## Consequências

- Nenhuma migration, nenhuma alteração de contrato.
- `../../INTEGRACAO_FRONT_BACK.md` atualizado para refletir a resolução.
