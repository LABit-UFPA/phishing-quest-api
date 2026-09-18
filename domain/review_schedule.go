package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReviewSchedule e o agendamento de revisao espacada (Leitner) para
// uma combinacao usuario+item+pista. Cada pista presente num item
// respondido tem seu proprio agendamento — o usuario pode dominar
// "urgency" mas ainda errar sistematicamente "typosquat", e cada uma
// evolui de forma independente.
type ReviewSchedule struct {
	Id         uuid.UUID `json:"id" gorm:"primaryKey"`
	UserId     uuid.UUID `json:"userId"`
	ItemId     uuid.UUID `json:"itemId"`
	CueId      uuid.UUID `json:"cueId"`
	Box        int       `json:"box"`
	DueAt      time.Time `json:"dueAt"`
	LastResult *bool     `json:"lastResult,omitempty"`
	CreatedAt  time.Time `json:"createdAt,omitempty"`
	UpdatedAt  time.Time `json:"updatedAt,omitempty"`
}

func (rs *ReviewSchedule) TableName() string {
	return "phishing_quest.review_schedule"
}

// MinLeitnerBox e MaxLeitnerBox delimitam as caixas do algoritmo.
const (
	MinLeitnerBox = 1
	MaxLeitnerBox = 5
)

// LeitnerIntervalDays mapeia cada caixa para o intervalo (em dias)
// antes do proximo due_at. Caixa 1 = revisar quase imediatamente
// (mesmo dia/dia seguinte); caixa 5 = espaçamento de 3 semanas.
var LeitnerIntervalDays = map[int]int{
	1: 0,
	2: 1,
	3: 3,
	4: 7,
	5: 21,
}

// ApplyResult atualiza a caixa e o due_at de acordo com o algoritmo
// Leitner classico: erro volta para a caixa 1 (revisao imediata);
// acerto avanca uma caixa (at 'e o maximo), espacando a proxima
// revisao. Isso satisfaz o criterio de aceite da issue #27: "item
// errado reaparece antes; item dominado espaca no tempo".
func (rs *ReviewSchedule) ApplyResult(correct bool, now time.Time) {
	if correct {
		if rs.Box < MaxLeitnerBox {
			rs.Box++
		}
	} else {
		rs.Box = MinLeitnerBox
	}

	rs.LastResult = &correct
	rs.DueAt = now.AddDate(0, 0, LeitnerIntervalDays[rs.Box])
	rs.UpdatedAt = now
}
