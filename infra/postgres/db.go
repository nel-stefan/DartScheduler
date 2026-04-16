// Package postgres implements the domain repository interfaces using a PostgreSQL database
// via github.com/jackc/pgx/v5.
//
// Schema migration runs by executing schema.sql (CREATE TABLE IF NOT EXISTS statements).
// The embedded SQL is idempotent and safe to run on every startup.
package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schema string

// Open connects to PostgreSQL using the given DSN, runs the embedded schema migration,
// and returns the ready-to-use connection pool.
//
// dsn should be a libpq-compatible connection string, e.g.:
//
//	"postgres://user:pass@host:5432/dbname?sslmode=disable"
func Open(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}

	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: migrate: %w", err)
	}

	return pool, nil
}
