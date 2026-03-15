package usecase

import (
	"time"

	"DartScheduler/domain"
)

// --- Player ---

type PlayerInput struct {
	Nr          string
	Name        string
	Email       string
	Sponsor     string
	Address     string
	PostalCode  string
	City        string
	Phone       string
	Mobile      string
	MemberSince string
	Class       string
}

type BuddyPairInput struct {
	PlayerID domain.PlayerID
	BuddyID  domain.PlayerID
}

// --- Schedule ---

type GenerateScheduleInput struct {
	CompetitionName string
	Season          string
	NumEvenings     int // total slots including inhaal and vrij
	StartDate       time.Time
	IntervalDays    int
	InhaalNrs       []int // slot numbers that become inhaalavonden
	VrijeNrs        []int // slot numbers that are skipped (no evening created)
}

// --- Score ---

type SubmitScoreInput struct {
	MatchID        domain.MatchID
	Leg1Winner     string // player ID string
	Leg1Turns      int
	Leg2Winner     string
	Leg2Turns      int
	Leg3Winner     string
	Leg3Turns      int
	ReportedBy     string
	RescheduleDate string
	SecretaryNr    string
	CounterNr      string
}

// --- Stats ---

type PlayerStats struct {
	Player        domain.Player `json:"player"`
	Played        int           `json:"played"`
	Wins          int           `json:"wins"`
	Losses        int           `json:"losses"`
	Draws         int           `json:"draws"`
	PointsFor     int           `json:"pointsFor"`
	PointsAgainst int           `json:"pointsAgainst"`
}

// DutyStats tracks how often a player has served as secretary or counter.
type DutyStats struct {
	Player domain.Player `json:"player"`
	Count  int           `json:"count"`
}

// SeasonSummary is a lightweight schedule list item.
type SeasonSummary struct {
	ID              string    `json:"id"`
	CompetitionName string    `json:"competitionName"`
	Season          string    `json:"season"`
	CreatedAt       time.Time `json:"createdAt"`
	EveningCount    int       `json:"eveningCount"`
}

// InhaalEvening represents a catch-up evening with no pre-assigned matches.
type InhaalEvening struct {
	EveningNr int
	Date      time.Time
}

// SeasonMatchRow holds one imported match row from a historical season Excel.
type SeasonMatchRow struct {
	EveningNr  int
	Date       time.Time
	NrA        string
	NameA      string
	NrB        string
	NameB      string
	Leg1Winner string
	Leg1Turns  int
	Leg2Winner string
	Leg2Turns  int
	Leg3Winner string
	Leg3Turns  int
	ScoreA     int
	ScoreB     int
	Secretary  string
	Counter    string
}
