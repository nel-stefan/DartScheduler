package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"

	"DartScheduler/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MatchRepo implements domain.MatchRepository using PostgreSQL.
type MatchRepo struct{ pool *pgxpool.Pool }

func NewMatchRepo(pool *pgxpool.Pool) *MatchRepo { return &MatchRepo{pool: pool} }

const matchColumns = `id, evening_id, player_a, player_b, score_a, score_b, played,
	leg1_winner, leg1_turns, leg2_winner, leg2_turns,
	leg3_winner, leg3_turns,
	reported_by, reschedule_date, secretary_nr, counter_nr,
	player_a_180s, player_b_180s, player_a_highest_finish, player_b_highest_finish,
	played_date`

const matchUpsert = `INSERT INTO matches(` + matchColumns + `)
	VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
	ON CONFLICT(id) DO UPDATE SET
	  score_a                 = EXCLUDED.score_a,
	  score_b                 = EXCLUDED.score_b,
	  played                  = EXCLUDED.played,
	  leg1_winner             = EXCLUDED.leg1_winner,
	  leg1_turns              = EXCLUDED.leg1_turns,
	  leg2_winner             = EXCLUDED.leg2_winner,
	  leg2_turns              = EXCLUDED.leg2_turns,
	  leg3_winner             = EXCLUDED.leg3_winner,
	  leg3_turns              = EXCLUDED.leg3_turns,
	  reported_by             = EXCLUDED.reported_by,
	  reschedule_date         = EXCLUDED.reschedule_date,
	  secretary_nr            = EXCLUDED.secretary_nr,
	  counter_nr              = EXCLUDED.counter_nr,
	  player_a_180s           = EXCLUDED.player_a_180s,
	  player_b_180s           = EXCLUDED.player_b_180s,
	  player_a_highest_finish = EXCLUDED.player_a_highest_finish,
	  player_b_highest_finish = EXCLUDED.player_b_highest_finish,
	  played_date             = EXCLUDED.played_date`

func matchArgs(m domain.Match) []any {
	return []any{
		m.ID, m.EveningID, m.PlayerA, m.PlayerB, m.ScoreA, m.ScoreB, m.Played,
		m.Leg1Winner, m.Leg1Turns, m.Leg2Winner, m.Leg2Turns,
		m.Leg3Winner, m.Leg3Turns,
		m.ReportedBy, m.RescheduleDate, m.SecretaryNr, m.CounterNr,
		m.PlayerA180s, m.PlayerB180s, m.PlayerAHighestFinish, m.PlayerBHighestFinish,
		m.PlayedDate,
	}
}

func (r *MatchRepo) Save(ctx context.Context, m domain.Match) error {
	_, err := r.pool.Exec(ctx, matchUpsert, matchArgs(m)...)
	return err
}

func (r *MatchRepo) SaveBatch(ctx context.Context, matches []domain.Match) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	for _, m := range matches {
		if _, err := tx.Exec(ctx, matchUpsert, matchArgs(m)...); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *MatchRepo) FindByID(ctx context.Context, id domain.MatchID) (domain.Match, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+matchColumns+` FROM matches WHERE id = $1`, id)
	return scanMatch(row)
}

func (r *MatchRepo) FindByEvening(ctx context.Context, eveningID domain.EveningID) ([]domain.Match, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+matchColumns+` FROM matches WHERE evening_id = $1`, eveningID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatches(rows)
}

func (r *MatchRepo) FindByPlayer(ctx context.Context, playerID domain.PlayerID) ([]domain.Match, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+matchColumns+` FROM matches WHERE player_a = $1 OR player_b = $1`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatches(rows)
}

func (r *MatchRepo) FindByPlayerAndSchedule(ctx context.Context, playerID domain.PlayerID, scheduleID domain.ScheduleID) ([]domain.Match, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT m.id, m.evening_id, m.player_a, m.player_b, m.score_a, m.score_b, m.played,
		        m.leg1_winner, m.leg1_turns, m.leg2_winner, m.leg2_turns,
		        m.leg3_winner, m.leg3_turns,
		        m.reported_by, m.reschedule_date, m.secretary_nr, m.counter_nr,
		        m.player_a_180s, m.player_b_180s, m.player_a_highest_finish, m.player_b_highest_finish,
		        m.played_date
		 FROM matches m
		 JOIN evenings e ON m.evening_id = e.id
		 WHERE (m.player_a = $1 OR m.player_b = $1) AND e.schedule_id = $2`,
		playerID, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatches(rows)
}

func (r *MatchRepo) FindAllPlayed(ctx context.Context) ([]domain.Match, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+matchColumns+` FROM matches WHERE played = TRUE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatches(rows)
}

