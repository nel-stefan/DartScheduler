package domain

import "context"

// EveningPlayerStat records per-player stats (180s, highest finish) for a single evening.
type EveningPlayerStat struct {
	EveningID     EveningID
	PlayerID      PlayerID
	OneEighties   int
	HighestFinish int
}

type EveningPlayerStatRepository interface {
	FindByEvening(ctx context.Context, eveningID EveningID) ([]EveningPlayerStat, error)
	Upsert(ctx context.Context, stat EveningPlayerStat) error
}
