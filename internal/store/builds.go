package store

import (
	"context"
	"fmt"

	"github.com/name/deadlock/internal/domain"
)

// ClearBuildResults deletes all build templates and coverage data.
func (db *DB) ClearBuildResults(ctx context.Context) error {
	if _, err := db.conn.ExecContext(ctx, "DELETE FROM build_templates"); err != nil {
		return fmt.Errorf("clearing build templates: %w", err)
	}
	if _, err := db.conn.ExecContext(ctx, "DELETE FROM build_coverage"); err != nil {
		return fmt.Errorf("clearing build coverage: %w", err)
	}
	return nil
}

// BulkInsertBuildTemplates inserts build templates in a single transaction.
func (db *DB) BulkInsertBuildTemplates(ctx context.Context, templates []domain.BuildTemplate) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO build_templates
		(hero_id, build_rank, item_ids, exact_count, fuzzy_count, wins, losses, win_rate, total_hero_players)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, t := range templates {
		if _, err := stmt.ExecContext(ctx, t.HeroID, t.BuildRank, t.ItemIDs,
			t.ExactCount, t.FuzzyCount, t.Wins, t.Losses, t.WinRate, t.TotalHeroPlayers); err != nil {
			return fmt.Errorf("inserting build template: %w", err)
		}
	}

	return tx.Commit()
}

// BulkInsertBuildCoverage inserts build coverage stats in a single transaction.
func (db *DB) BulkInsertBuildCoverage(ctx context.Context, coverages []domain.HeroBuildCoverage) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO build_coverage
		(hero_id, total_players, classified_count, coverage)
		VALUES (?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, c := range coverages {
		if _, err := stmt.ExecContext(ctx, c.HeroID, c.TotalPlayers, c.ClassifiedCount, c.Coverage); err != nil {
			return fmt.Errorf("inserting build coverage: %w", err)
		}
	}

	return tx.Commit()
}

// GetBuildTemplatesForHero returns up to 10 build templates for a hero.
func (db *DB) GetBuildTemplatesForHero(ctx context.Context, heroID int) ([]domain.BuildTemplate, error) {
	rows, err := db.conn.QueryContext(ctx,
		`SELECT hero_id, build_rank, item_ids, exact_count, fuzzy_count, wins, losses, win_rate, total_hero_players
		FROM build_templates WHERE hero_id = ? ORDER BY build_rank`, heroID)
	if err != nil {
		return nil, fmt.Errorf("querying build templates: %w", err)
	}
	defer rows.Close()

	results := make([]domain.BuildTemplate, 0)
	for rows.Next() {
		var t domain.BuildTemplate
		if err := rows.Scan(&t.HeroID, &t.BuildRank, &t.ItemIDs, &t.ExactCount,
			&t.FuzzyCount, &t.Wins, &t.Losses, &t.WinRate, &t.TotalHeroPlayers); err != nil {
			return nil, fmt.Errorf("scanning build template: %w", err)
		}
		results = append(results, t)
	}
	return results, rows.Err()
}

// GetBuildCoverage returns build coverage for all heroes.
func (db *DB) GetBuildCoverage(ctx context.Context) ([]domain.HeroBuildCoverage, error) {
	rows, err := db.conn.QueryContext(ctx,
		`SELECT hero_id, total_players, classified_count, coverage FROM build_coverage ORDER BY hero_id`)
	if err != nil {
		return nil, fmt.Errorf("querying build coverage: %w", err)
	}
	defer rows.Close()

	results := make([]domain.HeroBuildCoverage, 0)
	for rows.Next() {
		var c domain.HeroBuildCoverage
		if err := rows.Scan(&c.HeroID, &c.TotalPlayers, &c.ClassifiedCount, &c.Coverage); err != nil {
			return nil, fmt.Errorf("scanning build coverage: %w", err)
		}
		results = append(results, c)
	}
	return results, rows.Err()
}

// GetBuildCoverageForHero returns build coverage for a specific hero.
func (db *DB) GetBuildCoverageForHero(ctx context.Context, heroID int) (*domain.HeroBuildCoverage, error) {
	var c domain.HeroBuildCoverage
	err := db.conn.QueryRowContext(ctx,
		`SELECT hero_id, total_players, classified_count, coverage FROM build_coverage WHERE hero_id = ?`, heroID).
		Scan(&c.HeroID, &c.TotalPlayers, &c.ClassifiedCount, &c.Coverage)
	if err != nil {
		return nil, fmt.Errorf("querying build coverage for hero %d: %w", heroID, err)
	}
	return &c, nil
}