func (r *MatchRepo) UpdateResult(ctx context.Context, m domain.Match) error {
	res, err := r.pool.Exec(ctx,
		`UPDATE matches SET
		   score_a                 = $1,
		   score_b                 = $2,
		   played                  = $3,
		   leg1_winner             = $4,
		   leg1_turns              = $5,
		   leg2_winner             = $6,
		   leg2_turns              = $7,
		   leg3_winner             = $8,
		   leg3_turns              = $9,
		   reported_by             = $10,
		   reschedule_date         = $11,
		   secretary_nr            = $12,
		   counter_nr              = $13,
		   player_a_180s           = $14,
		   player_b_180s           = $15,
		   player_a_highest_finish = $16,
		   player_b_highest_finish = $17,
		   played_date             = $18
		 WHERE id = $19`,
		m.ScoreA, m.ScoreB, m.Played,
		m.Leg1Winner, m.Leg1Turns,
		m.Leg2Winner, m.Leg2Turns,
		m.Leg3Winner, m.Leg3Turns,
		m.ReportedBy, m.RescheduleDate,
		m.SecretaryNr, m.CounterNr,
		m.PlayerA180s, m.PlayerB180s,
		m.PlayerAHighestFinish, m.PlayerBHighestFinish,
		m.PlayedDate,
		m.ID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *MatchRepo) FindBySchedule(ctx context.Context, scheduleID domain.ScheduleID) ([]domain.Match, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT m.id, m.evening_id, m.player_a, m.player_b, m.score_a, m.score_b, m.played,
		        m.leg1_winner, m.leg1_turns, m.leg2_winner, m.leg2_turns,
		        m.leg3_winner, m.leg3_turns,
		        m.reported_by, m.reschedule_date, m.secretary_nr, m.counter_nr,
		        m.player_a_180s, m.player_b_180s, m.player_a_highest_finish, m.player_b_highest_finish,
		        m.played_date
		 FROM matches m
		 WHERE m.evening_id IN (SELECT id FROM evenings WHERE schedule_id = $1)`,
		scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatches(rows)
}

func (r *MatchRepo) FindCancelledBySchedule(ctx context.Context, scheduleID domain.ScheduleID) ([]domain.Match, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT m.id, m.evening_id, m.player_a, m.player_b, m.score_a, m.score_b, m.played,
		        m.leg1_winner, m.leg1_turns, m.leg2_winner, m.leg2_turns,
		        m.leg3_winner, m.leg3_turns,
		        m.reported_by, m.reschedule_date, m.secretary_nr, m.counter_nr,
		        m.player_a_180s, m.player_b_180s, m.player_a_highest_finish, m.player_b_highest_finish,
		        m.played_date
		 FROM matches m
		 JOIN evenings e ON m.evening_id = e.id
		 WHERE e.schedule_id = $1
		   AND e.is_inhaal_avond = FALSE
		   AND m.reported_by != ''
		 ORDER BY e.date, m.id`,
		scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	matches, err := scanMatches(rows)
	log.Printf("[FindCancelledBySchedule] scheduleID=%s → %d matches", scheduleID, len(matches))
	return matches, err
}

func (r *MatchRepo) DeleteByEvening(ctx context.Context, eveningID domain.EveningID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM matches WHERE evening_id = $1`, eveningID)
	return err
}

func (r *MatchRepo) DeleteBySchedule(ctx context.Context, scheduleID domain.ScheduleID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM matches WHERE evening_id IN (SELECT id FROM evenings WHERE schedule_id = $1)`,
		scheduleID)
	return err
}

func (r *MatchRepo) DeleteByPlayer(ctx context.Context, playerID domain.PlayerID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM matches WHERE player_a = $1 OR player_b = $1`, playerID)
	return err
}

func scanMatch(s pgxScanner) (domain.Match, error) {
	var m domain.Match
	var id, eveningID, playerA, playerB uuid.UUID
	if err := s.Scan(
		&id, &eveningID, &playerA, &playerB, &m.ScoreA, &m.ScoreB, &m.Played,
		&m.Leg1Winner, &m.Leg1Turns, &m.Leg2Winner, &m.Leg2Turns,
		&m.Leg3Winner, &m.Leg3Turns,
		&m.ReportedBy, &m.RescheduleDate, &m.SecretaryNr, &m.CounterNr,
		&m.PlayerA180s, &m.PlayerB180s, &m.PlayerAHighestFinish, &m.PlayerBHighestFinish,
		&m.PlayedDate,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return m, domain.ErrNotFound
		}
		return m, fmt.Errorf("scan match: %w", err)
	}
	m.ID = id
	m.EveningID = eveningID
	m.PlayerA = playerA
	m.PlayerB = playerB
	return m, nil
}

func scanMatches(rows pgx.Rows) ([]domain.Match, error) {
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
