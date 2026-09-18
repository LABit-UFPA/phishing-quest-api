package tests

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"phishing-quest/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestUser_ToDTO_NuncaExpoeSenha garante que a resposta de qualquer endpoint
// que devolve um usuário via ToDTO() nunca contém senha em texto puro nem
// hash — regressão da issue #12 (password vazando na resposta de cadastro).
func TestUser_ToDTO_NuncaExpoeSenha(t *testing.T) {
	// domain.User nao tem mais campo de senha em texto puro (issue #60);
	// resta garantir que o hash tambem nao sai na resposta.
	user := &domain.User{
		Id:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "$2a$10$hashsecretoquenaodeveaparecer",
		TotalScore:   42,
	}

	responseDTO := user.ToDTO()

	payload, err := json.Marshal(responseDTO)
	assert.NoError(t, err)

	assert.NotContains(t, string(payload), "hashsecreto")
	assert.NotContains(t, string(payload), "password")
	assert.NotContains(t, string(payload), "passwordHash")

	assert.Equal(t, user.Id, responseDTO.Id)
	assert.Equal(t, user.Username, responseDTO.Username)
	assert.Equal(t, user.Email, responseDTO.Email)
	assert.Equal(t, user.TotalScore, responseDTO.TotalScore)
}

// TestUser_StructNaoTemCampoDeSenhaEmTextoPuro e a regressao estrutural
// da issue #60: a senha vazava porque o repositorio logava a entidade
// com %+v e domain.User carregava a senha em claro. A correcao foi
// remover o campo, nao mascarar o log — este teste falha se alguem
// reintroduzir um campo de senha em texto puro no dominio, o que
// reabriria a porta para vazamento em log/dump/serializacao.
func TestUser_StructNaoTemCampoDeSenhaEmTextoPuro(t *testing.T) {
	userType := reflect.TypeOf(domain.User{})

	for i := 0; i < userType.NumField(); i++ {
		nome := userType.Field(i).Name
		if nome == "PasswordHash" {
			continue // o hash e legitimo e nunca sai em resposta (json:"-")
		}
		assert.NotContains(t, strings.ToLower(nome), "password",
			"domain.User voltou a ter campo de senha em texto puro (%s): use dto.UserRegisterDTO", nome)
		assert.NotContains(t, strings.ToLower(nome), "senha",
			"domain.User voltou a ter campo de senha em texto puro (%s): use dto.UserRegisterDTO", nome)
	}
}

// TestUser_FormatacaoComVerboMaisVNaoVazaSenha simula exatamente o que o
// log fazia antes (%+v da entidade) e garante que nao ha mais senha em
// claro para vazar.
func TestUser_FormatacaoComVerboMaisVNaoVazaSenha(t *testing.T) {
	user := &domain.User{
		Id:           uuid.New(),
		Username:     "logtest",
		Email:        "logtest@example.com",
		PasswordHash: "$2a$10$hash-irrelevante",
	}

	dump := fmt.Sprintf("%+v", user)

	assert.NotContains(t, strings.ToLower(dump), "password:")
	assert.Contains(t, dump, "logtest") // o dump segue util para rastreio
}
