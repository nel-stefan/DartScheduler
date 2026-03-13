package domain

import "github.com/google/uuid"

type MatchID = uuid.UUID

type Match struct {
	ID        MatchID
	EveningID EveningID
	PlayerA   PlayerID
	PlayerB   PlayerID
	ScoreA    *int
	ScoreB    *int
	Played    bool
}
