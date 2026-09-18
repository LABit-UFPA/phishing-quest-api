package usecase

import (
	"errors"
	"phishing-quest/adapter/repository"
	"phishing-quest/domain"

	"github.com/google/uuid"
)

// SelectionMode enumera as estrategias de selecao de GET /items/next.
// ModeAdaptive e o gancho previsto para a issue #28 (selecao
// adaptativa por dominio de pista) — por enquanto cai no mesmo
// comportamento de ModeBalanced, ja que a logica de mastery por pista
// ainda nao existe.
type SelectionMode string

const (
	ModeBalanced SelectionMode = "balanced"
	ModeAdaptive SelectionMode = "adaptive"
)

var ErrNoUnseenItems = errors.New("nao ha mais items disponiveis para esta sessao")

type ItemSelectionUseCase struct {
	itemRepo repository.IItemRepository
}

func NewItemSelectionUseCase(itemRepo repository.IItemRepository) *ItemSelectionUseCase {
	return &ItemSelectionUseCase{itemRepo: itemRepo}
}

// NextItem seleciona o proximo item para o usuario dentro de uma
// sessao, sem repetir items ja respondidos nela (via NOT IN sobre
// attempts.item_id). O modo "balanced" (piloto) equilibra a
// proporcao malicioso/legitimo dentro da sessao: se o usuario ja viu
// mais itens maliciosos que legitimos (ou vice-versa), prioriza o
// lado com menos exposicao, para o instrumento nao ficar desbalanceado
// so pelo acaso da aleatoriedade.
//
// O parametro mode existe como gancho para a issue #28 (selecao
// adaptativa por dominio de pista, priorizando pistas com pior
// desempenho do usuario) — ainda nao implementado, cai no mesmo
// comportamento de "balanced".
func (isu *ItemSelectionUseCase) NextItem(userID, sessionID uuid.UUID, mode SelectionMode) (*domain.Item, error) {
	maliciousSeen, legitimateSeen, err := isu.itemRepo.CountSeenInSession(userID, sessionID)
	if err != nil {
		return nil, err
	}

	var preferMalicious *bool
	switch {
	case maliciousSeen < legitimateSeen:
		v := true
		preferMalicious = &v
	case legitimateSeen < maliciousSeen:
		v := false
		preferMalicious = &v
	default:
		preferMalicious = nil // empatado: qualquer lado serve
	}

	item, err := isu.itemRepo.GetRandomUnseen(userID, sessionID, preferMalicious)
	if err != nil {
		// Se o lado preferido nao tem mais itens disponiveis (ex.: so
		// restam legitimos, mas o balanceamento pediu malicioso), cai
		// para "qualquer item nao visto" em vez de falhar — melhor
		// entregar um item desbalanceado do que travar a sessao.
		if preferMalicious != nil {
			item, err = isu.itemRepo.GetRandomUnseen(userID, sessionID, nil)
		}
		if err != nil {
			return nil, ErrNoUnseenItems
		}
	}

	return item, nil
}
