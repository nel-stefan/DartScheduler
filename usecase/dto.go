package usecase

import (
	"time"

	"DartScheduler/domain"
)

// --- Player ---

type PlayerInput struct {
	Name    string
	Email   string
	Sponsor string
}

type BuddyPairInput struct {
	PlayerID domain.PlayerID
	BuddyID  domain.PlayerID
}

// --- Schedule ---

type GenerateScheduleInput struct {
	CompetitionName string
	NumEvenings     int
	StartDate       time.Time
	IntervalDays    int
}

// --- Score ---

type SubmitScoreInput struct {
	MatchID domain.MatchID
	ScoreA  int
	ScoreB  int
}

// --- Stats ---

type PlayerStats struct {
	Player     domain.Player
	Played     int
	Wins       int
	Losses     int
	Draws      int
	PointsFor  int
	PointsAgainst int
}
