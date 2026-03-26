package domain

import "math"

// GameState captures the full game state at a point in time, used as input
// to the win probability model. Features are team-relative: team0 is the
// "perspective" team, team1 is the opponent.
type GameState struct {
	// Time context
	GameTimeS    int     // seconds into the match
	GameTimeFrac float64 // game_time / max_game_time (normalized)

	// Team-level differentials (team0 - team1)
	NetWorthDiff     int
	KillsDiff        int
	DeathsDiff       int
	AssistsDiff      int
	PlayerDamageDiff int

	// Per-player stats (6 per team, 12 total), normalized
	PlayerNetWorth     [12]float64
	PlayerKills        [12]float64
	PlayerDeaths       [12]float64
	PlayerAssists      [12]float64
	PlayerLevel        [12]float64

	// Rank context
	AvgBadge float64
}

// WPAResult stores the pre-computed WPA for a hero/item/context combination.
type WPAResult struct {
	HeroID       int     `json:"hero_id"`
	ItemID       int64   `json:"item_id"`
	ContextKey   string  `json:"context_key"`
	MeanDeltaW   float64 `json:"mean_delta_w"`
	MeanInitialW float64 `json:"mean_initial_w"`
	WinRate      float64 `json:"win_rate"`
	SampleSize   int     `json:"sample_size"`
	StdDeltaW    float64 `json:"std_delta_w"`
	PValue       float64 `json:"p_value"`
}

// ComputePValue sets the two-tailed p-value for H0: mean ΔW = 0.
func (r *WPAResult) ComputePValue() {
	if r.SampleSize < 2 || r.StdDeltaW == 0 {
		r.PValue = 1.0
		return
	}
	se := r.StdDeltaW / math.Sqrt(float64(r.SampleSize))
	t := math.Abs(r.MeanDeltaW / se)
	r.PValue = 2 * normalSurvival(t)
}

// normalSurvival returns P(Z > x) for standard normal using
// Abramowitz & Stegun approximation 26.2.17.
func normalSurvival(x float64) float64 {
	const (
		b1 = 0.319381530
		b2 = -0.356563782
		b3 = 1.781477937
		b4 = -1.821255978
		b5 = 1.330274429
		p  = 0.2316419
	)
	t := 1.0 / (1.0 + p*x)
	pdf := math.Exp(-x*x/2) / math.Sqrt(2*math.Pi)
	return pdf * (b1*t + b2*t*t + b3*t*t*t + b4*t*t*t*t + b5*t*t*t*t*t)
}

// ModelMetadata stores information about a trained win probability model.
type ModelMetadata struct {
	ID         int     `json:"id"`
	TrainedAt  int64   `json:"trained_at"`
	Accuracy   float64 `json:"accuracy"`
	ECE        float64 `json:"ece"` // expected calibration error
	NumMatches int     `json:"num_matches"`
	Weights    []byte  `json:"-"` // serialized model weights
	IsActive   bool    `json:"is_active"`
}
