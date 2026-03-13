package domain

import (
	"time"

	"github.com/google/uuid"
)

type EveningID = uuid.UUID

type Evening struct {
	ID      EveningID
	Number  int
	Date    time.Time
	Matches []Match
}
