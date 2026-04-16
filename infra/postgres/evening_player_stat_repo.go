package postgres

import (
	"context"
	"fmt"

	"DartScheduler/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EveningPlayerStatRepo implements domain.EveningPlayerStatRepository using PostgreSQL.
type EveningPlayerStatRepo struct{ pool *pgxpool.Pool }

func NewEveningPlayerStatRepo(pool *pgxpool.Pool) *EveningPlayerStatRepo {
	return &EveningPlayerStatRepo{pool: pool}
}

func (r *EveningPlayerStatRepo) FindByEvening(ctx context.Context, eveningID domain.EveningID) ([]domain.EveningPlayerStat, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT player_id, one_eighties, highest_finish
		 FROM evening_player_stats WHERE evening_id = $1`, eveningID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.EveningPlayerStat
	for rows.Next() {
		var s domain.EveningPlayerStat
		var playerID uuid.UUID
		if err := rows.Scan(&playerID, &s.OneEighties, &s.HighestFinish); err != nil {
			return nil, fmt.Errorf("scan evening_player_stat: %w", err)
		}
		s.EveningID = eveningID
		s.PlayerID = domain.PlayerID(playerID)
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *EveningPlayerStatRepo) Upsert(ctx context.Context, stat domain.EveningPlayerStat) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO evening_player_stats(evening_id, player_id, one_eighties, highest_finish)
		 VALUES($1, $2, $3, $4)
		 ON CONFLICT(evening_id, player_id) DO UPDATE SET
		   one_eighties   = EXCLUDED.one_eighties,
		   highest_finish = EXCLUDED.highest_finish`,
		stat.EveningID, stat.PlayerID, stat.OneEighties, stat.HighestFinish)
	return err
}
