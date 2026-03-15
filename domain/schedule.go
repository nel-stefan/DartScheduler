package domain

import (
	"time"

	"github.com/google/uuid"
)

// ScheduleID uniquely identifies a competition schedule.
type ScheduleID = uuid.UUID

// Schedule represents a complete competition schedule with all playing evenings.
// Each competition has one active schedule; generating a new one replaces the previous.
type Schedule struct {
	ID              ScheduleID `json:"id"`
	CompetitionName string     `json:"competitionName"`
	Season          string     `json:"season"`
	Evenings        []Evening  `json:"evenings"`
	CreatedAt       time.Time  `json:"createdAt"`
}
