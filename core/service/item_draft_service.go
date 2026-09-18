package service

import (
	"encoding/json"
	"fmt"
	"phishing-quest/domain"

	"gorm.io/datatypes"
)

// DraftSpec descreve o rascunho pedido ao gerador: canal, se deve ser
// malicioso e um contexto livre (ex.: "cobranca de fatura de energia").
type DraftSpec struct {
	Channel     domain.Channel
	IsMalicious bool
	Locale      string
	Context     string
}

// IItemDraftService gera o CONTEUDO de um rascunho de item. E o ponto
// de extensao previsto na issue #30 para a integracao opcional com LLM:
// basta implementar esta interface com um provedor real (OpenAI,
// Bedrock, etc) e injeta-lo no container.
//
// Independente do provedor, o resultado entra como rascunho e continua
// obrigado a passar pela revisao humana antes de ser publicado — a
// geracao automatica nunca encurta o pipeline.
type IItemDraftService interface {
	GenerateItemDraft(spec DraftSpec) (*domain.Item, error)
}

// TemplateDraftService e a implementacao default: monta o rascunho
// localmente a partir de templates por canal, sem chamada de rede nem
// credencial de provedor. Serve para o pipeline de curadoria ser
// exercitavel de ponta a ponta (e testavel de forma deterministica)
// antes de existir contrato com algum fornecedor de LLM.
type TemplateDraftService struct{}

func NewTemplateDraftService() IItemDraftService {
	return &TemplateDraftService{}
}

func (tds *TemplateDraftService) GenerateItemDraft(spec DraftSpec) (*domain.Item, error) {
	locale := spec.Locale
	if locale == "" {
		locale = "pt-BR"
	}

	content, err := buildDraftContent(spec)
	if err != nil {
		return nil, err
	}

	explanation := "Rascunho gerado automaticamente. REVISAR: confirmar plausibilidade, corrigir o texto e anotar as pistas antes de publicar."
	if !spec.IsMalicious {
		explanation = "Rascunho de item LEGITIMO gerado automaticamente. REVISAR: garantir que nao contem pistas de golpe (senao vira falso positivo no instrumento)."
	}

	return &domain.Item{
		Channel:     spec.Channel,
		IsMalicious: spec.IsMalicious,
		Locale:      locale,
		ContentJSON: content,
		Explanation: explanation,
		// source registra a procedencia para o revisor saber que o texto
		// nao foi escrito por uma pessoa.
		Source: "llm_draft",
		Status: domain.StatusDraft,
	}, nil
}

// buildDraftContent monta o content_json no shape esperado por cada
// canal. O shape varia por canal (mesma razao pela qual a coluna e
// JSONB e nao colunas fixas).
func buildDraftContent(spec DraftSpec) (datatypes.JSON, error) {
	context := spec.Context
	if context == "" {
		context = "assunto administrativo generico"
	}

	var payload map[string]string

	switch spec.Channel {
	case domain.ChannelEmail:
		payload = map[string]string{
			"subject": fmt.Sprintf("[RASCUNHO] %s", context),
			"body":    fmt.Sprintf("Rascunho de email sobre %s. Substituir por texto revisado.", context),
			"sender":  "rascunho@exemplo.invalid",
		}
	case domain.ChannelSMS, domain.ChannelWhatsApp:
		payload = map[string]string{
			"text":   fmt.Sprintf("[RASCUNHO] Mensagem sobre %s. Substituir por texto revisado.", context),
			"sender": "+550000000000",
		}
	case domain.ChannelWebsite:
		payload = map[string]string{
			"url":   "https://exemplo.invalid/rascunho",
			"title": fmt.Sprintf("[RASCUNHO] %s", context),
		}
	case domain.ChannelPhoneCall:
		payload = map[string]string{
			"transcript": fmt.Sprintf("[RASCUNHO] Transcricao de ligacao sobre %s. Substituir por texto revisado.", context),
			"caller":     "+550000000000",
		}
	case domain.ChannelPixQR:
		payload = map[string]string{
			"payload":   "00020126BR.GOV.BCB.PIX-RASCUNHO",
			"recipient": fmt.Sprintf("[RASCUNHO] %s", context),
		}
	default:
		// Canal desconhecido e recusado aqui em vez de gerar um shape
		// vazio que passaria a validacao e confundiria o revisor.
		return nil, fmt.Errorf("canal nao suportado pelo gerador de rascunho: %s", spec.Channel)
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(encoded), nil
}
