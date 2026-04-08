// Package sqlite implements the domain repository interfaces using a SQLite database
// via modernc.org/sqlite (a pure-Go SQLite driver without CGO).
//
// Schema migration runs in two steps:
//  1. The base schema (schema.sql) is executed via CREATE TABLE IF NOT EXISTS.
//  2. Column additions (ALTER TABLE ADD COLUMN) are applied for existing databases;
//     duplicate-column errors are silently ignored.
package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

// alterations are ALTER TABLE statements applied after the base schema to
// migrate existing databases. Duplicate-column errors are silently ignored.
var alterations = []string{
	`ALTER TABLE players ADD COLUMN nr           TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE players ADD COLUMN address      TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE players ADD COLUMN postal_code  TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE players ADD COLUMN city         TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE players ADD COLUMN phone        TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE players ADD COLUMN mobile       TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE players ADD COLUMN member_since TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE matches ADD COLUMN leg1_winner TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE matches ADD COLUMN leg1_turns  INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE matches ADD COLUMN leg2_winner TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE matches ADD COLUMN leg2_turns  INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE matches ADD COLUMN leg3_winner TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE matches ADD COLUMN leg3_turns  INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE matches ADD COLUMN reported_by TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE matches ADD COLUMN reschedule_date TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE matches ADD COLUMN secretary_nr TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE matches ADD COLUMN counter_nr TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE players ADD COLUMN class TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE schedules ADD COLUMN season TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE evenings ADD COLUMN is_inhaal_avond INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE players ADD COLUMN schedule_id TEXT REFERENCES schedules(id)`,
	`ALTER TABLE matches ADD COLUMN player_a_180s INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE matches ADD COLUMN player_b_180s INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE matches ADD COLUMN player_a_highest_finish INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE matches ADD COLUMN player_b_highest_finish INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE matches ADD COLUMN played_date TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE schedules ADD COLUMN active INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE players ADD COLUMN list_id TEXT REFERENCES player_lists(id)`,
}

// dataMigrations are named, one-time data migrations run after schema setup.
// Each entry is applied only once, tracked in the applied_migrations table.
var dataMigrations = []struct {
	name string
	fn   func(ctx context.Context, db *sql.DB) error
}{
	{
		name: "migrate_180s_hf_to_season_player_stats",
		fn:   migrate180sToSeasonStats,
	},
	{
		name: "assign_orphan_players_to_backup_list",
		fn:   assignOrphanPlayersToBackupList,
	},
}

// Open opens (or creates) a SQLite database at the given path, runs the
// embedded schema migration and returns the ready-to-use *sql.DB.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open %s: %w", path, err)
	}
	// WAL mode and foreign keys are set per connection; keep a single writer.
	db.SetMaxOpenConns(1)

	ctx := context.Background()
	if _, err = db.ExecContext(ctx, schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite: migrate: %w", err)
	}
	// Apply column additions for existing databases; ignore duplicate-column errors.
	for _, stmt := range alterations {
		if _, err := db.ExecContext(ctx, stmt); err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			db.Close()
			return nil, fmt.Errorf("sqlite: alter: %w", err)
		}
	}
	// Run named data migrations once.
	for _, m := range dataMigrations {
		var applied bool
		_ = db.QueryRowContext(ctx, `SELECT 1 FROM applied_migrations WHERE name = ?`, m.name).Scan(&applied)
		if applied {
			continue
		}
		if err := m.fn(ctx, db); err != nil {
			db.Close()
			return nil, fmt.Errorf("sqlite: data migration %q: %w", m.name, err)
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO applied_migrations(name) VALUES(?)`, m.name); err != nil {
			db.Close()
			return nil, fmt.Errorf("sqlite: record migration %q: %w", m.name, err)
		}
	}
	return db, nil
}

// migrate180sToSeasonStats aggregates 180s and highest-finish from existing
// match columns and evening_player_stats into season_player_stats, then
// resets the source columns/rows so the data lives in one place.
func migrate180sToSeasonStats(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	// Aggregate player_a stats from matches.
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO season_player_stats (schedule_id, player_id, one_eighties, highest_finish)
		SELECT e.schedule_id, m.player_a, SUM(m.player_a_180s), MAX(m.player_a_highest_finish)
		FROM matches m
		JOIN evenings e ON m.evening_id = e.id
		WHERE m.player_a_180s > 0 OR m.player_a_highest_finish > 0
		GROUP BY e.schedule_id, m.player_a
		ON CONFLICT(schedule_id, player_id) DO UPDATE SET
		  one_eighties   = season_player_stats.one_eighties + excluded.one_eighties,
		  highest_finish = MAX(season_player_stats.highest_finish, excluded.highest_finish)
	`); err != nil {
		return err
	}

	// Aggregate player_b stats from matches.
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO season_player_stats (schedule_id, player_id, one_eighties, highest_finish)
		SELECT e.schedule_id, m.player_b, SUM(m.player_b_180s), MAX(m.player_b_highest_finish)
		FROM matches m
		JOIN evenings e ON m.evening_id = e.id
		WHERE m.player_b_180s > 0 OR m.player_b_highest_finish > 0
		GROUP BY e.schedule_id, m.player_b
		ON CONFLICT(schedule_id, player_id) DO UPDATE SET
		  one_eighties   = season_player_stats.one_eighties + excluded.one_eighties,
		  highest_finish = MAX(season_player_stats.highest_finish, excluded.highest_finish)
	`); err != nil {
		return err
	}

	// Aggregate from evening_player_stats.
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO season_player_stats (schedule_id, player_id, one_eighties, highest_finish)
		SELECT e.schedule_id, eps.player_id, SUM(eps.one_eighties), MAX(eps.highest_finish)
		FROM evening_player_stats eps
		JOIN evenings e ON eps.evening_id = e.id
		GROUP BY e.schedule_id, eps.player_id
		ON CONFLICT(schedule_id, player_id) DO UPDATE SET
		  one_eighties   = season_player_stats.one_eighties + excluded.one_eighties,
		  highest_finish = MAX(season_player_stats.highest_finish, excluded.highest_finish)
	`); err != nil {
		return err
	}

	// Clear the now-redundant source data.
	if _, err := tx.ExecContext(ctx, `
		UPDATE matches SET
		  player_a_180s = 0, player_b_180s = 0,
		  player_a_highest_finish = 0, player_b_highest_finish = 0
	`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM evening_player_stats`); err != nil {
		return err
	}

	return tx.Commit()
}

// assignOrphanPlayersToBackupList moves all players without a list_id into a
// new "Oud" player list so that future imports cannot overwrite them.
func assignOrphanPlayersToBackupList(ctx context.Context, db *sql.DB) error {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM players WHERE list_id IS NULL`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return nil // nothing to migrate
	}

	listID := uuid.New()
	if _, err := db.ExecContext(ctx,
		`INSERT INTO player_lists(id, name, created_at) VALUES(?, 'Oud', datetime('now'))`,
		listID.String()); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx,
		`UPDATE players SET list_id = ? WHERE list_id IS NULL`, listID.String()); err != nil {
		return err
	}
	return nil
}
