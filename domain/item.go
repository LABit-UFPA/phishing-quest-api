package domain

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Channel enumera os canais de golpe simulados. Definido como string
// (nao um tipo enum do banco) para facilitar evolucao sem migration
// adicional; a validacao de valores permitidos e feita em Validate()
// e reforcada por uma CHECK constraint no banco (ver migration).
type Channel string

const (
	ChannelEmail     Channel = "email"
	ChannelSMS       Channel = "sms"
	ChannelWhatsApp  Channel = "whatsapp"
	ChannelWebsite   Channel = "website"
	ChannelPhoneCall Channel = "phone_call"
	ChannelPixQR     Channel = "pix_qr"
)

// ItemStatus enumera os estados do pipeline de curadoria de itens
// (issue #30). Um item errado ENSINA errado, entao nenhum item chega
// ao participante sem revisao humana registrada: so o estado
// published e servido pelo jogo.
type ItemStatus string

const (
	// StatusDraft e o estado inicial de qualquer item criado pela API,
	// inclusive os gerados automaticamente por LLM.
	StatusDraft ItemStatus = "draft"
	// StatusReviewed indica que uma pessoa (admin/researcher) revisou o
	// conteudo e assinou a revisao (ReviewedBy).
	StatusReviewed ItemStatus = "reviewed"
	// StatusPublished e o unico estado servido aos participantes.
	StatusPublished ItemStatus = "published"
)

// ErrInvalidItemTransition e retornado quando se tenta uma transicao
// fora do fluxo draft -> reviewed -> published. Em especial, pular a
// revisao (draft -> published) e sempre recusado.
var ErrInvalidItemTransition = errors.New("transicao de estado invalida para o item")

// ErrReviewerRequired e retornado quando se tenta registrar revisao
// sem identificar o revisor.
var ErrReviewerRequired = errors.New("revisao exige um revisor identificado")

// Item generaliza o antigo conceito de "phishing email" (so mockado no
// front) para qualquer canal de golpe simulado. ContentJSON guarda o
// conteudo especifico do canal (assunto+corpo, texto de SMS, etc).
type Item struct {
	Id                         uuid.UUID `json:"id" gorm:"primaryKey"`
	Channel                    Channel   `json:"channel" validate:"required,oneof=email sms whatsapp website phone_call pix_qr"`
	IsMalicious                bool      `json:"isMalicious"`
	Locale                     string    `json:"locale" validate:"required"`
	PhishScaleCueCount         *int      `json:"phishScaleCueCount,omitempty"`
	PhishScalePremiseAlignment *string   `json:"phishScalePremiseAlignment,omitempty"`
	// DifficultyEstimated e a priori (do gerador, phishforge-api #9);
	// DifficultyCalibrated e a posteriori (medida a partir de attempts
	// reais, issue #66). Colunas SEPARADAS de proposito -- ver
	// migration V20260919140000 e issue #67: gravar a estimativa em
	// DifficultyCalibrated sobrescreveria o dado que a calibracao
	// deveria estar avaliando.
	DifficultyEstimated  *string        `json:"difficultyEstimated,omitempty"`
	DifficultyCalibrated *string        `json:"difficultyCalibrated,omitempty"`
	ContentJSON          datatypes.JSON `json:"contentJson" validate:"required" gorm:"column:content_json"`
	Explanation          string         `json:"explanation"`
	Source               string         `json:"source"`
	Status               ItemStatus     `json:"status" validate:"required,oneof=draft reviewed published"`
	ReviewedBy           *uuid.UUID     `json:"reviewedBy,omitempty"`
	ReviewedAt           *time.Time     `json:"reviewedAt,omitempty"`
	PublishedAt          *time.Time     `json:"publishedAt,omitempty"`
}

func (i *Item) TableName() string {
	return "phishing_quest.items"
}

func (i *Item) Validate() error {
	validate := validator.New()
	return validate.Struct(i)
}

// MarkReviewed registra a revisao humana do item. Só é aceita a partir
// de draft: revisar de novo um item já revisado ou já publicado seria
// sobrescrever a assinatura de quem revisou antes.
func (i *Item) MarkReviewed(reviewerID uuid.UUID, now time.Time) error {
	if i.Status != StatusDraft {
		return ErrInvalidItemTransition
	}
	if reviewerID == uuid.Nil {
		return ErrReviewerRequired
	}

	i.Status = StatusReviewed
	i.ReviewedBy = &reviewerID
	i.ReviewedAt = &now
	return nil
}

// Publish libera o item para o jogo. Exige que o item esteja revisado
// E que a revisao tenha um responsavel registrado — esta segunda
// checagem e redundante com o fluxo de estados, mas e o criterio de
// aceite da issue #30 e vale ser explicita: nenhum caminho de codigo
// publica item sem revisao humana atribuida.
func (i *Item) Publish(now time.Time) error {
	if i.Status != StatusReviewed {
		return ErrInvalidItemTransition
	}
	if i.ReviewedBy == nil {
		return ErrReviewerRequired
	}

	i.Status = StatusPublished
	i.PublishedAt = &now
	return nil
}
