package builds

import (
	"fmt"
	"sort"
	"strings"

	"github.com/name/deadlock/internal/domain"
)

const (
	TopN             = 10
	OverlapThreshold = 0.80
	MinFinalItems    = 4
)

// PlayerBuild holds a single player's final item set and match outcome.
type PlayerBuild struct {
	HeroID  int
	ItemIDs []int64 // sorted
	Won     bool
}

// CollectPlayerBuild extracts the final build from a player's items.
// Returns nil if the player has fewer than MinFinalItems.
func CollectPlayerBuild(heroID int, team int, winningTeam int, items []domain.ItemPurchase, playerSlot int, shopItemIDs map[int64]bool) *PlayerBuild {
	var finalItems []int64
	for _, it := range items {
		if it.PlayerSlot != playerSlot {
			continue
		}
		if it.SoldTimeS != 0 {
			continue
		}
		if !shopItemIDs[it.ItemID] {
			continue
		}
		finalItems = append(finalItems, it.ItemID)
	}

	if len(finalItems) < MinFinalItems {
		return nil
	}

	sort.Slice(finalItems, func(i, j int) bool { return finalItems[i] < finalItems[j] })

	return &PlayerBuild{
		HeroID:  heroID,
		ItemIDs: finalItems,
		Won:     team == winningTeam,
	}
}

// itemSetKey returns a string key for a sorted item set.
func itemSetKey(items []int64) string {
	parts := make([]string, len(items))
	for i, id := range items {
		parts[i] = fmt.Sprintf("%d", id)
	}
	return strings.Join(parts, ",")
}

// overlap computes |intersection(player, template)| / |template|.
func overlap(playerItems, templateItems []int64) float64 {
	if len(templateItems) == 0 {
		return 0
	}
	templateSet := make(map[int64]bool, len(templateItems))
	for _, id := range templateItems {
		templateSet[id] = true
	}
	count := 0
	for _, id := range playerItems {
		if templateSet[id] {
			count++
		}
	}
	return float64(count) / float64(len(templateItems))
}

type buildFreq struct {
	key     string
	items   []int64
	count   int
}

// ComputeBuildWinRates computes top builds, classifies players, and returns templates + coverage.
func ComputeBuildWinRates(allBuilds []PlayerBuild) ([]domain.BuildTemplate, []domain.HeroBuildCoverage) {
	// Group builds by hero
	heroBuilds := make(map[int][]PlayerBuild)
	for _, b := range allBuilds {
		heroBuilds[b.HeroID] = append(heroBuilds[b.HeroID], b)
	}

	var allTemplates []domain.BuildTemplate
	var allCoverage []domain.HeroBuildCoverage

	for heroID, builds := range heroBuilds {
		totalPlayers := len(builds)

		// Count exact build frequencies
		freqMap := make(map[string]*buildFreq)
		for _, b := range builds {
			key := itemSetKey(b.ItemIDs)
			if f, ok := freqMap[key]; ok {
				f.count++
			} else {
				freqMap[key] = &buildFreq{key: key, items: b.ItemIDs, count: 1}
			}
		}

		// Sort by frequency descending, take top N
		freqs := make([]*buildFreq, 0, len(freqMap))
		for _, f := range freqMap {
			freqs = append(freqs, f)
		}
		sort.Slice(freqs, func(i, j int) bool { return freqs[i].count > freqs[j].count })
		if len(freqs) > TopN {
			freqs = freqs[:TopN]
		}

		// Classify each player against templates
		classifiedCount := 0
		type templateStats struct {
			exactCount int
			fuzzyWins  int
			fuzzyTotal int
		}
		stats := make([]templateStats, len(freqs))

		for _, b := range builds {
			playerKey := itemSetKey(b.ItemIDs)

			bestIdx := -1
			bestOverlap := 0.0

			for ti, tmpl := range freqs {
				if playerKey == tmpl.key {
					// Exact match
					bestIdx = ti
					bestOverlap = 1.0
					break
				}
				ov := overlap(b.ItemIDs, tmpl.items)
				if ov >= OverlapThreshold && ov > bestOverlap {
					bestOverlap = ov
					bestIdx = ti
				}
			}

			if bestIdx >= 0 {
				classifiedCount++
				if playerKey == freqs[bestIdx].key {
					stats[bestIdx].exactCount++
				}
				stats[bestIdx].fuzzyTotal++
				if b.Won {
					stats[bestIdx].fuzzyWins++
				}
			}
		}

		// Build results
		for i, tmpl := range freqs {
			winRate := 0.0
			if stats[i].fuzzyTotal > 0 {
				winRate = float64(stats[i].fuzzyWins) / float64(stats[i].fuzzyTotal)
			}
			losses := stats[i].fuzzyTotal - stats[i].fuzzyWins

			allTemplates = append(allTemplates, domain.BuildTemplate{
				HeroID:           heroID,
				BuildRank:        i + 1,
				ItemIDs:          tmpl.key,
				ExactCount:       stats[i].exactCount,
				FuzzyCount:       stats[i].fuzzyTotal,
				Wins:             stats[i].fuzzyWins,
				Losses:           losses,
				WinRate:          winRate,
				TotalHeroPlayers: totalPlayers,
			})
		}

		coverage := 0.0
		if totalPlayers > 0 {
			coverage = float64(classifiedCount) / float64(totalPlayers)
		}
		allCoverage = append(allCoverage, domain.HeroBuildCoverage{
			HeroID:          heroID,
			TotalPlayers:    totalPlayers,
			ClassifiedCount: classifiedCount,
			Coverage:        coverage,
		})
	}

	return allTemplates, allCoverage
}
