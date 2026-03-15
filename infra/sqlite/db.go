// Package sqlite implementeert de domain repository-interfaces met een SQLite-database
// via modernc.org/sqlite (een pure-Go SQLite driver zonder CGO).
//
// Schema-migratie verloopt in twee stappen:
//  1. Het basisschema (schema.sql) wordt uitgevoerd via CREATE TABLE IF NOT EXISTS.
//  2. Alteraties (ALTER TABLE ADD COLUMN) worden toegepast voor bestaande databases;
//     fouten over dubbele kolommen worden stilzwijgend genegeerd.
package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"

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
	return db, nil
}
