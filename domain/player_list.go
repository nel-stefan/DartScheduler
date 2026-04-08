package domain

import (
	"time"

	"github.com/google/uuid"
)

// PlayerListID uniquely identifies a named player list.
type PlayerListID = uuid.UUID

// PlayerList is a named set of players (e.g. "Ledenlijst 2026-2027").
type PlayerList struct {
	ID        PlayerListID
	Name      string
	CreatedAt time.Time
}
