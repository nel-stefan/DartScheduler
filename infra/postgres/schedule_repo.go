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

// ScheduleRepo implements domain.ScheduleRepository using PostgreSQL.
type ScheduleRepo struct{ pool *pgxpool.Pool }

func NewScheduleRepo(pool *pgxpool.Pool) *ScheduleRepo { return &ScheduleRepo{pool: pool} }

func (r *ScheduleRepo) Save(ctx context.Context, s domain.Schedule) error {
	var listID *uuid.UUID
	if s.PlayerListID != nil {
		v := *s.PlayerListID
		listID = &v
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO schedules(id, competition_name, season, active, created_at, player_list_id)
		 VALUES($1, $2, $3, $4, $5, $6)
		 ON CONFLICT(id) DO UPDATE SET
		   competition_name = EXCLUDED.competition_name,
		   season           = EXCLUDED.season,
		   active           = EXCLUDED.active,
		   player_list_id   = EXCLUDED.player_list_id`,
		s.ID, s.CompetitionName, s.Season, s.Active, s.CreatedAt, listID)
	return err
}

func (r *ScheduleRepo) FindLatest(ctx context.Context) (domain.Schedule, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, competition_name, season, active, created_at, player_list_id
		 FROM schedules ORDER BY created_at DESC LIMIT 1`)
	return scanSchedule(row)
}

func (r *ScheduleRepo) FindByID(ctx context.Context, id domain.ScheduleID) (domain.Schedule, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, competition_name, season, active, created_at, player_list_id
		 FROM schedules WHERE id = $1`, id)
	return scanSchedule(row)
}

func (r *ScheduleRepo) FindAll(ctx context.Context) ([]domain.Schedule, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, competition_name, season, active, created_at, player_list_id
		 FROM schedules ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Schedule, 0)
	for rows.Next() {
		s, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *ScheduleRepo) Delete(ctx context.Context, id domain.ScheduleID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM schedules WHERE id = $1`, id)
	return err
}

// SetActive marks the given schedule as active and all others as inactive.
func (r *ScheduleRepo) SetActive(ctx context.Context, id domain.ScheduleID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `UPDATE schedules SET active = FALSE`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE schedules SET active = TRUE WHERE id = $1`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// pgxScanner is satisfied by both pgx.Row and pgx.Rows.
type pgxScanner interface {
	Scan(dest ...any) error
}

func scanSchedule(s pgxScanner) (domain.Schedule, error) {
	var sc domain.Schedule
	var listID *uuid.UUID
	if err := s.Scan(&sc.ID, &sc.CompetitionName, &sc.Season, &sc.Active, &sc.CreatedAt, &listID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sc, domain.ErrNotFound
		}
		return sc, fmt.Errorf("scan schedule: %w", err)
	}
	sc.PlayerListID = listID
	return sc, nil
}
