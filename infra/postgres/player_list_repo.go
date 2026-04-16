package postgres

import (
	"context"
	"errors"
	"fmt"

	"DartScheduler/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PlayerListRepo implements domain.PlayerListRepository using PostgreSQL.
type PlayerListRepo struct{ pool *pgxpool.Pool }

func NewPlayerListRepo(pool *pgxpool.Pool) *PlayerListRepo { return &PlayerListRepo{pool: pool} }

func (r *PlayerListRepo) Save(ctx context.Context, list domain.PlayerList) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO player_lists(id, name, created_at) VALUES($1, $2, $3)
		 ON CONFLICT(id) DO UPDATE SET name = EXCLUDED.name`,
		list.ID, list.Name, list.CreatedAt.UTC())
	return err
}

func (r *PlayerListRepo) FindAll(ctx context.Context) ([]domain.PlayerList, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, created_at FROM player_lists ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lists []domain.PlayerList
	for rows.Next() {
		var l domain.PlayerList
		var id uuid.UUID
		if err := rows.Scan(&id, &l.Name, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan player_list: %w", err)
		}
		l.ID = id
		lists = append(lists, l)
	}
	return lists, rows.Err()
}

func (r *PlayerListRepo) FindByName(ctx context.Context, name string) (domain.PlayerList, bool, error) {
	var l domain.PlayerList
	var id uuid.UUID
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, created_at FROM player_lists WHERE name = $1`, name).
		Scan(&id, &l.Name, &l.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return l, false, nil
		}
		return l, false, fmt.Errorf("find player_list by name: %w", err)
	}
	l.ID = id
	return l, true, nil
}
