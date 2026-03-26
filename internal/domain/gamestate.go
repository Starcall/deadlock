package domain

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
	HeroID      int     `json:"hero_id"`
	ItemID      int64   `json:"item_id"`
	ContextKey  string  `json:"context_key"`  // e.g. "all", "rank:0-30", "phase:early"
	MeanDeltaW  float64 `json:"mean_delta_w"` // average ΔW
	MeanInitialW float64 `json:"mean_initial_w"` // average W before purchase
	WinRate     float64 `json:"win_rate"`     // raw outcome rate
	SampleSize  int     `json:"sample_size"`  // number of purchase events
	StdDeltaW   float64 `json:"std_delta_w"`  // standard deviation of ΔW
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
