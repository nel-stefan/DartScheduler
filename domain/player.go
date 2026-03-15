// Package domain bevat de kerndomeintypes en repository-interfaces van DartScheduler.
// Dit pakket heeft geen afhankelijkheden op infrastructuur of use-case pakketten;
// alle andere lagen mogen alleen inwaarts verwijzen naar domain.
package domain

import "github.com/google/uuid"

// PlayerID is een UUID die een speler uniek identificeert.
type PlayerID = uuid.UUID

// Player stelt een lid van de dartclub voor.
type Player struct {
	ID          PlayerID `json:"id"`
	Nr          string   `json:"nr"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Sponsor     string   `json:"sponsor"`
	Address     string   `json:"address"`
	PostalCode  string   `json:"postalCode"`
	City        string   `json:"city"`
	Phone       string   `json:"phone"`
	Mobile      string   `json:"mobile"`
	MemberSince string   `json:"memberSince"`
	Class       string   `json:"class"`
}

// BuddyPreference geeft aan dat twee spelers bij voorkeur op dezelfde avond spelen.
// De scheduler probeert buddy-koppels op dezelfde avond in te plannen.
type BuddyPreference struct {
	PlayerID PlayerID
	BuddyID  PlayerID
}
