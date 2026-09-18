package domain

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Cue e uma pista de phishing da taxonomia usada para anotar items e
// diagnosticar o desempenho do usuario por tipo de pista (ex.:
// "usuario erra sistematicamente typosquat mas acerta urgency").
type Cue struct {
	Id       uuid.UUID `json:"id" gorm:"primaryKey"`
	Code     string    `json:"code" validate:"required,min=1,max=64"`
	LabelPt  string    `json:"labelPt" validate:"required,min=1,max=255"`
	Category string    `json:"category" validate:"required,min=1,max=64"`
}

func (c *Cue) TableName() string {
	return "phishing_quest.cues"
}

func (c *Cue) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

// ItemCue associa um Item a uma Cue presente nele, com o trecho (span)
// onde a pista aparece no conteudo, quando aplicavel.
type ItemCue struct {
	Id        uuid.UUID `json:"id" gorm:"primaryKey"`
	ItemId    uuid.UUID `json:"itemId" validate:"required"`
	CueId     uuid.UUID `json:"cueId" validate:"required"`
	SpanStart *int      `json:"spanStart,omitempty"`
	SpanEnd   *int      `json:"spanEnd,omitempty"`
}

func (ic *ItemCue) TableName() string {
	return "phishing_quest.item_cues"
}

func (ic *ItemCue) Validate() error {
	validate := validator.New()
	return validate.Struct(ic)
}
