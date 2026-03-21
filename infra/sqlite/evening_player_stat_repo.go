package sqlite

import (
	"context"
	"database/sql"

	"DartScheduler/domain"

	"github.com/google/uuid"
)

type EveningPlayerStatRepo struct{ db *sql.DB }

func NewEveningPlayerStatRepo(db *sql.DB) *EveningPlayerStatRepo {
	return &EveningPlayerStatRepo{db: db}
}

func (r *EveningPlayerStatRepo) FindByEvening(ctx context.Context, eveningID domain.EveningID) ([]domain.EveningPlayerStat, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT player_id, one_eighties, highest_finish
		   FROM evening_player_stats WHERE evening_id = ?`,
		eveningID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.EveningPlayerStat
	for rows.Next() {
		var s domain.EveningPlayerStat
		var playerIDStr string
		if err := rows.Scan(&playerIDStr, &s.OneEighties, &s.HighestFinish); err != nil {
			return nil, err
		}
		pid, err := uuid.Parse(playerIDStr)
		if err != nil {
			return nil, err
		}
		s.EveningID = eveningID
		s.PlayerID = domain.PlayerID(pid)
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *EveningPlayerStatRepo) Upsert(ctx context.Context, stat domain.EveningPlayerStat) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO evening_player_stats (evening_id, player_id, one_eighties, highest_finish)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(evening_id, player_id) DO UPDATE SET
		   one_eighties   = excluded.one_eighties,
		   highest_finish = excluded.highest_finish`,
		stat.EveningID.String(), stat.PlayerID.String(), stat.OneEighties, stat.HighestFinish)
	return err
}
