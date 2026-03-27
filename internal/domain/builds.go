package domain

// BuildTemplate represents a common end-game item build for a hero.
type BuildTemplate struct {
	HeroID           int     `json:"hero_id"`
	BuildRank        int     `json:"build_rank"`
	ItemIDs          string  `json:"item_ids"` // comma-separated sorted item IDs
	ExactCount       int     `json:"exact_count"`
	FuzzyCount       int     `json:"fuzzy_count"`
	Wins             int     `json:"wins"`
	Losses           int     `json:"losses"`
	WinRate          float64 `json:"win_rate"`
	TotalHeroPlayers int     `json:"total_hero_players"`
}

// HeroBuildCoverage represents build classification coverage for a hero.
type HeroBuildCoverage struct {
	HeroID          int     `json:"hero_id"`
	TotalPlayers    int     `json:"total_players"`
	ClassifiedCount int     `json:"classified_count"`
	Coverage        float64 `json:"coverage"`
}
