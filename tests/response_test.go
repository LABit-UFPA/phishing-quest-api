package tests

import (
	"testing"

	"phishing-quest/adapter/http/response"

	"github.com/stretchr/testify/assert"
)

func TestResponse_Error_MontaFormatoPadrao(t *testing.T) {
	body := response.Error("algo deu errado")

	assert.Equal(t, "algo deu errado", body["error"])
	assert.Len(t, body, 1)
}
