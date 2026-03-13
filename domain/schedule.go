package domain

import (
	"time"

	"github.com/google/uuid"
)

type ScheduleID = uuid.UUID

type Schedule struct {
	ID              ScheduleID
	CompetitionName string
	Evenings        []Evening
	CreatedAt       time.Time
}
