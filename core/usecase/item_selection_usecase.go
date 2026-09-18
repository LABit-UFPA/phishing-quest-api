package usecase

import (
	"errors"
	"os"
	"phishing-quest/adapter/repository"
	"phishing-quest/domain"
	"sort"
	"strconv"

	"github.com/google/uuid"
)

// SelectionMode enumera as estrategias de selecao de GET /items/next.
type SelectionMode string

const (
	// ModeBalanced equilibra a proporcao malicioso/legitimo na sessao.
	ModeBalanced SelectionMode = "balanced"
	// ModeAdaptive prioriza items que contenham as pistas que o
	// usuario ainda erra (issue #28), mantendo o balanceamento de lado
	// como critério secundário.
	ModeAdaptive SelectionMode = "adaptive"
)

var ErrNoUnseenItems = errors.New("nao ha mais items disponiveis para esta sessao")

// errNoAdaptiveCandidate e interno: sinaliza que a selecao adaptativa
// nao encontrou candidato (usuario sem pista fraca diagnosticavel, ou
// sem item nao visto contendo essas pistas). Nunca sai do usecase —
// vira fallback para o modo balanceado.
var errNoAdaptiveCandidate = errors.New("sem candidato adaptativo")

// MasteryParams sao os parametros de "dominio de pista" da selecao
// adaptativa, configuraveis por ambiente (criterio de aceite da issue
// #28). Os defaults valem para o piloto e devem ser recalibrados com
// dados reais antes do estudo principal.
type MasteryParams struct {
	// Threshold e a acuracia a partir da qual a pista e considerada
	// dominada. Abaixo dela, a pista entra na fila de reforco.
	Threshold float64
	// MinExposures e o minimo de tentativas com a pista antes de
	// julgar dominio. Evita reagir a ruido: errar 1 de 1 nao
	// caracteriza dificuldade sistematica.
	MinExposures int
	// MaxWeakCues limita quantas pistas fracas entram no foco de uma
	// vez, para a selecao nao virar "qualquer item" quando o usuario
	// vai mal em quase tudo (o que anularia o efeito adaptativo).
	MaxWeakCues int
}

const (
	defaultMasteryThreshold    = 0.75
	defaultMasteryMinExposures = 3
	defaultMasteryMaxWeakCues  = 3
)

// loadMasteryParams le os parametros do ambiente, caindo nos defaults
// quando a variavel esta ausente ou invalida (nunca sobe com valor
// absurdo por erro de digitacao no .env).
func loadMasteryParams() MasteryParams {
	params := MasteryParams{
		Threshold:    defaultMasteryThreshold,
		MinExposures: defaultMasteryMinExposures,
		MaxWeakCues:  defaultMasteryMaxWeakCues,
	}

	if raw := os.Getenv("ADAPTIVE_MASTERY_THRESHOLD"); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil && v > 0 && v <= 1 {
			params.Threshold = v
		}
	}
	if raw := os.Getenv("ADAPTIVE_MASTERY_MIN_EXPOSURES"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			params.MinExposures = v
		}
	}
	if raw := os.Getenv("ADAPTIVE_MASTERY_MAX_WEAK_CUES"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			params.MaxWeakCues = v
		}
	}

	return params
}

type ItemSelectionUseCase struct {
	itemRepo  repository.IItemRepository
	statsRepo repository.IUserStatsRepository
	mastery   MasteryParams
}

func NewItemSelectionUseCase(itemRepo repository.IItemRepository, statsRepo repository.IUserStatsRepository) *ItemSelectionUseCase {
	return &ItemSelectionUseCase{
		itemRepo:  itemRepo,
		statsRepo: statsRepo,
		mastery:   loadMasteryParams(),
	}
}

