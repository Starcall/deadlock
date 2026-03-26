package store

import (
	"context"
	"fmt"

	"github.com/name/deadlock/internal/domain"
)

// UpsertWPAResult inserts or replaces a WPA result.
func (db *DB) UpsertWPAResult(ctx context.Context, r domain.WPAResult) error {
	_, err := db.conn.ExecContext(ctx,
		`INSERT OR REPLACE INTO wpa_item_results (hero_id, item_id, context_key, mean_delta_w, mean_initial_w, win_rate, sample_size, std_delta_w)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		r.HeroID, r.ItemID, r.ContextKey, r.MeanDeltaW, r.MeanInitialW, r.WinRate, r.SampleSize, r.StdDeltaW,
	)
	return err
}

// BulkUpsertWPAResults inserts multiple WPA results in a transaction.
func (db *DB) BulkUpsertWPAResults(ctx context.Context, results []domain.WPAResult) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT OR REPLACE INTO wpa_item_results (hero_id, item_id, context_key, mean_delta_w, mean_initial_w, win_rate, sample_size, std_delta_w)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, r := range results {
		if _, err := stmt.ExecContext(ctx, r.HeroID, r.ItemID, r.ContextKey, r.MeanDeltaW, r.MeanInitialW, r.WinRate, r.SampleSize, r.StdDeltaW); err != nil {
			return fmt.Errorf("inserting WPA result: %w", err)
		}
	}

	return tx.Commit()
}

// GetWPAForHero returns all WPA results for a hero, optionally filtered by context.
func (db *DB) GetWPAForHero(ctx context.Context, heroID int, contextKey string, minSampleSize int) ([]domain.WPAResult, error) {
	query := `SELECT hero_id, item_id, context_key, mean_delta_w, mean_initial_w, win_rate, sample_size, std_delta_w
		FROM wpa_item_results WHERE hero_id = ? AND sample_size >= ?`
	args := []any{heroID, minSampleSize}

	if contextKey != "" {
		query += " AND context_key = ?"
		args = append(args, contextKey)
	}

	query += " ORDER BY mean_delta_w DESC"

	rows, err := db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying WPA results: %w", err)
	}
	defer rows.Close()

	results := make([]domain.WPAResult, 0)
	for rows.Next() {
		var r domain.WPAResult
		if err := rows.Scan(&r.HeroID, &r.ItemID, &r.ContextKey, &r.MeanDeltaW, &r.MeanInitialW, &r.WinRate, &r.SampleSize, &r.StdDeltaW); err != nil {
			return nil, fmt.Errorf("scanning WPA result: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// GetWPAForHeroItem returns WPA results for a specific hero/item across all contexts.
func (db *DB) GetWPAForHeroItem(ctx context.Context, heroID int, itemID int64) ([]domain.WPAResult, error) {
	rows, err := db.conn.QueryContext(ctx,
		`SELECT hero_id, item_id, context_key, mean_delta_w, mean_initial_w, win_rate, sample_size, std_delta_w
		 FROM wpa_item_results WHERE hero_id = ? AND item_id = ? ORDER BY context_key`,
		heroID, itemID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying WPA results: %w", err)
	}
	defer rows.Close()

	results2 := make([]domain.WPAResult, 0)
	for rows.Next() {
		var r domain.WPAResult
		if err := rows.Scan(&r.HeroID, &r.ItemID, &r.ContextKey, &r.MeanDeltaW, &r.MeanInitialW, &r.WinRate, &r.SampleSize, &r.StdDeltaW); err != nil {
			return nil, fmt.Errorf("scanning WPA result: %w", err)
		}
		results2 = append(results2, r)
	}
	return results2, rows.Err()
}

// ClearWPAResults deletes all WPA results (before recomputation).
func (db *DB) ClearWPAResults(ctx context.Context) error {
	_, err := db.conn.ExecContext(ctx, "DELETE FROM wpa_item_results")
	return err
}

// InsertModelMetadata stores model training metadata.
func (db *DB) InsertModelMetadata(ctx context.Context, m domain.ModelMetadata) error {
	// Deactivate previous models
	if _, err := db.conn.ExecContext(ctx, "UPDATE model_metadata SET is_active = 0"); err != nil {
		return fmt.Errorf("deactivating models: %w", err)
	}

	_, err := db.conn.ExecContext(ctx,
		`INSERT INTO model_metadata (trained_at, accuracy, ece, num_matches, weights, is_active)
		 VALUES (?, ?, ?, ?, ?, 1)`,
		m.TrainedAt, m.Accuracy, m.ECE, m.NumMatches, m.Weights,
	)
	return err
}

// GetActiveModel returns the currently active model metadata.
func (db *DB) GetActiveModel(ctx context.Context) (*domain.ModelMetadata, error) {
	var m domain.ModelMetadata
	err := db.conn.QueryRowContext(ctx,
		`SELECT id, trained_at, accuracy, ece, num_matches, weights, is_active
		 FROM model_metadata WHERE is_active = 1 ORDER BY id DESC LIMIT 1`,
	).Scan(&m.ID, &m.TrainedAt, &m.Accuracy, &m.ECE, &m.NumMatches, &m.Weights, &m.IsActive)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// GetStatus returns data freshness info.
func (db *DB) GetStatus(ctx context.Context) (matchCount int, latestMatchTime int64, modelAccuracy float64, err error) {
	err = db.conn.QueryRowContext(ctx, "SELECT COUNT(*), COALESCE(MAX(start_time), 0) FROM matches").Scan(&matchCount, &latestMatchTime)
	if err != nil {
		return
	}
	_ = db.conn.QueryRowContext(ctx, "SELECT accuracy FROM model_metadata WHERE is_active = 1 ORDER BY id DESC LIMIT 1").Scan(&modelAccuracy)
	return
}
