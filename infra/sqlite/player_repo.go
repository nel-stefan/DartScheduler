package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"DartScheduler/domain"

	"github.com/google/uuid"
)

type PlayerRepo struct{ db *sql.DB }

func NewPlayerRepo(db *sql.DB) *PlayerRepo { return &PlayerRepo{db: db} }

func (r *PlayerRepo) Save(ctx context.Context, p domain.Player) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO players(id,name,email,sponsor) VALUES(?,?,?,?)
         ON CONFLICT(id) DO UPDATE SET name=excluded.name, email=excluded.email, sponsor=excluded.sponsor`,
		p.ID.String(), p.Name, p.Email, p.Sponsor)
	return err
}

func (r *PlayerRepo) SaveBatch(ctx context.Context, players []domain.Player) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO players(id,name,email,sponsor) VALUES(?,?,?,?)
         ON CONFLICT(id) DO UPDATE SET name=excluded.name, email=excluded.email, sponsor=excluded.sponsor`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range players {
		if _, err := stmt.ExecContext(ctx, p.ID.String(), p.Name, p.Email, p.Sponsor); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *PlayerRepo) FindByID(ctx context.Context, id domain.PlayerID) (domain.Player, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,name,email,sponsor FROM players WHERE id=?`, id.String())
	return scanPlayer(row)
}

func (r *PlayerRepo) FindAll(ctx context.Context) ([]domain.Player, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,name,email,sponsor FROM players ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Player
	for rows.Next() {
		p, err := scanPlayer(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PlayerRepo) DeleteAll(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM players`)
	return err
}

func (r *PlayerRepo) SaveBuddyPreference(ctx context.Context, bp domain.BuddyPreference) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO buddy_preferences(player_id,buddy_id) VALUES(?,?)`,
		bp.PlayerID.String(), bp.BuddyID.String())
	return err
}

func (r *PlayerRepo) FindBuddiesForPlayer(ctx context.Context, id domain.PlayerID) ([]domain.PlayerID, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT buddy_id FROM buddy_preferences WHERE player_id=?`, id.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.PlayerID
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		uid, err := uuid.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("invalid buddy_id %q: %w", s, err)
		}
		out = append(out, uid)
	}
	return out, rows.Err()
}

func (r *PlayerRepo) FindAllBuddyPairs(ctx context.Context) ([]domain.BuddyPreference, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT player_id, buddy_id FROM buddy_preferences`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.BuddyPreference
	for rows.Next() {
		var ps, bs string
		if err := rows.Scan(&ps, &bs); err != nil {
			return nil, err
		}
		pid, err := uuid.Parse(ps)
		if err != nil {
			return nil, err
		}
		bid, err := uuid.Parse(bs)
		if err != nil {
			return nil, err
		}
		out = append(out, domain.BuddyPreference{PlayerID: pid, BuddyID: bid})
	}
	return out, rows.Err()
}

// scanPlayer works for both *sql.Row and *sql.Rows via a shared interface.
type scanner interface {
	Scan(dest ...any) error
}

func scanPlayer(s scanner) (domain.Player, error) {
	var p domain.Player
	var idStr string
	if err := s.Scan(&idStr, &p.Name, &p.Email, &p.Sponsor); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return p, domain.ErrNotFound
		}
		return p, err
	}
	uid, err := uuid.Parse(idStr)
	if err != nil {
		return p, fmt.Errorf("invalid player id %q: %w", idStr, err)
	}
	p.ID = uid
	return p, nil
}
