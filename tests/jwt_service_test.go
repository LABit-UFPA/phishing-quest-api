package tests

import (
	"os"
	"testing"
	"time"

	"phishing-quest/core/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestJWTService_GenerateAndParse(t *testing.T) {
	t.Setenv("JWT_SECRET", "segredo-de-teste")
	t.Setenv("JWT_EXPIRES_IN", "1h")

	svc := service.NewJWTService()
	userID := uuid.New()

	token, err := svc.Generate(userID, "player")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := svc.Parse(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, "player", claims.Role)
}

func TestJWTService_ParseRejectsTamperedToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "segredo-de-teste")

	svc := service.NewJWTService()
	token, err := svc.Generate(uuid.New(), "player")
	assert.NoError(t, err)

	// Token gerado com um segredo diferente nao deve validar contra o servico atual.
	os.Setenv("JWT_SECRET", "outro-segredo")
	outroServico := service.NewJWTService()

	_, err = outroServico.Parse(token)
	assert.Error(t, err)
}

func TestJWTService_ParseRejectsExpiredToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "segredo-de-teste")
	t.Setenv("JWT_EXPIRES_IN", "1ms")

	svc := service.NewJWTService()
	token, err := svc.Generate(uuid.New(), "player")
	assert.NoError(t, err)

	time.Sleep(5 * time.Millisecond)

	_, err = svc.Parse(token)
	assert.Error(t, err)
}
