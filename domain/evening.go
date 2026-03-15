package domain

import (
	"time"

	"github.com/google/uuid"
)

// EveningID is een UUID die een speelavond uniek identificeert.
type EveningID = uuid.UUID

// Evening stelt één speelavond voor binnen een competitieschema.
// Number is het volgnummer (1-gebaseerd); Date is de geplande speeldatum.
// IsInhaalAvond geeft aan dat het een inhaalavond is: een avond zonder vaste
// wedstrijden waarop uitgestelde partijen ingehaald kunnen worden.
type Evening struct {
	ID            EveningID `json:"id"`
	Number        int       `json:"number"`
	Date          time.Time `json:"date"`
	IsInhaalAvond bool      `json:"isInhaalAvond"`
	Matches       []Match   `json:"matches"`
}
