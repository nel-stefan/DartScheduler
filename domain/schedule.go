package domain

import (
	"time"

	"github.com/google/uuid"
)

// ScheduleID is een UUID die een competitieschema uniek identificeert.
type ScheduleID = uuid.UUID

// Schedule stelt een volledig competitieschema voor met alle speelavonden.
// Per competitie is er één actief schema; een nieuw gegenereerd schema vervangt het vorige.
type Schedule struct {
	ID              ScheduleID `json:"id"`
	CompetitionName string     `json:"competitionName"`
	Season          string     `json:"season"`
	Evenings        []Evening  `json:"evenings"`
	CreatedAt       time.Time  `json:"createdAt"`
}
