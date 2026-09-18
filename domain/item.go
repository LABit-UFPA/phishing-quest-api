package domain

import (
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

// Item generaliza o antigo conceito de "phishing email" (so mockado no
// front) para qualquer canal de golpe simulado. ContentJSON guarda o
// conteudo especifico do canal (assunto+corpo, texto de SMS, etc).
type Item struct {
	Id                         uuid.UUID      `json:"id" gorm:"primaryKey"`
	Channel                    Channel        `json:"channel" validate:"required,oneof=email sms whatsapp website phone_call pix_qr"`
	IsMalicious                bool           `json:"isMalicious"`
	Locale                     string         `json:"locale" validate:"required"`
	PhishScaleCueCount         *int           `json:"phishScaleCueCount,omitempty"`
	PhishScalePremiseAlignment *string        `json:"phishScalePremiseAlignment,omitempty"`
	DifficultyCalibrated       *string        `json:"difficultyCalibrated,omitempty"`
	ContentJSON                datatypes.JSON `json:"contentJson" validate:"required" gorm:"column:content_json"`
	Explanation                string         `json:"explanation"`
	Source                     string         `json:"source"`
	ReviewedBy                 *uuid.UUID     `json:"reviewedBy,omitempty"`
}

func (i *Item) TableName() string {
	return "phishing_quest.items"
}

func (i *Item) Validate() error {
	validate := validator.New()
	return validate.Struct(i)
}
