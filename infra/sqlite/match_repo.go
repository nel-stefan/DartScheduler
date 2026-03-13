package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"DartScheduler/domain"

	"github.com/google/uuid"
)

type MatchRepo struct{ db *sql.DB }

func NewMatchRepo(db *sql.DB) *MatchRepo { return &MatchRepo{db: db} }

func (r *MatchRepo) Save(ctx context.Context, m domain.Match) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO matches(id,evening_id,player_a,player_b,score_a,score_b,played)
         VALUES(?,?,?,?,?,?,?)
         ON CONFLICT(id) DO UPDATE SET score_a=excluded.score_a, score_b=excluded.score_b, played=excluded.played`,
		m.ID.String(), m.EveningID.String(), m.PlayerA.String(), m.PlayerB.String(),
		m.ScoreA, m.ScoreB, boolToInt(m.Played))
	return err
}

func (r *MatchRepo) SaveBatch(ctx context.Context, matches []domain.Match) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO matches(id,evening_id,player_a,player_b,score_a,score_b,played)
         VALUES(?,?,?,?,?,?,?)
         ON CONFLICT(id) DO UPDATE SET score_a=excluded.score_a, score_b=excluded.score_b, played=excluded.played`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range matches {
		if _, err := stmt.ExecContext(ctx,
			m.ID.String(), m.EveningID.String(), m.PlayerA.String(), m.PlayerB.String(),
			m.ScoreA, m.ScoreB, boolToInt(m.Played)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *MatchRepo) FindByID(ctx context.Context, id domain.MatchID) (domain.Match, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id,evening_id,player_a,player_b,score_a,score_b,played FROM matches WHERE id=?`, id.String())
	return scanMatch(row)
}

func (r *MatchRepo) FindByEvening(ctx context.Context, eveningID domain.EveningID) ([]domain.Match, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,evening_id,player_a,player_b,score_a,score_b,played FROM matches WHERE evening_id=?`,
		eveningID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatches(rows)
}

func (r *MatchRepo) FindByPlayer(ctx context.Context, playerID domain.PlayerID) ([]domain.Match, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id,evening_id,player_a,player_b,score_a,score_b,played FROM matches
         WHERE player_a=? OR player_b=?`,
		playerID.String(), playerID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMatches(rows)
}

func (r *MatchRepo) UpdateScore(ctx context.Context, id domain.MatchID, scoreA, scoreB int) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE matches SET score_a=?, score_b=?, played=1 WHERE id=?`,
		scoreA, scoreB, id.String())
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
	if err := s.Scan(&idStr, &eveningStr, &paStr, &pbStr, &m.ScoreA, &m.ScoreB, &played); err != nil {
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
	var out []domain.Match
	for rows.Next() {
		m, err := scanMatch(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
