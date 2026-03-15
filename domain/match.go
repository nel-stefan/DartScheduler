package domain

import "github.com/google/uuid"

// MatchID is een UUID die een wedstrijd uniek identificeert.
type MatchID = uuid.UUID

// Match stelt één wedstrijd voor tussen twee spelers op een speelavond.
// ScoreA en ScoreB zijn nil totdat de wedstrijd is gespeeld.
// Played wordt true nadat UpdateResult succesvol is aangeroepen.
type Match struct {
	ID        MatchID   `json:"id"`
	EveningID EveningID `json:"eveningId"`
	PlayerA   PlayerID  `json:"playerA"`
	PlayerB   PlayerID  `json:"playerB"`
	ScoreA    *int      `json:"scoreA"`
	ScoreB    *int      `json:"scoreB"`
	Played    bool      `json:"played"`
	// Leg detail fields (empty string = leg not played)
	Leg1Winner string `json:"leg1Winner"`
	Leg1Turns  int    `json:"leg1Turns"`
	Leg2Winner string `json:"leg2Winner"`
	Leg2Turns  int    `json:"leg2Turns"`
	Leg3Winner string `json:"leg3Winner"`
	Leg3Turns  int    `json:"leg3Turns"`
	// Administrative
	ReportedBy     string `json:"reportedBy"`
	RescheduleDate string `json:"rescheduleDate"`
	SecretaryNr    string `json:"secretaryNr"`
	CounterNr      string `json:"counterNr"`
}
