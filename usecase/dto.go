package usecase

import (
	"time"

	"DartScheduler/domain"

	"github.com/google/uuid"
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
	BuddyNr     string // member nr of the buddy player (from the "samen" column in the Excel import)
}

type BuddyPairInput struct {
	PlayerID domain.PlayerID
	BuddyID  domain.PlayerID
}

// --- Schedule ---

type GenerateScheduleInput struct {
	CompetitionName string
	Season          string
	NumEvenings     int // total slot count including catch-up and skipped slots
	StartDate       time.Time
	IntervalDays    int
	CatchUpNrs      []int // slot numbers that become catch-up evenings (no pre-assigned matches)
	SkipNrs         []int // slot numbers that are entirely skipped (no evening created)
	// ProgressFn is called during annealing with (step, total) to report progress.
	// If nil, no progress is reported.
	ProgressFn func(step, total int)
	// PlayerListID selects which named player list to use for generation.
	// If nil, all players (FindAll) are used — backwards compatible.
	PlayerListID *uuid.UUID
}

// --- Score ---

type SubmitScoreInput struct {
	MatchID              domain.MatchID
	Leg1Winner           string // player ID string
	Leg1Turns            int
	Leg2Winner           string
	Leg2Turns            int
	Leg3Winner           string
	Leg3Turns            int
	ReportedBy           string
	RescheduleDate       string
	SecretaryNr          string
	CounterNr            string
	PlayedDate           string
	PlayerA180s          int
	PlayerB180s          int
	PlayerAHighestFinish int
	PlayerBHighestFinish int
}

// --- Stats ---

type PlayerStats struct {
	Player          domain.Player `json:"player"`
	Played          int           `json:"played"`
	Wins            int           `json:"wins"`
	Losses          int           `json:"losses"`
	Draws           int           `json:"draws"`
	PointsFor       int           `json:"pointsFor"`
	PointsAgainst   int           `json:"pointsAgainst"`
	OneEighties     int           `json:"oneEighties"`
	HighestFinish   int           `json:"highestFinish"`
	MinTurns        int           `json:"minTurns"`
	AvgTurns        float64       `json:"avgTurns"`
	AvgScorePerTurn float64       `json:"avgScorePerTurn"`
}

// PlayedMatchItem is a flat row for the "all played matches" overview.
type PlayedMatchItem struct {
	MatchID     string `json:"matchId"`
	EveningNr   int    `json:"eveningNr"`
	EveningDate string `json:"eveningDate"` // yyyy-mm-dd
	PlayerANr   string `json:"playerANr"`
	PlayerAName string `json:"playerAName"`
	PlayerBNr   string `json:"playerBNr"`
	PlayerBName string `json:"playerBName"`
	ScoreA      int    `json:"scoreA"`
	ScoreB      int    `json:"scoreB"`
	Leg1Winner  string `json:"leg1Winner"` // "A", "B", or ""
	Leg1Turns   int    `json:"leg1Turns"`
	Leg2Winner  string `json:"leg2Winner"`
	Leg2Turns   int    `json:"leg2Turns"`
	Leg3Winner  string `json:"leg3Winner"`
	Leg3Turns   int    `json:"leg3Turns"`
	SecretaryNr string `json:"secretaryNr"`
	CounterNr   string `json:"counterNr"`
	PlayedDate  string `json:"playedDate"` // actual date for catch-up matches
}

// DutyMatch holds the evening number and match players for a single duty entry.
type DutyMatch struct {
	EveningNr   int    `json:"eveningNr"`
	PlayerANr   string `json:"playerANr"`
	PlayerAName string `json:"playerAName"`
	PlayerBNr   string `json:"playerBNr"`
	PlayerBName string `json:"playerBName"`
}

// DutyStats tracks how often a player has served as secretary or counter.
type DutyStats struct {
	Player           domain.Player `json:"player"`
	Count            int           `json:"count"`
	SecretaryCount   int           `json:"secretaryCount"`
	CounterCount     int           `json:"counterCount"`
	SecretaryMatches []DutyMatch   `json:"secretaryMatches"`
	CounterMatches   []DutyMatch   `json:"counterMatches"`
}

// SeasonSummary is a lightweight schedule list item.
type SeasonSummary struct {
	ID              string    `json:"id"`
	CompetitionName string    `json:"competitionName"`
	Season          string    `json:"season"`
	Active          bool      `json:"active"`
	CreatedAt       time.Time `json:"createdAt"`
	EveningCount    int       `json:"eveningCount"`
	PlayerListID    *string   `json:"playerListId,omitempty"`
}

// CatchUpEvening represents a catch-up evening with no pre-assigned matches.
type CatchUpEvening struct {
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
	ID    string `json:"id"`
	Nr    string `json:"nr"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Class string `json:"class"`
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

// PlayerListSummary is a lightweight player list item returned by the API.
type PlayerListSummary struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// --- Auth ---

// LoginOutput is returned by AuthUseCase.Login.
type LoginOutput struct {
	Token    string
	Username string
	Role     string
}

// CreateUserInput is the input for AuthUseCase.CreateUser.
type CreateUserInput struct {
	Username string
	Password string
	Role     string
}

// UpdateUserInput is the input for AuthUseCase.UpdateUser.
// Empty fields are ignored (no update performed for that field).
type UpdateUserInput struct {
	Role     string
	Password string
}

// UserDTO is the output representation of a User (never includes PasswordHash).
type UserDTO struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}
