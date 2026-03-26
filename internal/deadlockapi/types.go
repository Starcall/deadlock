package deadlockapi

// MatchMetadataBulk is the response from GET /v1/matches/metadata (bulk endpoint).
// Uses string enums for match_mode, winning_team, etc.
type MatchMetadataBulk struct {
	MatchID       int64  `json:"match_id"`
	StartTime     string `json:"start_time"` // "2025-02-01 18:28:43"
	DurationS     int    `json:"duration_s"`
	MatchMode     string `json:"match_mode"`     // "Ranked", "Unranked", etc.
	GameMode      string `json:"game_mode"`      // "Normal"
	MatchOutcome  string `json:"match_outcome"`  // "TeamWin"
	WinningTeam   string `json:"winning_team"`   // "Team0", "Team1"
	AvgBadgeTeam0 *int   `json:"average_badge_team0"`
	AvgBadgeTeam1 *int   `json:"average_badge_team1"`
}

// MatchDetailResponse is the top-level wrapper from GET /v1/matches/{id}/metadata.
type MatchDetailResponse struct {
	MatchInfo MatchInfo `json:"match_info"`
}

// MatchInfo is the full match data from the individual match endpoint.
// Uses integer enums for match_mode, winning_team, etc.
type MatchInfo struct {
	MatchID       int64  `json:"match_id"`
	StartTime     int64  `json:"start_time"` // Unix timestamp
	DurationS     int    `json:"duration_s"`
	MatchOutcome  int    `json:"match_outcome"`  // 0 = TeamWin
	WinningTeam   int    `json:"winning_team"`   // 0 or 1
	MatchMode     int    `json:"match_mode"`     // 1 = Unranked
	GameMode      int    `json:"game_mode"`      // 1 = Normal
	AvgBadgeTeam0 *int   `json:"average_badge_team0"`
	AvgBadgeTeam1 *int   `json:"average_badge_team1"`
	Players       []PlayerInfo    `json:"players"`
	Objectives    []ObjectiveInfo `json:"objectives"`
	MidBoss       []MidBossInfo   `json:"mid_boss"`
}

// MatchModeString returns the string representation of the match mode.
func (m *MatchInfo) MatchModeString() string {
	switch m.MatchMode {
	case 0:
		return "Unranked"
	case 1:
		return "Unranked"
	case 2:
		return "PrivateLobby"
	case 3:
		return "CoopBot"
	case 4:
		return "Ranked"
	default:
		return "Unknown"
	}
}

// PlayerInfo is a player in a match detail response.
type PlayerInfo struct {
	AccountID  int    `json:"account_id"`
	PlayerSlot int    `json:"player_slot"`
	Team       int    `json:"team"`
	HeroID     int    `json:"hero_id"`
	Level      int    `json:"level"`
	Kills      int    `json:"kills"`
	Deaths     int    `json:"deaths"`
	Assists    int    `json:"assists"`
	NetWorth   int    `json:"net_worth"`
	LastHits   int    `json:"last_hits"`
	Denies     int    `json:"denies"`
	Items      []ItemPurchaseInfo `json:"items"`
	Stats      []StatSnapshotInfo `json:"stats"`
}

// ItemPurchaseInfo is an item purchase event in the match detail response.
type ItemPurchaseInfo struct {
	GameTimeS      int   `json:"game_time_s"`
	ItemID         int64 `json:"item_id"`
	UpgradeID      int64 `json:"upgrade_id"`
	SoldTimeS      int   `json:"sold_time_s"`
	Flags          int   `json:"flags"`
	ImbuedAbilityID int64 `json:"imbued_ability_id"`
}

// StatSnapshotInfo is a periodic stat snapshot from the match detail response.
type StatSnapshotInfo struct {
	TimeStampS    int `json:"time_stamp_s"`
	Level         int `json:"level"`
	NetWorth      int `json:"net_worth"`
	Kills         int `json:"kills"`
	Deaths        int `json:"deaths"`
	Assists       int `json:"assists"`
	CreepKills    int `json:"creep_kills"`
	NeutralKills  int `json:"neutral_kills"`
	Denies        int `json:"denies"`
	CreepDamage   int `json:"creep_damage"`
	PlayerDamage  int `json:"player_damage"`
	NeutralDamage int `json:"neutral_damage"`
	BossDamage    int `json:"boss_damage"`
	MaxHealth     int `json:"max_health"`
}

// ObjectiveInfo is an objective in the match detail response.
type ObjectiveInfo struct {
	TeamObjectiveID   int `json:"team_objective_id"`
	Team              int `json:"team"`
	DestroyedTimeS    int `json:"destroyed_time_s"`
	FirstDamageTimeS  int `json:"first_damage_time_s"`
}

// MidBossInfo is a mid boss kill in the match detail response.
type MidBossInfo struct {
	TeamKilled     int `json:"team_killed"`
	TeamClaimed    int `json:"team_claimed"`
	DestroyedTimeS int `json:"destroyed_time_s"`
}
