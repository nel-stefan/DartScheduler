package domain

import "github.com/google/uuid"

type PlayerID = uuid.UUID

type Player struct {
	ID      PlayerID
	Name    string
	Email   string
	Sponsor string
}

type BuddyPreference struct {
	PlayerID PlayerID
	BuddyID  PlayerID
}
