package sqlite

import (
	"context"
	"database/sql"
	"time"

	"DartScheduler/domain"

	"github.com/google/uuid"
)

// PlayerListRepo implements domain.PlayerListRepository using SQLite.
type PlayerListRepo struct{ db *sql.DB }

func NewPlayerListRepo(db *sql.DB) *PlayerListRepo { return &PlayerListRepo{db: db} }

func (r *PlayerListRepo) Save(ctx context.Context, list domain.PlayerList) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO player_lists(id, name, created_at) VALUES(?,?,?)
         ON CONFLICT(id) DO UPDATE SET name=excluded.name`,
		list.ID.String(), list.Name, list.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"))
	return err
}

func (r *PlayerListRepo) FindAll(ctx context.Context) ([]domain.PlayerList, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, created_at FROM player_lists ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lists []domain.PlayerList
	for rows.Next() {
		var l domain.PlayerList
		var idStr, createdAt string
		if err := rows.Scan(&idStr, &l.Name, &createdAt); err != nil {
			return nil, err
		}
		l.ID, _ = uuid.Parse(idStr)
		l.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdAt)
		lists = append(lists, l)
	}
	return lists, rows.Err()
}

func (r *PlayerListRepo) FindByName(ctx context.Context, name string) (domain.PlayerList, bool, error) {
	var l domain.PlayerList
	var idStr, createdAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, created_at FROM player_lists WHERE name = ?`, name).
		Scan(&idStr, &l.Name, &createdAt)
	if err == sql.ErrNoRows {
		return l, false, nil
	}
	if err != nil {
		return l, false, err
	}
	l.ID, _ = uuid.Parse(idStr)
	l.CreatedAt, _ = time.Parse("2006-01-02T15:04:05Z", createdAt)
	return l, true, nil
}
