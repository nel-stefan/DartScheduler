package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"DartScheduler/domain"

	"github.com/google/uuid"
)

type MatchRepo struct{ db *sql.DB }

func NewMatchRepo(db *sql.DB) *MatchRepo { return &MatchRepo{db: db} }

func (r *MatchRepo) Save(ctx context.Context, m domain.Match) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO matches(
			id, evening_id, player_a, player_b, score_a, score_b, played,
			leg1_winner, leg1_turns, leg2_winner, leg2_turns,
			leg3_winner, leg3_turns,
			reported_by, reschedule_date, secretary_nr, counter_nr)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET
			score_a=excluded.score_a, score_b=excluded.score_b, played=excluded.played,
			leg1_winner=excluded.leg1_winner, leg1_turns=excluded.leg1_turns,
			leg2_winner=excluded.leg2_winner, leg2_turns=excluded.leg2_turns,
			leg3_winner=excluded.leg3_winner, leg3_turns=excluded.leg3_turns,
			reported_by=excluded.reported_by, reschedule_date=excluded.reschedule_date,
			secretary_nr=excluded.secretary_nr, counter_nr=excluded.counter_nr`,
		m.ID.String(), m.EveningID.String(), m.PlayerA.String(), m.PlayerB.String(),
		m.ScoreA, m.ScoreB, boolToInt(m.Played),
		m.Leg1Winner, m.Leg1Turns, m.Leg2Winner, m.Leg2Turns,
		m.Leg3Winner, m.Leg3Turns,
		m.ReportedBy, m.RescheduleDate, m.SecretaryNr, m.CounterNr)
	return err
}

func (r *MatchRepo) SaveBatch(ctx context.Context, matches []domain.Match) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO matches(
			id, evening_id, player_a, player_b, score_a, score_b, played,
			leg1_winner, leg1_turns, leg2_winner, leg2_turns,
			leg3_winner, leg3_turns,
			reported_by, reschedule_date, secretary_nr, counter_nr)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET
			score_a=excluded.score_a, score_b=excluded.score_b, played=excluded.played,
			leg1_winner=excluded.leg1_winner, leg1_turns=excluded.leg1_turns,
			leg2_winner=excluded.leg2_winner, leg2_turns=excluded.leg2_turns,
			leg3_winner=excluded.leg3_winner, leg3_turns=excluded.leg3_turns,
			reported_by=excluded.reported_by, reschedule_date=excluded.reschedule_date,
			secretary_nr=excluded.secretary_nr, counter_nr=excluded.counter_nr`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range matches {
		if _, err := stmt.ExecContext(ctx,
			m.ID.String(), m.EveningID.String(), m.PlayerA.String(), m.PlayerB.String(),
			m.ScoreA, m.ScoreB, boolToInt(m.Played),
			m.Leg1Winner, m.Leg1Turns, m.Leg2Winner, m.Leg2Turns,
			m.Leg3Winner, m.Leg3Turns,
			m.ReportedBy, m.RescheduleDate, m.SecretaryNr, m.CounterNr); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *MatchRepo) FindByID(ctx context.Context, id domain.MatchID) (domain.Match, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, evening_id, player_a, player_b, score_a, score_b, played,
			leg1_winner, leg1_turns, leg2_winner, leg2_turns,
			leg3_winner, leg3_turns,
			reported_by, reschedule_date, secretary_nr, counter_nr
		FROM matches WHERE id=?`, id.String())
	return scanMatch(row)
}

func (r *MatchRepo) FindByEvening(ctx context.Context, eveningID domain.EveningID) ([]domain.Match, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, evening_id, player_a, player_b, score_a, score_b, played,
			leg1_winner, leg1_turns, leg2_winner, leg2_turns,
			leg3_winner, leg3_turns,
			reported_by, reschedule_date, secretary_nr, counter_nr
		FROM matches WHERE evening_id=?`, eveningID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatches(rows)
}

