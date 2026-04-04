package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"DartScheduler/domain"

	"github.com/google/uuid"
)

type ScheduleRepo struct{ db *sql.DB }

func NewScheduleRepo(db *sql.DB) *ScheduleRepo { return &ScheduleRepo{db: db} }

func (r *ScheduleRepo) Save(ctx context.Context, s domain.Schedule) error {
	active := 0
	if s.Active {
		active = 1
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO schedules(id,competition_name,season,active,created_at) VALUES(?,?,?,?,?)
         ON CONFLICT(id) DO UPDATE SET competition_name=excluded.competition_name,season=excluded.season,active=excluded.active`,
		s.ID.String(), s.CompetitionName, s.Season, active, s.CreatedAt)
	return err
}

func (r *ScheduleRepo) FindLatest(ctx context.Context) (domain.Schedule, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,competition_name,season,active,created_at FROM schedules ORDER BY created_at DESC LIMIT 1`)
	return scanSchedule(row)
}

func (r *ScheduleRepo) FindByID(ctx context.Context, id domain.ScheduleID) (domain.Schedule, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,competition_name,season,active,created_at FROM schedules WHERE id=?`, id.String())
	return scanSchedule(row)
}

func (r *ScheduleRepo) FindAll(ctx context.Context) ([]domain.Schedule, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,competition_name,season,active,created_at FROM schedules ORDER BY created_at DESC`)
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
	_, err := r.db.ExecContext(ctx, `DELETE FROM schedules WHERE id=?`, id.String())
	return err
}

// SetActive marks the given schedule as active and all others as inactive.
func (r *ScheduleRepo) SetActive(ctx context.Context, id domain.ScheduleID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `UPDATE schedules SET active = 0`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE schedules SET active = 1 WHERE id = ?`, id.String()); err != nil {
		return err
	}
	return tx.Commit()
}

func scanSchedule(s scanner) (domain.Schedule, error) {
	var sc domain.Schedule
	var idStr string
	var active int
	if err := s.Scan(&idStr, &sc.CompetitionName, &sc.Season, &active, &sc.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sc, domain.ErrNotFound
		}
		return sc, err
	}
	sc.Active = active != 0
	uid, err := uuid.Parse(idStr)
	if err != nil {
		return sc, fmt.Errorf("invalid schedule id %q: %w", idStr, err)
	}
	sc.ID = uid
	return sc, nil
}
