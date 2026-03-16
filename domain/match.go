package domain

import "github.com/google/uuid"

// MatchID uniquely identifies a match.
type MatchID = uuid.UUID

// Match represents a single match between two players on a playing evening.
// ScoreA and ScoreB are nil until the match has been played.
// Played becomes true after UpdateResult is called successfully.
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
	// Statistics (0 = not recorded)
	PlayerA180s          int `json:"playerA180s"`
	PlayerB180s          int `json:"playerB180s"`
	PlayerAHighestFinish int `json:"playerAHighestFinish"`
	PlayerBHighestFinish int `json:"playerBHighestFinish"`
}
