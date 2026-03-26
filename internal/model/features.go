package model

import (
	"github.com/name/deadlock/internal/domain"
)

// FeatureCount is the number of features in the model's input vector.
// 5 team-level diffs + 12*5 per-player stats + 1 game time + 1 avg badge = 67
const FeatureCount = 67

// MaxGameTime is the normalization constant for game time (seconds).
const MaxGameTime = 4800.0 // 80 minutes

// MaxNetWorth is the normalization constant for net worth.
const MaxNetWorth = 100000.0

// MaxKills is the normalization constant for kills.
const MaxKills = 40.0

// MaxDeaths is the normalization constant for deaths.
const MaxDeaths = 20.0

// MaxAssists is the normalization constant for assists.
const MaxAssists = 40.0

// MaxLevel is the normalization constant for hero level.
const MaxLevel = 50.0

// MaxBadge is the normalization constant for average badge.
const MaxBadge = 116.0

// MaxPlayerDamage is the normalization constant for player damage.
const MaxPlayerDamage = 200000.0

// ExtractFeatures builds a feature vector from a game state.
// The perspective is team 0. Label = 1 if team 0 won, 0 otherwise.
func ExtractFeatures(state *domain.GameState) []float64 {
	f := make([]float64, FeatureCount)
	idx := 0

	// Team-level differentials (normalized)
	f[idx] = float64(state.NetWorthDiff) / MaxNetWorth
	idx++
	f[idx] = float64(state.KillsDiff) / MaxKills
	idx++
	f[idx] = float64(state.DeathsDiff) / MaxDeaths
	idx++
	f[idx] = float64(state.AssistsDiff) / MaxAssists
	idx++
	f[idx] = float64(state.PlayerDamageDiff) / MaxPlayerDamage
	idx++

	// Per-player stats (12 players × 5 features)
	for i := range 12 {
		f[idx] = state.PlayerNetWorth[i] / MaxNetWorth
		idx++
		f[idx] = state.PlayerKills[i] / MaxKills
		idx++
		f[idx] = state.PlayerDeaths[i] / MaxDeaths
		idx++
		f[idx] = state.PlayerAssists[i] / MaxAssists
		idx++
		f[idx] = state.PlayerLevel[i] / MaxLevel
		idx++
	}

	// Game time (normalized)
	f[idx] = state.GameTimeFrac
	idx++

	// Average badge (normalized)
	f[idx] = state.AvgBadge / MaxBadge

	return f
}

// BuildGameState constructs a GameState from match snapshots at a given time.
// snapsBySlot maps player_slot to their time-series snapshots.
// Returns nil if insufficient data.
func BuildGameState(
	snapsBySlot map[int][]domain.StatSnapshot,
	players []domain.MatchPlayer,
	gameTimeS int,
	durationS int,
	avgBadge float64,
) *domain.GameState {
	state := &domain.GameState{
		GameTimeS:    gameTimeS,
		GameTimeFrac: float64(gameTimeS) / MaxGameTime,
		AvgBadge:     avgBadge,
	}

	if state.GameTimeFrac > 1.0 {
		state.GameTimeFrac = 1.0
	}

	// Build player lookup
	playerBySlot := make(map[int]*domain.MatchPlayer, len(players))
	for i := range players {
		playerBySlot[players[i].PlayerSlot] = &players[i]
	}

	// For each player, find the closest snapshot <= gameTimeS
	var team0NW, team1NW int
	var team0K, team1K, team0D, team1D, team0A, team1A int
	var team0Dmg, team1Dmg int

	for slot := range 12 {
		snaps, ok := snapsBySlot[slot]
		if !ok || len(snaps) == 0 {
			continue
		}

		// Find closest snapshot at or before gameTimeS
		var snap *domain.StatSnapshot
		for i := len(snaps) - 1; i >= 0; i-- {
			if snaps[i].GameTimeS <= gameTimeS {
				snap = &snaps[i]
				break
			}
		}
		if snap == nil {
			snap = &snaps[0]
		}

		player := playerBySlot[slot]
		if player == nil {
			continue
		}

		// Fill per-player features
		state.PlayerNetWorth[slot] = float64(snap.NetWorth)
		state.PlayerKills[slot] = float64(snap.Kills)
		state.PlayerDeaths[slot] = float64(snap.Deaths)
		state.PlayerAssists[slot] = float64(snap.Assists)
		state.PlayerLevel[slot] = float64(snap.HeroLevel)

		if player.Team == 0 {
			team0NW += snap.NetWorth
			team0K += snap.Kills
			team0D += snap.Deaths
			team0A += snap.Assists
			team0Dmg += snap.PlayerDamage
		} else {
			team1NW += snap.NetWorth
			team1K += snap.Kills
			team1D += snap.Deaths
			team1A += snap.Assists
			team1Dmg += snap.PlayerDamage
		}
	}

	state.NetWorthDiff = team0NW - team1NW
	state.KillsDiff = team0K - team1K
	state.DeathsDiff = team0D - team1D
	state.AssistsDiff = team0A - team1A
	state.PlayerDamageDiff = team0Dmg - team1Dmg

	return state
}
