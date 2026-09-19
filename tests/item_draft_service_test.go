package tests

import (
	"encoding/json"
	"testing"

	"phishing-quest/core/service"
	"phishing-quest/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTemplateDraftService_Email_LinksVemComoObjetos e a regressao
// central da issue #68: content_json de email precisa aceitar `links`
// como array de objetos {text, href} -- nao mais List[str] -- para o
// lado Python (phishforge-api #5) poder expressar a pista
// link_text_mismatch (texto do link diferente do destino real).
func TestTemplateDraftService_Email_LinksVemComoObjetos(t *testing.T) {
	tds := service.NewTemplateDraftService()

	item, err := tds.GenerateItemDraft(service.DraftSpec{
		Channel:     domain.ChannelEmail,
		IsMalicious: true,
		Context:     "cobranca de fatura",
	})
	require.NoError(t, err)

	var content struct {
		Sender  string `json:"sender"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
		Links   []struct {
			Text string `json:"text"`
			Href string `json:"href"`
		} `json:"links"`
	}
	require.NoError(t, json.Unmarshal(item.ContentJSON, &content))

	assert.NotEmpty(t, content.Sender)
	assert.NotEmpty(t, content.Subject)
	assert.NotEmpty(t, content.Body)
	if assert.Len(t, content.Links, 1) {
		assert.NotEmpty(t, content.Links[0].Text)
		assert.NotEmpty(t, content.Links[0].Href)
	}
}

// TestTemplateDraftService_SMS_LinksVemComoObjetos cobre o mesmo
// contrato para SMS -- o outro canal plano que ganha `links`
// estruturado na issue #68.
func TestTemplateDraftService_SMS_LinksVemComoObjetos(t *testing.T) {
	tds := service.NewTemplateDraftService()

	item, err := tds.GenerateItemDraft(service.DraftSpec{
		Channel:     domain.ChannelSMS,
		IsMalicious: true,
		Context:     "confirmacao de entrega",
	})
	require.NoError(t, err)

	var content struct {
		Text  string `json:"text"`
		Links []struct {
			Text string `json:"text"`
			Href string `json:"href"`
		} `json:"links"`
	}
	require.NoError(t, json.Unmarshal(item.ContentJSON, &content))

	assert.NotEmpty(t, content.Text)
	assert.Len(t, content.Links, 1)
}

// TestTemplateDraftService_WhatsApp_MessagesVemComoObjetos e a
// regressao central da issue #68 para o canal WhatsApp: `messages`
// precisa ser um array de objetos {author, text} (historico de
// conversa), nao uma unica string -- shape acordado no comentario da
// phishforge-api #6.
func TestTemplateDraftService_WhatsApp_MessagesVemComoObjetos(t *testing.T) {
	tds := service.NewTemplateDraftService()

	item, err := tds.GenerateItemDraft(service.DraftSpec{
		Channel:     domain.ChannelWhatsApp,
		IsMalicious: true,
		Context:     "premio de sorteio",
	})
	require.NoError(t, err)

	var content struct {
		Sender      string `json:"sender"`
		DisplayName string `json:"display_name"`
		Messages    []struct {
			Author string `json:"author"`
			Text   string `json:"text"`
		} `json:"messages"`
	}
	require.NoError(t, json.Unmarshal(item.ContentJSON, &content))

	assert.NotEmpty(t, content.Sender)
	assert.NotEmpty(t, content.DisplayName)
	if assert.Len(t, content.Messages, 1) {
		assert.NotEmpty(t, content.Messages[0].Author)
		assert.NotEmpty(t, content.Messages[0].Text)
	}
}

// TestTemplateDraftService_CanaisPlanosContinuamFuncionando e a
// regressao de nao-quebra: website/phone_call/pix_qr nao foram
// alterados pela issue #68 (shape plano, sem aninhamento), e a troca
// do tipo intermediario para map[string]any nao pode mudar o
// conteudo desses tres canais.
func TestTemplateDraftService_CanaisPlanosContinuamFuncionando(t *testing.T) {
	tds := service.NewTemplateDraftService()

	casos := []struct {
		canal           domain.Channel
		camposEsperados []string
	}{
		{domain.ChannelWebsite, []string{"url", "title"}},
		{domain.ChannelPhoneCall, []string{"transcript", "caller"}},
		{domain.ChannelPixQR, []string{"payload", "recipient"}},
	}

	for _, caso := range casos {
		item, err := tds.GenerateItemDraft(service.DraftSpec{
			Channel:     caso.canal,
			IsMalicious: true,
			Context:     "contexto de teste",
		})
		require.NoError(t, err)

		var content map[string]any
		require.NoError(t, json.Unmarshal(item.ContentJSON, &content))

		for _, campo := range caso.camposEsperados {
			assert.Contains(t, content, campo, "canal %s deveria ter o campo %s", caso.canal, campo)
		}
	}
}
