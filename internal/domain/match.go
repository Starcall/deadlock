package domain

// Match represents a Deadlock match.
type Match struct {
	MatchID       int64  `json:"match_id"`
	DurationS     int    `json:"duration_s"`
	WinningTeam   int    `json:"winning_team"` // 0 or 1
	MatchMode     string `json:"match_mode"`
	AvgBadgeTeam0 *int   `json:"average_badge_team0"`
	AvgBadgeTeam1 *int   `json:"average_badge_team1"`
	StartTime     int64  `json:"start_time"`
	Players       []MatchPlayer `json:"-"`
}

// MatchPlayer represents a player's end-of-match stats.
type MatchPlayer struct {
	MatchID      int64  `json:"match_id"`
	PlayerSlot   int    `json:"player_slot"` // 0-11
	HeroID       int    `json:"hero_id"`
	Team         int    `json:"team"` // 0 or 1
	Kills        int    `json:"kills"`
	Deaths       int    `json:"deaths"`
	Assists      int    `json:"assists"`
	NetWorth     int    `json:"net_worth"`
	PlayerDamage int    `json:"player_damage"`
	CreepDamage  int    `json:"creep_damage"`
	HeroLevel    int    `json:"level"`
	Items        []ItemPurchase `json:"-"`
	Snapshots    []StatSnapshot `json:"-"`
}

// ItemPurchase represents an item purchase event within a match.
type ItemPurchase struct {
	MatchID    int64 `json:"match_id"`
	PlayerSlot int   `json:"player_slot"`
	ItemID     int64 `json:"item_id"`
	GameTimeS  int   `json:"game_time_s"`
	SoldTimeS  int   `json:"sold_time_s"`
}

// StatSnapshot represents a periodic stat snapshot for a player.
type StatSnapshot struct {
	MatchID      int64 `json:"match_id"`
	PlayerSlot   int   `json:"player_slot"`
	GameTimeS    int   `json:"time_stamp_s"`
	NetWorth     int   `json:"net_worth"`
	Kills        int   `json:"kills"`
	Deaths       int   `json:"deaths"`
	Assists      int   `json:"assists"`
	PlayerDamage int   `json:"player_damage"`
	CreepDamage  int   `json:"creep_damage"`
	HeroLevel    int   `json:"level"`
}
