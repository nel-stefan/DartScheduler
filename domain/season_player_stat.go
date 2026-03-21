package domain

import "context"

// SeasonPlayerStat records per-player stats (180s, highest finish) for an entire season.
type SeasonPlayerStat struct {
	ScheduleID    ScheduleID
	PlayerID      PlayerID
	OneEighties   int
	HighestFinish int
}

type SeasonPlayerStatRepository interface {
	FindBySchedule(ctx context.Context, scheduleID ScheduleID) ([]SeasonPlayerStat, error)
	Upsert(ctx context.Context, stat SeasonPlayerStat) error
}
