package wpa

// ContextDef defines a filtering context for WPA aggregation.
type ContextDef struct {
	Key         string
	Description string
	MatchFilter func(avgBadge float64) bool
	PhaseFilter func(gameTimeS int) bool
}

type rankDef struct {
	key         string
	description string
	filter      func(float64) bool
}

type timeDef struct {
	key         string
	description string
	filter      func(int) bool
}

var ranks = []rankDef{
	{"all", "All Ranks", func(float64) bool { return true }},
	{"rank:low", "Low Rank (0-30)", func(b float64) bool { return b <= 30 }},
	{"rank:mid", "Mid Rank (31-70)", func(b float64) bool { return b > 30 && b <= 70 }},
	{"rank:high", "High Rank (71+)", func(b float64) bool { return b > 70 }},
}

var times = []timeDef{
	{"all", "All Time", func(int) bool { return true }},
	{"before:5m", "Before 5 min", func(t int) bool { return t <= 300 }},
	{"before:8m", "Before 8 min", func(t int) bool { return t <= 480 }},
	{"before:10m", "Before 10 min", func(t int) bool { return t <= 600 }},
	{"before:15m", "Before 15 min", func(t int) bool { return t <= 900 }},
	{"before:20m", "Before 20 min", func(t int) bool { return t <= 1200 }},
	{"phase:early", "Early (0-10m)", func(t int) bool { return t <= 600 }},
	{"phase:mid", "Mid (10-25m)", func(t int) bool { return t > 600 && t <= 1500 }},
	{"phase:late", "Late (25m+)", func(t int) bool { return t > 1500 }},
}

// AllContexts returns all rank x time context combinations.
func AllContexts() []ContextDef {
	var out []ContextDef
	for _, r := range ranks {
		for _, t := range times {
			key := r.key + "|" + t.key
			desc := r.description + " / " + t.description

			// Capture loop vars
			rf := r.filter
			tf := t.filter

			out = append(out, ContextDef{
				Key:         key,
				Description: desc,
				MatchFilter: rf,
				PhaseFilter: tf,
			})
		}
	}
	return out
}
