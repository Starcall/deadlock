package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB wraps a SQLite database connection.
type DB struct {
	conn *sql.DB
}

// New opens a SQLite database and runs migrations.
func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// SQLite only supports a single writer — limit pool to 1 connection
	// so concurrent goroutines queue at the Go level instead of hitting SQLITE_BUSY.
	conn.SetMaxOpenConns(1)

	// SQLite performance tuning
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=-64000", // 64MB cache
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	}
	for _, p := range pragmas {
		if _, err := conn.Exec(p); err != nil {
			conn.Close()
			return nil, fmt.Errorf("setting pragma %q: %w", p, err)
		}
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Conn returns the underlying sql.DB for advanced queries.
func (db *DB) Conn() *sql.DB {
	return db.conn
}

func (db *DB) migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS heroes (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			class_name TEXT NOT NULL,
			image_url TEXT
		)`,

		`CREATE TABLE IF NOT EXISTS items (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			class_name TEXT NOT NULL,
			category TEXT,
			tier INTEGER,
			cost INTEGER,
			image_url TEXT
		)`,

		`CREATE TABLE IF NOT EXISTS matches (
			match_id INTEGER PRIMARY KEY,
			duration_s INTEGER NOT NULL,
			winning_team INTEGER NOT NULL,
			match_mode TEXT NOT NULL,
			avg_badge_team0 INTEGER,
			avg_badge_team1 INTEGER,
			start_time INTEGER NOT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS match_players (
			match_id INTEGER NOT NULL,
			player_slot INTEGER NOT NULL,
			hero_id INTEGER NOT NULL,
			team INTEGER NOT NULL,
			kills INTEGER NOT NULL DEFAULT 0,
			deaths INTEGER NOT NULL DEFAULT 0,
			assists INTEGER NOT NULL DEFAULT 0,
			net_worth INTEGER NOT NULL DEFAULT 0,
			player_damage INTEGER NOT NULL DEFAULT 0,
			creep_damage INTEGER NOT NULL DEFAULT 0,
			hero_level INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (match_id, player_slot)
		)`,

		`CREATE TABLE IF NOT EXISTS match_player_items (
			match_id INTEGER NOT NULL,
			player_slot INTEGER NOT NULL,
			item_id INTEGER NOT NULL,
			game_time_s INTEGER NOT NULL,
			sold_time_s INTEGER NOT NULL DEFAULT 0
		)`,

		`CREATE TABLE IF NOT EXISTS match_snapshots (
			match_id INTEGER NOT NULL,
			player_slot INTEGER NOT NULL,
			game_time_s INTEGER NOT NULL,
			net_worth INTEGER NOT NULL DEFAULT 0,
			kills INTEGER NOT NULL DEFAULT 0,
			deaths INTEGER NOT NULL DEFAULT 0,
			assists INTEGER NOT NULL DEFAULT 0,
			player_damage INTEGER NOT NULL DEFAULT 0,
			creep_damage INTEGER NOT NULL DEFAULT 0,
			hero_level INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (match_id, player_slot, game_time_s)
		)`,

		`CREATE TABLE IF NOT EXISTS wpa_item_results (
			hero_id INTEGER NOT NULL,
			item_id INTEGER NOT NULL,
			context_key TEXT NOT NULL,
			mean_delta_w REAL NOT NULL,
			mean_initial_w REAL NOT NULL,
			win_rate REAL NOT NULL,
			sample_size INTEGER NOT NULL,
			std_delta_w REAL NOT NULL DEFAULT 0,
			PRIMARY KEY (hero_id, item_id, context_key)
		)`,

		`CREATE TABLE IF NOT EXISTS model_metadata (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			trained_at INTEGER NOT NULL,
			accuracy REAL NOT NULL,
			ece REAL NOT NULL,
			num_matches INTEGER NOT NULL,
			weights BLOB,
			is_active INTEGER NOT NULL DEFAULT 1
		)`,

		// Indexes for common queries
		`CREATE INDEX IF NOT EXISTS idx_match_players_hero ON match_players(hero_id)`,
		`CREATE INDEX IF NOT EXISTS idx_match_players_match ON match_players(match_id)`,
		`CREATE INDEX IF NOT EXISTS idx_match_player_items_match ON match_player_items(match_id, player_slot)`,
		`CREATE INDEX IF NOT EXISTS idx_match_player_items_item ON match_player_items(item_id)`,
		`CREATE INDEX IF NOT EXISTS idx_match_snapshots_match ON match_snapshots(match_id, player_slot)`,
		`CREATE INDEX IF NOT EXISTS idx_wpa_hero ON wpa_item_results(hero_id)`,
		`CREATE INDEX IF NOT EXISTS idx_wpa_hero_context ON wpa_item_results(hero_id, context_key)`,
		`CREATE INDEX IF NOT EXISTS idx_matches_mode ON matches(match_mode)`,
	}

	for _, m := range migrations {
		if _, err := db.conn.Exec(m); err != nil {
			return fmt.Errorf("executing migration: %w\nSQL: %s", err, m)
		}
	}
	return nil
}
