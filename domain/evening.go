package domain

import (
	"time"

	"github.com/google/uuid"
)

// EveningID uniquely identifies a playing evening.
type EveningID = uuid.UUID

// Evening represents a single playing evening within a competition schedule.
// Number is the sequential index (1-based); Date is the scheduled playing date.
// IsCatchUpEvening indicates this is a catch-up evening: an evening with no
// pre-assigned matches where postponed games can be rescheduled.
type Evening struct {
	ID               EveningID `json:"id"`
	Number           int       `json:"number"`
	Date             time.Time `json:"date"`
	IsCatchUpEvening bool      `json:"isInhaalAvond"`
	Matches          []Match   `json:"matches"`
}
