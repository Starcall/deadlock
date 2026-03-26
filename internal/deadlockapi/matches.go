package deadlockapi

import (
	"context"
	"fmt"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/name/deadlock/internal/domain"
)

// MatchQuery defines parameters for bulk match metadata queries.
type MatchQuery struct {
	MinMatchID      int64
	MaxMatchID      int64
	MinUnixTime     int64
	MaxUnixTime     int64
	MinDurationS    int
	MaxDurationS    int
	MinAvgBadge     int
	MaxAvgBadge     int
	GameMode        string // "normal", "street_brawl", etc.
	IncludePlayerInfo bool
}

// FetchMatchMetadata retrieves bulk match metadata with the given filters.
func (c *Client) FetchMatchMetadata(ctx context.Context, q MatchQuery) ([]MatchMetadataBulk, error) {
	u, _ := url.Parse(fmt.Sprintf("%s/v1/matches/metadata", baseURL))
	params := u.Query()

	if q.MinMatchID > 0 {
		params.Set("min_match_id", strconv.FormatInt(q.MinMatchID, 10))
	}
	if q.MaxMatchID > 0 {
		params.Set("max_match_id", strconv.FormatInt(q.MaxMatchID, 10))
	}
	if q.MinUnixTime > 0 {
		params.Set("min_unix_timestamp", strconv.FormatInt(q.MinUnixTime, 10))
	}
	if q.MaxUnixTime > 0 {
		params.Set("max_unix_timestamp", strconv.FormatInt(q.MaxUnixTime, 10))
	}
	if q.MinDurationS > 0 {
		params.Set("min_duration_s", strconv.Itoa(q.MinDurationS))
	}
	if q.MaxDurationS > 0 {
		params.Set("max_duration_s", strconv.Itoa(q.MaxDurationS))
	}
	if q.MinAvgBadge > 0 {
		params.Set("min_average_badge", strconv.Itoa(q.MinAvgBadge))
	}
	if q.MaxAvgBadge > 0 {
		params.Set("max_average_badge", strconv.Itoa(q.MaxAvgBadge))
	}
	if q.GameMode != "" {
		params.Set("game_mode", q.GameMode)
	}
	if q.IncludePlayerInfo {
		params.Set("include_player_info", "true")
	}

	u.RawQuery = params.Encode()

	body, err := c.get(ctx, u.String())
	if err != nil {
		// 404 means no matches found — return empty, not error
		if strings.Contains(err.Error(), "API returned 404") {
			return nil, nil
		}
		return nil, fmt.Errorf("fetching match metadata: %w", err)
	}
	var result []MatchMetadataBulk
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshalling match metadata: %w", err)
	}
	return result, nil
}

// FetchMatchDetail retrieves full match data for a single match.
func (c *Client) FetchMatchDetail(ctx context.Context, matchID int64) (*MatchInfo, error) {
	url := fmt.Sprintf("%s/v1/matches/%d/metadata", baseURL, matchID)

	var resp MatchDetailResponse
	if err := c.getJSON(ctx, url, &resp); err != nil {
		return nil, fmt.Errorf("fetching match %d: %w", matchID, err)
	}
	return &resp.MatchInfo, nil
}

// ConvertMatch converts API match info to domain types.
func ConvertMatch(info *MatchInfo) (*domain.Match, []domain.MatchPlayer, []domain.ItemPurchase, []domain.StatSnapshot) {
	match := &domain.Match{
		MatchID:       info.MatchID,
		DurationS:     info.DurationS,
		WinningTeam:   info.WinningTeam,
		MatchMode:     info.MatchModeString(),
		AvgBadgeTeam0: info.AvgBadgeTeam0,
		AvgBadgeTeam1: info.AvgBadgeTeam1,
		StartTime:     info.StartTime,
	}

	var players []domain.MatchPlayer
	var items []domain.ItemPurchase
	var snapshots []domain.StatSnapshot

	for _, p := range info.Players {
		player := domain.MatchPlayer{
			MatchID:      info.MatchID,
			PlayerSlot:   p.PlayerSlot,
			HeroID:       p.HeroID,
			Team:         p.Team,
			Kills:        p.Kills,
			Deaths:       p.Deaths,
			Assists:      p.Assists,
			NetWorth:     p.NetWorth,
			HeroLevel:    p.Level,
		}

		// Extract player_damage and creep_damage from last snapshot if available
		if len(p.Stats) > 0 {
			last := p.Stats[len(p.Stats)-1]
			player.PlayerDamage = last.PlayerDamage
			player.CreepDamage = last.CreepDamage
		}

		players = append(players, player)

		for _, item := range p.Items {
			items = append(items, domain.ItemPurchase{
				MatchID:    info.MatchID,
				PlayerSlot: p.PlayerSlot,
				ItemID:     item.ItemID,
				GameTimeS:  item.GameTimeS,
				SoldTimeS:  item.SoldTimeS,
			})
		}

		for _, s := range p.Stats {
			snapshots = append(snapshots, domain.StatSnapshot{
				MatchID:      info.MatchID,
				PlayerSlot:   p.PlayerSlot,
				GameTimeS:    s.TimeStampS,
				NetWorth:     s.NetWorth,
				Kills:        s.Kills,
				Deaths:       s.Deaths,
				Assists:      s.Assists,
				PlayerDamage: s.PlayerDamage,
				CreepDamage:  s.CreepDamage,
				HeroLevel:    s.Level,
			})
		}
	}

	return match, players, items, snapshots
}

// ParseBulkMatchIDs extracts match IDs from bulk metadata responses,
// filtering for ranked matches of sufficient duration.
// The API uses "Unranked" for all public matches; ranked matches are identified
// by having non-null average_badge fields.
func ParseBulkMatchIDs(metas []MatchMetadataBulk, minDurationS int) []int64 {
	var ids []int64
	for _, m := range metas {
		if m.DurationS < minDurationS {
			continue
		}
		// Ranked = has badges + normal game mode (excludes brawl, coop, private)
		if m.AvgBadgeTeam0 == nil || m.AvgBadgeTeam1 == nil {
			continue
		}
		if !strings.EqualFold(m.GameMode, "Normal") {
			continue
		}
		if strings.EqualFold(m.MatchMode, "PrivateLobby") || strings.EqualFold(m.MatchMode, "CoopBot") {
			continue
		}
		ids = append(ids, m.MatchID)
	}
	return ids
}

// ParseBulkStartTime parses the start_time string from bulk metadata.
func ParseBulkStartTime(s string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", s)
}
