package domain

import "context"

// PlayerRepository defines persistence operations for players and buddy preferences.
// Implementations are in infra/sqlite.
type PlayerRepository interface {
	Save(ctx context.Context, p Player) error
	SaveBatch(ctx context.Context, players []Player) error
	FindByID(ctx context.Context, id PlayerID) (Player, error)
	FindAll(ctx context.Context) ([]Player, error)
	Delete(ctx context.Context, id PlayerID) error
	DeleteAll(ctx context.Context) error
	SaveBuddyPreference(ctx context.Context, bp BuddyPreference) error
	FindBuddiesForPlayer(ctx context.Context, id PlayerID) ([]PlayerID, error)
	FindAllBuddyPairs(ctx context.Context) ([]BuddyPreference, error)
	DeleteBuddiesForPlayer(ctx context.Context, id PlayerID) error
	DeleteAllBuddyPairs(ctx context.Context) error
}

// EveningRepository defines persistence operations for playing evenings.
type EveningRepository interface {
	Save(ctx context.Context, e Evening, scheduleID ScheduleID) error
	FindByID(ctx context.Context, id EveningID) (Evening, error)
	FindBySchedule(ctx context.Context, scheduleID ScheduleID) ([]Evening, error)
	Delete(ctx context.Context, id EveningID) error
	DeleteBySchedule(ctx context.Context, scheduleID ScheduleID) error
}

// MatchRepository defines persistence operations for matches and scores.
type MatchRepository interface {
	Save(ctx context.Context, m Match) error
	SaveBatch(ctx context.Context, matches []Match) error
	FindByID(ctx context.Context, id MatchID) (Match, error)
	FindByEvening(ctx context.Context, eveningID EveningID) ([]Match, error)
	FindByPlayer(ctx context.Context, playerID PlayerID) ([]Match, error)
	FindByPlayerAndSchedule(ctx context.Context, playerID PlayerID, scheduleID ScheduleID) ([]Match, error)
	FindAllPlayed(ctx context.Context) ([]Match, error)
	UpdateResult(ctx context.Context, m Match) error
	// FindBySchedule returns all matches for a schedule in a single query.
	FindBySchedule(ctx context.Context, scheduleID ScheduleID) ([]Match, error)
	// FindCancelledBySchedule returns all matches with a non-empty ReportedBy
	// from non-inhaal evenings in the given schedule. Used to populate inhaalavonden.
	FindCancelledBySchedule(ctx context.Context, scheduleID ScheduleID) ([]Match, error)
	DeleteByEvening(ctx context.Context, eveningID EveningID) error
	DeleteBySchedule(ctx context.Context, scheduleID ScheduleID) error
	DeleteByPlayer(ctx context.Context, playerID PlayerID) error
}

// ScheduleRepository defines persistence operations for schedules.
// FindLatest returns the most recently created schedule.
type ScheduleRepository interface {
	Save(ctx context.Context, s Schedule) error
	FindLatest(ctx context.Context) (Schedule, error)
	FindByID(ctx context.Context, id ScheduleID) (Schedule, error)
	FindAll(ctx context.Context) ([]Schedule, error)
	Delete(ctx context.Context, id ScheduleID) error
	SetActive(ctx context.Context, id ScheduleID) error
}
