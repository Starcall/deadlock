package wpa

// ContextDef defines a filtering context for WPA aggregation.
type ContextDef struct {
	Key         string
	Description string
	MatchFilter func(avgBadge float64) bool
	PhaseFilter func(gameTimeS int) bool
}

// AllContexts returns the standard set of WPA contexts.
func AllContexts() []ContextDef {
	return []ContextDef{
		{
			Key:         "all",
			Description: "All matches",
			MatchFilter: func(float64) bool { return true },
			PhaseFilter: func(int) bool { return true },
		},
		// Rank brackets (badge ranges)
		{
			Key:         "rank:low",
			Description: "Low rank (badge 0-30)",
			MatchFilter: func(b float64) bool { return b <= 30 },
			PhaseFilter: func(int) bool { return true },
		},
		{
			Key:         "rank:mid",
			Description: "Mid rank (badge 31-70)",
			MatchFilter: func(b float64) bool { return b > 30 && b <= 70 },
			PhaseFilter: func(int) bool { return true },
		},
		{
			Key:         "rank:high",
			Description: "High rank (badge 71+)",
			MatchFilter: func(b float64) bool { return b > 70 },
			PhaseFilter: func(int) bool { return true },
		},
		// Game phases
		{
			Key:         "phase:early",
			Description: "Early game (0-10 min)",
			MatchFilter: func(float64) bool { return true },
			PhaseFilter: func(t int) bool { return t <= 600 },
		},
		{
			Key:         "phase:mid",
			Description: "Mid game (10-25 min)",
			MatchFilter: func(float64) bool { return true },
			PhaseFilter: func(t int) bool { return t > 600 && t <= 1500 },
		},
		{
			Key:         "phase:late",
			Description: "Late game (25+ min)",
			MatchFilter: func(float64) bool { return true },
			PhaseFilter: func(t int) bool { return t > 1500 },
		},
		// Cumulative time filters (items bought before minute X)
		{
			Key:         "before:5m",
			Description: "Items bought before 5 min",
			MatchFilter: func(float64) bool { return true },
			PhaseFilter: func(t int) bool { return t <= 300 },
		},
		{
			Key:         "before:8m",
			Description: "Items bought before 8 min",
			MatchFilter: func(float64) bool { return true },
			PhaseFilter: func(t int) bool { return t <= 480 },
		},
		{
			Key:         "before:10m",
			Description: "Items bought before 10 min",
			MatchFilter: func(float64) bool { return true },
			PhaseFilter: func(t int) bool { return t <= 600 },
		},
		{
			Key:         "before:15m",
			Description: "Items bought before 15 min",
			MatchFilter: func(float64) bool { return true },
			PhaseFilter: func(t int) bool { return t <= 900 },
		},
		{
			Key:         "before:20m",
			Description: "Items bought before 20 min",
			MatchFilter: func(float64) bool { return true },
			PhaseFilter: func(t int) bool { return t <= 1200 },
		},
	}
}
