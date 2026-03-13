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
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO schedules(id,competition_name,created_at) VALUES(?,?,?)
         ON CONFLICT(id) DO UPDATE SET competition_name=excluded.competition_name`,
		s.ID.String(), s.CompetitionName, s.CreatedAt)
	return err
}

func (r *ScheduleRepo) FindLatest(ctx context.Context) (domain.Schedule, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,competition_name,created_at FROM schedules ORDER BY created_at DESC LIMIT 1`)
	return scanSchedule(row)
}

func (r *ScheduleRepo) FindByID(ctx context.Context, id domain.ScheduleID) (domain.Schedule, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,competition_name,created_at FROM schedules WHERE id=?`, id.String())
	return scanSchedule(row)
}

func scanSchedule(s scanner) (domain.Schedule, error) {
	var sc domain.Schedule
	var idStr string
	if err := s.Scan(&idStr, &sc.CompetitionName, &sc.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sc, domain.ErrNotFound
		}
		return sc, err
	}
	uid, err := uuid.Parse(idStr)
	if err != nil {
		return sc, fmt.Errorf("invalid schedule id %q: %w", idStr, err)
	}
	sc.ID = uid
	return sc, nil
}
