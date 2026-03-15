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
	inhaal := 0
	if e.IsInhaalAvond {
		inhaal = 1
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO evenings(id,schedule_id,number,date,is_inhaal_avond) VALUES(?,?,?,?,?)
         ON CONFLICT(id) DO UPDATE SET number=excluded.number, date=excluded.date, is_inhaal_avond=excluded.is_inhaal_avond`,
		e.ID.String(), scheduleID.String(), e.Number, e.Date, inhaal)
	return err
}

func (r *EveningRepo) FindByID(ctx context.Context, id domain.EveningID) (domain.Evening, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,number,date,is_inhaal_avond FROM evenings WHERE id=?`, id.String())
	return scanEvening(row)
}

func (r *EveningRepo) FindBySchedule(ctx context.Context, scheduleID domain.ScheduleID) ([]domain.Evening, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,number,date,is_inhaal_avond FROM evenings WHERE schedule_id=? ORDER BY number`, scheduleID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Evening, 0)
	for rows.Next() {
		e, err := scanEvening(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *EveningRepo) Delete(ctx context.Context, id domain.EveningID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM evenings WHERE id=?`, id.String())
	return err
}

func (r *EveningRepo) DeleteBySchedule(ctx context.Context, scheduleID domain.ScheduleID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM evenings WHERE schedule_id=?`, scheduleID.String())
	return err
}

func scanEvening(s scanner) (domain.Evening, error) {
	var e domain.Evening
	var idStr string
	var inhaal int
	if err := s.Scan(&idStr, &e.Number, &e.Date, &inhaal); err != nil {
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
	e.IsInhaalAvond = inhaal != 0
	return e, nil
}
