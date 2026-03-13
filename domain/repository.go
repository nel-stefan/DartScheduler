package domain

import "context"

type PlayerRepository interface {
	Save(ctx context.Context, p Player) error
	SaveBatch(ctx context.Context, players []Player) error
	FindByID(ctx context.Context, id PlayerID) (Player, error)
	FindAll(ctx context.Context) ([]Player, error)
	DeleteAll(ctx context.Context) error
	SaveBuddyPreference(ctx context.Context, bp BuddyPreference) error
	FindBuddiesForPlayer(ctx context.Context, id PlayerID) ([]PlayerID, error)
	FindAllBuddyPairs(ctx context.Context) ([]BuddyPreference, error)
}

type EveningRepository interface {
	Save(ctx context.Context, e Evening, scheduleID ScheduleID) error
	FindByID(ctx context.Context, id EveningID) (Evening, error)
	FindBySchedule(ctx context.Context, scheduleID ScheduleID) ([]Evening, error)
}

type MatchRepository interface {
	Save(ctx context.Context, m Match) error
	SaveBatch(ctx context.Context, matches []Match) error
	FindByID(ctx context.Context, id MatchID) (Match, error)
	FindByEvening(ctx context.Context, eveningID EveningID) ([]Match, error)
	FindByPlayer(ctx context.Context, playerID PlayerID) ([]Match, error)
	UpdateScore(ctx context.Context, id MatchID, scoreA, scoreB int) error
}

type ScheduleRepository interface {
	Save(ctx context.Context, s Schedule) error
	FindLatest(ctx context.Context) (Schedule, error)
	FindByID(ctx context.Context, id ScheduleID) (Schedule, error)
}
