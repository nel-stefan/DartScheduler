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
	SamenNr     string // member nr of the buddy to play together with (from "samen" column)
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

// --- Schedule Info ---

// ScheduleInfoResult is the response for the schedule info endpoint.
type ScheduleInfoResult struct {
	Players    []PlayerInfoItem  `json:"players"`
	Evenings   []EveningInfoItem `json:"evenings"`
	Matrix     []MatrixCellItem  `json:"matrix"`
	BuddyPairs []BuddyPairItem   `json:"buddyPairs"`
}

type PlayerInfoItem struct {
	ID   string `json:"id"`
	Nr   string `json:"nr"`
	Name string `json:"name"`
}

type EveningInfoItem struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	Date   string `json:"date"`
}

// MatrixCellItem represents how many matches a player is scheduled for on a given evening.
type MatrixCellItem struct {
	PlayerID  string `json:"playerId"`
	EveningID string `json:"eveningId"`
	Count     int    `json:"count"`
}

// BuddyPairItem represents a buddy pair and the evenings they share.
type BuddyPairItem struct {
	PlayerAID   string   `json:"playerAId"`
	PlayerANr   string   `json:"playerANr"`
	PlayerAName string   `json:"playerAName"`
	PlayerBID   string   `json:"playerBId"`
	PlayerBNr   string   `json:"playerBNr"`
	PlayerBName string   `json:"playerBName"`
	EveningIDs  []string `json:"eveningIds"`
	EveningNrs  []int    `json:"eveningNrs"`
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
