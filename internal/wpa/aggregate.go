package wpa

import (
	"math"

	"github.com/name/deadlock/internal/domain"
)

// AggregateKey groups purchase events for aggregation.
type AggregateKey struct {
	HeroID     int
	ItemID     int64
	ContextKey string
}

// Accumulator collects stats for WPA aggregation.
type Accumulator struct {
	SumDeltaW   float64
	SumSqDeltaW float64
	SumInitialW float64
	Wins        int
	Count       int
}

// AggregateWPA groups purchase events by hero/item/context and computes mean WPA.
func AggregateWPA(events []PurchaseEvent, contexts []ContextDef) []domain.WPAResult {
	accums := make(map[AggregateKey]*Accumulator)

	for _, e := range events {
		for _, ctx := range contexts {
			if !ctx.MatchFilter(e.AvgBadge) {
				continue
			}
			if !ctx.PhaseFilter(e.GameTimeS) {
				continue
			}

			key := AggregateKey{
				HeroID:     e.HeroID,
				ItemID:     e.ItemID,
				ContextKey: ctx.Key,
			}

			acc, ok := accums[key]
			if !ok {
				acc = &Accumulator{}
				accums[key] = acc
			}

			acc.SumDeltaW += e.DeltaW
			acc.SumSqDeltaW += e.DeltaW * e.DeltaW
			acc.SumInitialW += e.InitialW
			acc.Count++
			if e.Won {
				acc.Wins++
			}
		}
	}

	results := make([]domain.WPAResult, 0, len(accums))
	for key, acc := range accums {
		if acc.Count == 0 {
			continue
		}
		n := float64(acc.Count)
		meanDW := acc.SumDeltaW / n
		meanIW := acc.SumInitialW / n
		winRate := float64(acc.Wins) / n

		// Standard deviation of delta_w
		variance := acc.SumSqDeltaW/n - meanDW*meanDW
		variance = max(variance, 0)
		stdDW := math.Sqrt(variance)

		results = append(results, domain.WPAResult{
			HeroID:       key.HeroID,
			ItemID:       key.ItemID,
			ContextKey:   key.ContextKey,
			MeanDeltaW:   meanDW,
			MeanInitialW: meanIW,
			WinRate:      winRate,
			SampleSize:   acc.Count,
			StdDeltaW:    stdDW,
		})
	}

	return results
}
