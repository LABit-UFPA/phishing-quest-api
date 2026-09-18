package usecase

import (
	"math"
	"phishing-quest/adapter/repository"
	"phishing-quest/domain"

	"github.com/google/uuid"
)

type UserStatsUseCase struct {
	statsRepo repository.IUserStatsRepository
}

func NewUserStatsUseCase(statsRepo repository.IUserStatsRepository) *UserStatsUseCase {
	return &UserStatsUseCase{statsRepo: statsRepo}
}

// GetStats calcula o desempenho do usuario via teoria de deteccao de
// sinal, mais o breakdown por pista (ROADMAP_PESQUISA_2027.md, Fase 2
// e 3).
func (usu *UserStatsUseCase) GetStats(userID uuid.UUID) (*domain.UserStats, error) {
	attempts, err := usu.statsRepo.GetSignalDetectionAttempts(userID)
	if err != nil {
		return nil, err
	}

	cueOutcomes, err := usu.statsRepo.GetCueOutcomes(userID)
	if err != nil {
		return nil, err
	}

	stats := computeSignalDetectionStats(attempts)
	stats.ByCue = computeCuePerformance(cueOutcomes)
	return stats, nil
}

// computeSignalDetectionStats implementa a teoria de deteccao de
// sinal com a correcao log-linear de Hautus (1995): soma 0.5 a
// hits/falsos-alarmes e 1 ao total de cada condicao antes de calcular
// as taxas. Isso evita z-scores infinitos quando a taxa bruta e 0%
// ou 100% (o que aconteceria com poucos itens ou desempenho perfeito/
// nulo), sem exigir logica condicional ad-hoc por caso.
//
// d' (discriminacao) = z(hitRateAdj) - z(falseAlarmRateAdj)
// c (criterio/vies)  = -0.5 * (z(hitRateAdj) + z(falseAlarmRateAdj))
func computeSignalDetectionStats(attempts []repository.SignalDetectionAttempt) *domain.UserStats {
	var maliciousTotal, maliciousHits int
	var legitimateTotal, falseAlarms int
	var correctTotal int

	for _, a := range attempts {
		if a.ItemIsMalicious {
			maliciousTotal++
			if a.Verdict {
				maliciousHits++
				correctTotal++
			}
		} else {
			legitimateTotal++
			if a.Verdict {
				falseAlarms++
			} else {
				correctTotal++
			}
		}
	}

	totalAttempts := maliciousTotal + legitimateTotal

	stats := &domain.UserStats{TotalAttempts: totalAttempts}
	if totalAttempts == 0 {
		return stats
	}

	stats.Accuracy = float64(correctTotal) / float64(totalAttempts)

	// Taxas brutas (reportadas ao usuario), sem a correcao — a
	// correcao e usada so internamente para o calculo de d'/c.
	if maliciousTotal > 0 {
		stats.HitRate = float64(maliciousHits) / float64(maliciousTotal)
	}
	if legitimateTotal > 0 {
		stats.FalseAlarmRate = float64(falseAlarms) / float64(legitimateTotal)
	}

	hitRateAdj := (float64(maliciousHits) + 0.5) / (float64(maliciousTotal) + 1)
	falseAlarmRateAdj := (float64(falseAlarms) + 0.5) / (float64(legitimateTotal) + 1)

	zHit := probitZ(hitRateAdj)
	zFalseAlarm := probitZ(falseAlarmRateAdj)

	stats.DPrime = zHit - zFalseAlarm
	stats.Criterion = -0.5 * (zHit + zFalseAlarm)

	return stats
}

// probitZ converte uma probabilidade (0,1) no z-score correspondente
// da normal padrao (funcao quantil/probit), via math.Erfinv da
// biblioteca padrao do Go — nao exige dependencia externa de
// estatistica so para esta conversao.
func probitZ(p float64) float64 {
	return math.Sqrt2 * math.Erfinv(2*p-1)
}

func computeCuePerformance(outcomes []repository.CueAttemptOutcome) []domain.CuePerformance {
	type acc struct {
		labelPt  string
		answered int
		correct  int
	}
	byCue := make(map[string]*acc)
	order := make([]string, 0)

	for _, o := range outcomes {
		entry, exists := byCue[o.CueCode]
		if !exists {
			entry = &acc{labelPt: o.LabelPt}
			byCue[o.CueCode] = entry
			order = append(order, o.CueCode)
		}
		entry.answered++
		if o.IsCorrect {
			entry.correct++
		}
	}

	performance := make([]domain.CuePerformance, 0, len(order))
	for _, code := range order {
		entry := byCue[code]
		accuracy := 0.0
		if entry.answered > 0 {
			accuracy = float64(entry.correct) / float64(entry.answered)
		}
		performance = append(performance, domain.CuePerformance{
			CueCode:  code,
			LabelPt:  entry.labelPt,
			Answered: entry.answered,
			Correct:  entry.correct,
			Accuracy: accuracy,
		})
	}
	return performance
}
