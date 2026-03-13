package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"DartScheduler/domain"

	"github.com/google/uuid"
)

type EveningRepo struct{ db *sql.DB }

func NewEveningRepo(db *sql.DB) *EveningRepo { return &EveningRepo{db: db} }

func (r *EveningRepo) Save(ctx context.Context, e domain.Evening, scheduleID domain.ScheduleID) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO evenings(id,schedule_id,number,date) VALUES(?,?,?,?)
         ON CONFLICT(id) DO UPDATE SET number=excluded.number, date=excluded.date`,
		e.ID.String(), scheduleID.String(), e.Number, e.Date)
	return err
}

func (r *EveningRepo) FindByID(ctx context.Context, id domain.EveningID) (domain.Evening, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,number,date FROM evenings WHERE id=?`, id.String())
	return scanEvening(row)
}

func (r *EveningRepo) FindBySchedule(ctx context.Context, scheduleID domain.ScheduleID) ([]domain.Evening, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,number,date FROM evenings WHERE schedule_id=? ORDER BY number`, scheduleID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Evening
	for rows.Next() {
		e, err := scanEvening(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func scanEvening(s scanner) (domain.Evening, error) {
	var e domain.Evening
	var idStr string
	if err := s.Scan(&idStr, &e.Number, &e.Date); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e, domain.ErrNotFound
		}
		return e, err
	}
	uid, err := uuid.Parse(idStr)
	if err != nil {
		return e, fmt.Errorf("invalid evening id %q: %w", idStr, err)
	}
	e.ID = uid
	return e, nil
}