// NextItem seleciona o proximo item para o usuario dentro de uma
// sessao, sem repetir items ja respondidos nela (via NOT IN sobre
// attempts.item_id).
//
// O modo "balanced" (piloto) equilibra a proporcao malicioso/legitimo
// dentro da sessao: se o usuario ja viu mais itens maliciosos que
// legitimos (ou vice-versa), prioriza o lado com menos exposicao, para
// o instrumento nao ficar desbalanceado so pelo acaso.
//
// O modo "adaptive" (issue #28) concentra a pratica nas pistas que o
// usuario ainda nao domina, mas mantem o balanceamento de lado como
// critério secundário — sem isso, focar em pistas (que so existem em
// items maliciosos, na pratica) enviesaria a sessao para "tudo e
// phishing", inflando a taxa de acerto sem ganho de discriminacao
// (d'). Se nao houver pista fraca diagnosticavel ou item nao visto que
// a contenha, cai no comportamento balanceado.
func (isu *ItemSelectionUseCase) NextItem(userID, sessionID uuid.UUID, mode SelectionMode) (*domain.Item, error) {
	preferMalicious, err := isu.preferredSide(userID, sessionID)
	if err != nil {
		return nil, err
	}

	if mode == ModeAdaptive {
		item, err := isu.nextAdaptive(userID, sessionID, preferMalicious)
		if err == nil {
			return item, nil
		}
		if !errors.Is(err, errNoAdaptiveCandidate) {
			// Erro real de banco na leitura de mastery: propaga em vez
			// de mascarar como "nao ha pista fraca".
			return nil, err
		}
	}

	return isu.nextBalanced(userID, sessionID, preferMalicious)
}

// preferredSide decide qual lado (malicioso/legitimo) esta
// sub-representado na sessao. nil = empatado, qualquer lado serve.
func (isu *ItemSelectionUseCase) preferredSide(userID, sessionID uuid.UUID) (*bool, error) {
	maliciousSeen, legitimateSeen, err := isu.itemRepo.CountSeenInSession(userID, sessionID)
	if err != nil {
		return nil, err
	}

	switch {
	case maliciousSeen < legitimateSeen:
		v := true
		return &v, nil
	case legitimateSeen < maliciousSeen:
		v := false
		return &v, nil
	default:
		return nil, nil
	}
}

// nextAdaptive tenta um item que contenha alguma das pistas fracas do
// usuario, primeiro respeitando o lado sub-representado e, se nao
// houver, em qualquer lado (a pista importa mais que o balanceamento
// nesse ponto).
func (isu *ItemSelectionUseCase) nextAdaptive(userID, sessionID uuid.UUID, preferMalicious *bool) (*domain.Item, error) {
	weakCueIDs, err := isu.WeakCues(userID)
	if err != nil {
		return nil, err
	}
	if len(weakCueIDs) == 0 {
		return nil, errNoAdaptiveCandidate
	}

	if preferMalicious != nil {
		if item, err := isu.itemRepo.GetRandomUnseenByCues(userID, sessionID, weakCueIDs, preferMalicious); err == nil {
			return item, nil
		}
	}

	item, err := isu.itemRepo.GetRandomUnseenByCues(userID, sessionID, weakCueIDs, nil)
	if err != nil {
		return nil, errNoAdaptiveCandidate
	}
	return item, nil
}

// WeakCues retorna os ids das pistas que o usuario ainda nao domina,
// das piores para as menos piores, limitado a MaxWeakCues. Pistas com
// menos de MinExposures respostas sao ignoradas: sem exposicao
// suficiente nao ha como distinguir dificuldade real de azar.
func (isu *ItemSelectionUseCase) WeakCues(userID uuid.UUID) ([]uuid.UUID, error) {
	mastery, err := isu.statsRepo.GetCueMastery(userID)
	if err != nil {
		return nil, err
	}

	type scored struct {
		cueID    uuid.UUID
		accuracy float64
	}

	weak := make([]scored, 0, len(mastery))
	for _, m := range mastery {
		if m.Answered < isu.mastery.MinExposures {
			continue
		}
		accuracy := float64(m.Correct) / float64(m.Answered)
		if accuracy < isu.mastery.Threshold {
			weak = append(weak, scored{cueID: m.CueId, accuracy: accuracy})
		}
	}

	// Pior desempenho primeiro; desempate estavel pelo id da pista
	// para a selecao nao variar entre chamadas identicas.
	sort.Slice(weak, func(i, j int) bool {
		if weak[i].accuracy != weak[j].accuracy {
			return weak[i].accuracy < weak[j].accuracy
		}
		return weak[i].cueID.String() < weak[j].cueID.String()
	})

	if len(weak) > isu.mastery.MaxWeakCues {
		weak = weak[:isu.mastery.MaxWeakCues]
	}

	cueIDs := make([]uuid.UUID, 0, len(weak))
	for _, w := range weak {
		cueIDs = append(cueIDs, w.cueID)
	}
	return cueIDs, nil
}

// nextBalanced e a selecao do piloto: lado sub-representado primeiro,
// com fallback para qualquer item nao visto.
func (isu *ItemSelectionUseCase) nextBalanced(userID, sessionID uuid.UUID, preferMalicious *bool) (*domain.Item, error) {
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
