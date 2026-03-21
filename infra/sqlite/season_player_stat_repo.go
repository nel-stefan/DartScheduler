package sqlite

import (
	"context"
	"database/sql"

	"DartScheduler/domain"

	"github.com/google/uuid"
)

type SeasonPlayerStatRepo struct{ db *sql.DB }

func NewSeasonPlayerStatRepo(db *sql.DB) *SeasonPlayerStatRepo {
	return &SeasonPlayerStatRepo{db: db}
}

func (r *SeasonPlayerStatRepo) FindBySchedule(ctx context.Context, scheduleID domain.ScheduleID) ([]domain.SeasonPlayerStat, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT player_id, one_eighties, highest_finish
		 FROM season_player_stats WHERE schedule_id = ?`,
		scheduleID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.SeasonPlayerStat
	for rows.Next() {
		var playerIDStr string
		var s domain.SeasonPlayerStat
		if err := rows.Scan(&playerIDStr, &s.OneEighties, &s.HighestFinish); err != nil {
			return nil, err
		}
		pid, err := uuid.Parse(playerIDStr)
		if err != nil {
			return nil, err
		}
		s.ScheduleID = scheduleID
		s.PlayerID = domain.PlayerID(pid)
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SeasonPlayerStatRepo) Upsert(ctx context.Context, stat domain.SeasonPlayerStat) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO season_player_stats (schedule_id, player_id, one_eighties, highest_finish)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(schedule_id, player_id) DO UPDATE SET
		   one_eighties   = excluded.one_eighties,
		   highest_finish = excluded.highest_finish`,
		stat.ScheduleID.String(), stat.PlayerID.String(), stat.OneEighties, stat.HighestFinish)
	return err
}
