package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/name/deadlock/internal/domain"
)

// InsertMatch inserts a match and its players, items, and snapshots in a transaction.
func (db *DB) InsertMatch(ctx context.Context, match *domain.Match, players []domain.MatchPlayer, items []domain.ItemPurchase, snapshots []domain.StatSnapshot) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert match
	_, err = tx.ExecContext(ctx,
		`INSERT OR IGNORE INTO matches (match_id, duration_s, winning_team, match_mode, avg_badge_team0, avg_badge_team1, start_time)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		match.MatchID, match.DurationS, match.WinningTeam, match.MatchMode,
		match.AvgBadgeTeam0, match.AvgBadgeTeam1, match.StartTime,
	)
	if err != nil {
		return fmt.Errorf("inserting match: %w", err)
	}

	// Insert players
	playerStmt, err := tx.PrepareContext(ctx,
		`INSERT OR IGNORE INTO match_players (match_id, player_slot, hero_id, team, kills, deaths, assists, net_worth, player_damage, creep_damage, hero_level)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing player statement: %w", err)
	}
	defer playerStmt.Close()

	for _, p := range players {
		if _, err := playerStmt.ExecContext(ctx, p.MatchID, p.PlayerSlot, p.HeroID, p.Team, p.Kills, p.Deaths, p.Assists, p.NetWorth, p.PlayerDamage, p.CreepDamage, p.HeroLevel); err != nil {
			return fmt.Errorf("inserting player: %w", err)
		}
	}

	// Insert items
	itemStmt, err := tx.PrepareContext(ctx,
		`INSERT OR IGNORE INTO match_player_items (match_id, player_slot, item_id, game_time_s, sold_time_s)
		 VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing item statement: %w", err)
	}
	defer itemStmt.Close()

	for _, item := range items {
		if _, err := itemStmt.ExecContext(ctx, item.MatchID, item.PlayerSlot, item.ItemID, item.GameTimeS, item.SoldTimeS); err != nil {
			return fmt.Errorf("inserting item: %w", err)
		}
	}

	// Insert snapshots
	snapStmt, err := tx.PrepareContext(ctx,
		`INSERT OR IGNORE INTO match_snapshots (match_id, player_slot, game_time_s, net_worth, kills, deaths, assists, player_damage, creep_damage, hero_level)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing snapshot statement: %w", err)
	}
	defer snapStmt.Close()

	for _, s := range snapshots {
		if _, err := snapStmt.ExecContext(ctx, s.MatchID, s.PlayerSlot, s.GameTimeS, s.NetWorth, s.Kills, s.Deaths, s.Assists, s.PlayerDamage, s.CreepDamage, s.HeroLevel); err != nil {
			return fmt.Errorf("inserting snapshot: %w", err)
		}
	}

	return tx.Commit()
}

// MatchExists checks if a match is already stored.
func (db *DB) MatchExists(ctx context.Context, matchID int64) (bool, error) {
	var exists bool
	err := db.conn.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM matches WHERE match_id = ?)", matchID,
	).Scan(&exists)
	return exists, err
}

// CountMatches returns the total number of stored matches.
func (db *DB) CountMatches(ctx context.Context) (int, error) {
	var count int
	err := db.conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM matches").Scan(&count)
	return count, err
}

// GetMaxMatchID returns the highest stored match ID, or 0 if none.
func (db *DB) GetMaxMatchID(ctx context.Context) (int64, error) {
	var maxID sql.NullInt64
	err := db.conn.QueryRowContext(ctx, "SELECT MAX(match_id) FROM matches").Scan(&maxID)
	if err != nil {
		return 0, err
	}
	if !maxID.Valid {
		return 0, nil
	}
	return maxID.Int64, nil
}

// InsertHero inserts or replaces a hero.
func (db *DB) InsertHero(ctx context.Context, h domain.Hero) error {
	_, err := db.conn.ExecContext(ctx,
		`INSERT OR REPLACE INTO heroes (id, name, class_name, image_url) VALUES (?, ?, ?, ?)`,
		h.ID, h.Name, h.ClassName, h.ImageURL,
	)
	return err
}

// InsertItem inserts or replaces an item.
func (db *DB) InsertItem(ctx context.Context, item domain.Item) error {
	_, err := db.conn.ExecContext(ctx,
		`INSERT OR REPLACE INTO items (id, name, class_name, category, tier, cost, image_url) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		item.ID, item.Name, item.ClassName, item.ItemSlotType, item.Tier, item.Cost, item.ImageURL,
	)
	return err
}

// GetAllHeroes returns all heroes.
func (db *DB) GetAllHeroes(ctx context.Context) ([]domain.Hero, error) {
	rows, err := db.conn.QueryContext(ctx, "SELECT id, name, class_name, image_url FROM heroes ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	heroes := make([]domain.Hero, 0)
	for rows.Next() {
		var h domain.Hero
		var imageURL sql.NullString
		if err := rows.Scan(&h.ID, &h.Name, &h.ClassName, &imageURL); err != nil {
			return nil, err
		}
		h.ImageURL = imageURL.String
		heroes = append(heroes, h)
	}
	return heroes, rows.Err()
}

// GetAllItems returns all items.
func (db *DB) GetAllItems(ctx context.Context) ([]domain.Item, error) {
	rows, err := db.conn.QueryContext(ctx, "SELECT id, name, class_name, category, tier, cost, image_url FROM items ORDER BY tier, cost, name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Item, 0)
	for rows.Next() {
		var item domain.Item
		var category, imageURL sql.NullString
		var tier, cost sql.NullInt64
		if err := rows.Scan(&item.ID, &item.Name, &item.ClassName, &category, &tier, &cost, &imageURL); err != nil {
			return nil, err
		}
		item.ItemSlotType = category.String
		item.Tier = int(tier.Int64)
		item.Cost = int(cost.Int64)
		item.ImageURL = imageURL.String
		items = append(items, item)
	}
	return items, rows.Err()
}
