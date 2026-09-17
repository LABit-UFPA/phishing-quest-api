package tests

import (
	"encoding/json"
	"testing"

	"phishing-quest/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestUser_ToDTO_NuncaExpoeSenha garante que a resposta de qualquer endpoint
// que devolve um usuário via ToDTO() nunca contém senha em texto puro nem
// hash — regressão da issue #12 (password vazando na resposta de cadastro).
func TestUser_ToDTO_NuncaExpoeSenha(t *testing.T) {
	user := &domain.User{
		Id:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		Password:     "senha-em-texto-puro",
		PasswordHash: "$2a$10$hashsecretoquenaodeveaparecer",
		TotalScore:   42,
	}

	responseDTO := user.ToDTO()

	payload, err := json.Marshal(responseDTO)
	assert.NoError(t, err)

	assert.NotContains(t, string(payload), "senha-em-texto-puro")
	assert.NotContains(t, string(payload), "hashsecreto")
	assert.NotContains(t, string(payload), "password")
	assert.NotContains(t, string(payload), "passwordHash")

	assert.Equal(t, user.Id, responseDTO.Id)
	assert.Equal(t, user.Username, responseDTO.Username)
	assert.Equal(t, user.Email, responseDTO.Email)
	assert.Equal(t, user.TotalScore, responseDTO.TotalScore)
}
