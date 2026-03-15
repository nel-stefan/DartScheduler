package domain

import "context"

// PlayerRepository definieert persistentieoperaties voor spelers en buddy-voorkeuren.
// Implementaties bevinden zich in infra/sqlite.
type PlayerRepository interface {
	Save(ctx context.Context, p Player) error
	SaveBatch(ctx context.Context, players []Player) error
	FindByID(ctx context.Context, id PlayerID) (Player, error)
	FindAll(ctx context.Context) ([]Player, error)
	DeleteAll(ctx context.Context) error
	SaveBuddyPreference(ctx context.Context, bp BuddyPreference) error
	FindBuddiesForPlayer(ctx context.Context, id PlayerID) ([]PlayerID, error)
	FindAllBuddyPairs(ctx context.Context) ([]BuddyPreference, error)
	DeleteBuddiesForPlayer(ctx context.Context, id PlayerID) error
}

// EveningRepository definieert persistentieoperaties voor speelavonden.
type EveningRepository interface {
	Save(ctx context.Context, e Evening, scheduleID ScheduleID) error
	FindByID(ctx context.Context, id EveningID) (Evening, error)
	FindBySchedule(ctx context.Context, scheduleID ScheduleID) ([]Evening, error)
}

// MatchRepository definieert persistentieoperaties voor wedstrijden en scores.
type MatchRepository interface {
	Save(ctx context.Context, m Match) error
	SaveBatch(ctx context.Context, matches []Match) error
	FindByID(ctx context.Context, id MatchID) (Match, error)
	FindByEvening(ctx context.Context, eveningID EveningID) ([]Match, error)
	FindByPlayer(ctx context.Context, playerID PlayerID) ([]Match, error)
	FindByPlayerAndSchedule(ctx context.Context, playerID PlayerID, scheduleID ScheduleID) ([]Match, error)
	FindAllPlayed(ctx context.Context) ([]Match, error)
	UpdateResult(ctx context.Context, m Match) error
	// FindCancelledBySchedule returns all matches with a non-empty ReportedBy
	// from non-inhaal evenings in the given schedule. Used to populate inhaalavonden.
	FindCancelledBySchedule(ctx context.Context, scheduleID ScheduleID) ([]Match, error)
}

// ScheduleRepository definieert persistentieoperaties voor schema's.
// FindLatest geeft het meest recent aangemaakte schema terug.
type ScheduleRepository interface {
	Save(ctx context.Context, s Schedule) error
	FindLatest(ctx context.Context) (Schedule, error)
	FindByID(ctx context.Context, id ScheduleID) (Schedule, error)
	FindAll(ctx context.Context) ([]Schedule, error)
}
