package wpa

import (
	"github.com/name/deadlock/internal/domain"
	"github.com/name/deadlock/internal/model"
)

// PurchaseEvent holds the computed WPA for a single item purchase.
type PurchaseEvent struct {
	HeroID    int
	ItemID    int64
	Team      int
	GameTimeS int
	InitialW  float64 // w(state_before)
	DeltaW    float64 // w(state_after) - w(state_before)
	Won       bool    // did this player's team win?
	AvgBadge  float64
}

// ComputeMatchWPA calculates WPA for all item purchases in a match.
// Returns purchase events from the perspective of the purchasing player's team.
func ComputeMatchWPA(
	m *model.LogisticModel,
	match *domain.Match,
	players []domain.MatchPlayer,
	items []domain.ItemPurchase,
	snaps []domain.StatSnapshot,
) []PurchaseEvent {
	if len(players) != 12 || len(snaps) == 0 {
		return nil
	}

	snapsBySlot := groupSnapsBySlot(snaps)

	// Build player lookup
	playerBySlot := make(map[int]*domain.MatchPlayer, len(players))
	for i := range players {
		playerBySlot[players[i].PlayerSlot] = &players[i]
	}

	avgBadge := computeAvgBadge(match)

	var events []PurchaseEvent

	for _, item := range items {
		player := playerBySlot[item.PlayerSlot]
		if player == nil {
			continue
		}

		// Skip very early purchases (likely starting items with no meaningful state)
		if item.GameTimeS < 60 {
			continue
		}

		// Build game state BEFORE purchase (use snapshot just before purchase time)
		beforeTime := max(item.GameTimeS-1, 0)
		stateBefore := model.BuildGameState(snapsBySlot, players, beforeTime, match.DurationS, avgBadge)
		if stateBefore == nil {
			continue
		}

		// Build game state AFTER purchase — use +200s to capture the next
		// snapshot (snapshots arrive every 180-300s)
		afterTime := min(item.GameTimeS+200, match.DurationS)
		stateAfter := model.BuildGameState(snapsBySlot, players, afterTime, match.DurationS, avgBadge)
		if stateAfter == nil {
			continue
		}

		// Get win probabilities from perspective of team 0
		featuresBefore := model.ExtractFeatures(stateBefore)
		featuresAfter := model.ExtractFeatures(stateAfter)

		wBefore := m.PredictCalibrated(featuresBefore)
		wAfter := m.PredictCalibrated(featuresAfter)

		// Adjust perspective: if player is on team 1, flip probabilities
		if player.Team == 1 {
			wBefore = 1 - wBefore
			wAfter = 1 - wAfter
		}

		won := player.Team == match.WinningTeam

		events = append(events, PurchaseEvent{
			HeroID:    player.HeroID,
			ItemID:    item.ItemID,
			Team:      player.Team,
			GameTimeS: item.GameTimeS,
			InitialW:  wBefore,
			DeltaW:    wAfter - wBefore,
			Won:       won,
			AvgBadge:  avgBadge,
		})
	}

	return events
}

func groupSnapsBySlot(snaps []domain.StatSnapshot) map[int][]domain.StatSnapshot {
	m := make(map[int][]domain.StatSnapshot)
	for _, s := range snaps {
		m[s.PlayerSlot] = append(m[s.PlayerSlot], s)
	}
	return m
}

func computeAvgBadge(match *domain.Match) float64 {
	var total, count float64
	if match.AvgBadgeTeam0 != nil {
		total += float64(*match.AvgBadgeTeam0)
		count++
	}
	if match.AvgBadgeTeam1 != nil {
		total += float64(*match.AvgBadgeTeam1)
		count++
	}
	if count == 0 {
		return 50.0
	}
	return total / count
}