func (r *MatchRepo) FindByPlayer(ctx context.Context, playerID domain.PlayerID) ([]domain.Match, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, evening_id, player_a, player_b, score_a, score_b, played,
			leg1_winner, leg1_turns, leg2_winner, leg2_turns,
			leg3_winner, leg3_turns,
			reported_by, reschedule_date, secretary_nr, counter_nr
		FROM matches WHERE player_a=? OR player_b=?`,
		playerID.String(), playerID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatches(rows)
}

func (r *MatchRepo) FindByPlayerAndSchedule(ctx context.Context, playerID domain.PlayerID, scheduleID domain.ScheduleID) ([]domain.Match, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT m.id, m.evening_id, m.player_a, m.player_b, m.score_a, m.score_b, m.played,
            m.leg1_winner, m.leg1_turns, m.leg2_winner, m.leg2_turns,
            m.leg3_winner, m.leg3_turns,
            m.reported_by, m.reschedule_date, m.secretary_nr, m.counter_nr
        FROM matches m
        JOIN evenings e ON m.evening_id = e.id
        WHERE (m.player_a=? OR m.player_b=?) AND e.schedule_id=?`,
		playerID.String(), playerID.String(), scheduleID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatches(rows)
}

func (r *MatchRepo) FindAllPlayed(ctx context.Context) ([]domain.Match, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, evening_id, player_a, player_b, score_a, score_b, played,
            leg1_winner, leg1_turns, leg2_winner, leg2_turns,
            leg3_winner, leg3_turns,
            reported_by, reschedule_date, secretary_nr, counter_nr
        FROM matches WHERE played=1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatches(rows)
}

func (r *MatchRepo) UpdateResult(ctx context.Context, m domain.Match) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE matches SET
			score_a=?, score_b=?, played=?,
			leg1_winner=?, leg1_turns=?,
			leg2_winner=?, leg2_turns=?,
			leg3_winner=?, leg3_turns=?,
			reported_by=?, reschedule_date=?,
			secretary_nr=?, counter_nr=?
		WHERE id=?`,
		m.ScoreA, m.ScoreB, boolToInt(m.Played),
		m.Leg1Winner, m.Leg1Turns,
		m.Leg2Winner, m.Leg2Turns,
		m.Leg3Winner, m.Leg3Turns,
		m.ReportedBy, m.RescheduleDate,
		m.SecretaryNr, m.CounterNr,
		m.ID.String())
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanMatch(s scanner) (domain.Match, error) {
	var m domain.Match
	var idStr, eveningStr, paStr, pbStr string
	var played int
	if err := s.Scan(
		&idStr, &eveningStr, &paStr, &pbStr, &m.ScoreA, &m.ScoreB, &played,
		&m.Leg1Winner, &m.Leg1Turns, &m.Leg2Winner, &m.Leg2Turns,
		&m.Leg3Winner, &m.Leg3Turns,
		&m.ReportedBy, &m.RescheduleDate, &m.SecretaryNr, &m.CounterNr,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return m, domain.ErrNotFound
		}
		return m, err
	}
	var err error
	if m.ID, err = uuid.Parse(idStr); err != nil {
		return m, fmt.Errorf("invalid match id: %w", err)
	}
	if m.EveningID, err = uuid.Parse(eveningStr); err != nil {
		return m, fmt.Errorf("invalid evening id: %w", err)
	}
	if m.PlayerA, err = uuid.Parse(paStr); err != nil {
		return m, fmt.Errorf("invalid player_a: %w", err)
	}
	if m.PlayerB, err = uuid.Parse(pbStr); err != nil {
		return m, fmt.Errorf("invalid player_b: %w", err)
	}
	m.Played = played != 0
	return m, nil
}

func scanMatches(rows *sql.Rows) ([]domain.Match, error) {
	out := make([]domain.Match, 0)
	for rows.Next() {
		m, err := scanMatch(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *MatchRepo) FindCancelledByScheduleBeforeDate(ctx context.Context, scheduleID domain.ScheduleID, before time.Time) ([]domain.Match, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT m.id, m.evening_id, m.player_a, m.player_b, m.score_a, m.score_b, m.played,
			m.leg1_winner, m.leg1_turns, m.leg2_winner, m.leg2_turns,
			m.leg3_winner, m.leg3_turns,
			m.reported_by, m.reschedule_date, m.secretary_nr, m.counter_nr
		FROM matches m
		JOIN evenings e ON m.evening_id = e.id
		WHERE e.schedule_id = ?
		  AND e.is_inhaal_avond = 0
		  AND datetime(e.date) < datetime(?)
		  AND m.reported_by != ''
		  AND m.played = 0
		ORDER BY e.date, m.rowid`,
		scheduleID.String(), before.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatches(rows)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
