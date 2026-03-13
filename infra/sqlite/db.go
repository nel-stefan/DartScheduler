package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

// Open opens (or creates) a SQLite database at the given path, runs the
// embedded schema migration and returns the ready-to-use *sql.DB.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open %s: %w", path, err)
	}
	// WAL mode and foreign keys are set per connection; keep a single writer.
	db.SetMaxOpenConns(1)

	if _, err = db.ExecContext(context.Background(), schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite: migrate: %w", err)
	}
	return db, nil
}
