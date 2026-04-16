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

// EveningRepo implements domain.EveningRepository using PostgreSQL.
type EveningRepo struct{ pool *pgxpool.Pool }

func NewEveningRepo(pool *pgxpool.Pool) *EveningRepo { return &EveningRepo{pool: pool} }

func (r *EveningRepo) Save(ctx context.Context, e domain.Evening, scheduleID domain.ScheduleID) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO evenings(id, schedule_id, number, date, is_inhaal_avond)
		 VALUES($1, $2, $3, $4, $5)
		 ON CONFLICT(id) DO UPDATE SET
		   number          = EXCLUDED.number,
		   date            = EXCLUDED.date,
		   is_inhaal_avond = EXCLUDED.is_inhaal_avond`,
		e.ID, scheduleID, e.Number, e.Date, e.IsCatchUpEvening)
	return err
}

func (r *EveningRepo) FindByID(ctx context.Context, id domain.EveningID) (domain.Evening, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, number, date, is_inhaal_avond FROM evenings WHERE id = $1`, id)
	return scanEvening(row)
}

func (r *EveningRepo) FindBySchedule(ctx context.Context, scheduleID domain.ScheduleID) ([]domain.Evening, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, number, date, is_inhaal_avond FROM evenings WHERE schedule_id = $1 ORDER BY number`,
		scheduleID)
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
	_, err := r.pool.Exec(ctx, `DELETE FROM evenings WHERE id = $1`, id)
	return err
}

func (r *EveningRepo) DeleteBySchedule(ctx context.Context, scheduleID domain.ScheduleID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM evenings WHERE schedule_id = $1`, scheduleID)
	return err
}

func scanEvening(s pgxScanner) (domain.Evening, error) {
	var e domain.Evening
	var id uuid.UUID
	if err := s.Scan(&id, &e.Number, &e.Date, &e.IsCatchUpEvening); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return e, domain.ErrNotFound
		}
		return e, fmt.Errorf("scan evening: %w", err)
	}
	e.ID = id
	return e, nil
}
