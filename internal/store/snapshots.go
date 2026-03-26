package store

import (
	"context"
	"fmt"

	"github.com/name/deadlock/internal/domain"
)

// GetMatchSnapshots returns all stat snapshots for a match, ordered by time.
func (db *DB) GetMatchSnapshots(ctx context.Context, matchID int64) ([]domain.StatSnapshot, error) {
	rows, err := db.conn.QueryContext(ctx,
		`SELECT match_id, player_slot, game_time_s, net_worth, kills, deaths, assists, player_damage, creep_damage, hero_level
		 FROM match_snapshots WHERE match_id = ? ORDER BY game_time_s, player_slot`, matchID)
	if err != nil {
		return nil, fmt.Errorf("querying snapshots: %w", err)
	}
	defer rows.Close()

	var snaps []domain.StatSnapshot
	for rows.Next() {
		var s domain.StatSnapshot
		if err := rows.Scan(&s.MatchID, &s.PlayerSlot, &s.GameTimeS, &s.NetWorth, &s.Kills, &s.Deaths, &s.Assists, &s.PlayerDamage, &s.CreepDamage, &s.HeroLevel); err != nil {
			return nil, fmt.Errorf("scanning snapshot: %w", err)
		}
		snaps = append(snaps, s)
	}
	return snaps, rows.Err()
}

// GetMatchPlayers returns all players for a match.
func (db *DB) GetMatchPlayers(ctx context.Context, matchID int64) ([]domain.MatchPlayer, error) {
	rows, err := db.conn.QueryContext(ctx,
		`SELECT match_id, player_slot, hero_id, team, kills, deaths, assists, net_worth, player_damage, creep_damage, hero_level
		 FROM match_players WHERE match_id = ? ORDER BY player_slot`, matchID)
	if err != nil {
		return nil, fmt.Errorf("querying players: %w", err)
	}
	defer rows.Close()

	var players []domain.MatchPlayer
	for rows.Next() {
		var p domain.MatchPlayer
		if err := rows.Scan(&p.MatchID, &p.PlayerSlot, &p.HeroID, &p.Team, &p.Kills, &p.Deaths, &p.Assists, &p.NetWorth, &p.PlayerDamage, &p.CreepDamage, &p.HeroLevel); err != nil {
			return nil, fmt.Errorf("scanning player: %w", err)
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

// GetMatchItems returns all item purchases for a match.
func (db *DB) GetMatchItems(ctx context.Context, matchID int64) ([]domain.ItemPurchase, error) {
	rows, err := db.conn.QueryContext(ctx,
		`SELECT match_id, player_slot, item_id, game_time_s, sold_time_s
		 FROM match_player_items WHERE match_id = ? ORDER BY game_time_s`, matchID)
	if err != nil {
		return nil, fmt.Errorf("querying items: %w", err)
	}
	defer rows.Close()

	var items []domain.ItemPurchase
	for rows.Next() {
		var item domain.ItemPurchase
		if err := rows.Scan(&item.MatchID, &item.PlayerSlot, &item.ItemID, &item.GameTimeS, &item.SoldTimeS); err != nil {
			return nil, fmt.Errorf("scanning item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// GetAllMatchIDs returns all stored match IDs.
func (db *DB) GetAllMatchIDs(ctx context.Context) ([]int64, error) {
	rows, err := db.conn.QueryContext(ctx, "SELECT match_id FROM matches ORDER BY match_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetMatchIDsSince returns match IDs with start_time >= the given unix timestamp.
func (db *DB) GetMatchIDsSince(ctx context.Context, sinceUnix int64) ([]int64, error) {
	rows, err := db.conn.QueryContext(ctx,
		"SELECT match_id FROM matches WHERE start_time >= ? ORDER BY match_id", sinceUnix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetMatch returns a single match by ID.
func (db *DB) GetMatch(ctx context.Context, matchID int64) (*domain.Match, error) {
	var m domain.Match
	err := db.conn.QueryRowContext(ctx,
		`SELECT match_id, duration_s, winning_team, match_mode, avg_badge_team0, avg_badge_team1, start_time
		 FROM matches WHERE match_id = ?`, matchID,
	).Scan(&m.MatchID, &m.DurationS, &m.WinningTeam, &m.MatchMode, &m.AvgBadgeTeam0, &m.AvgBadgeTeam1, &m.StartTime)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
