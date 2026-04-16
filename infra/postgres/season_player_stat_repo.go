package postgres

import (
	"context"
	"fmt"

	"DartScheduler/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SeasonPlayerStatRepo implements domain.SeasonPlayerStatRepository using PostgreSQL.
type SeasonPlayerStatRepo struct{ pool *pgxpool.Pool }

func NewSeasonPlayerStatRepo(pool *pgxpool.Pool) *SeasonPlayerStatRepo {
	return &SeasonPlayerStatRepo{pool: pool}
}

func (r *SeasonPlayerStatRepo) FindBySchedule(ctx context.Context, scheduleID domain.ScheduleID) ([]domain.SeasonPlayerStat, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT player_id, one_eighties, highest_finish
		 FROM season_player_stats WHERE schedule_id = $1`, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.SeasonPlayerStat
	for rows.Next() {
		var s domain.SeasonPlayerStat
		var playerID uuid.UUID
		if err := rows.Scan(&playerID, &s.OneEighties, &s.HighestFinish); err != nil {
			return nil, fmt.Errorf("scan season_player_stat: %w", err)
		}
		s.ScheduleID = scheduleID
		s.PlayerID = domain.PlayerID(playerID)
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SeasonPlayerStatRepo) Upsert(ctx context.Context, stat domain.SeasonPlayerStat) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO season_player_stats(schedule_id, player_id, one_eighties, highest_finish)
		 VALUES($1, $2, $3, $4)
		 ON CONFLICT(schedule_id, player_id) DO UPDATE SET
		   one_eighties   = EXCLUDED.one_eighties,
		   highest_finish = EXCLUDED.highest_finish`,
		stat.ScheduleID, stat.PlayerID, stat.OneEighties, stat.HighestFinish)
	return err
}
