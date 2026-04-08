// Package domain contains the core domain types and repository interfaces for DartScheduler.
// This package has no dependencies on infrastructure or use-case packages;
// all other layers may only reference domain inward.
package domain

import (
	"strings"

	"github.com/google/uuid"
)

// FormatDisplayName converts a stored "Achternaam, Voornaam" name to "Voornaam Achternaam".
// Names without a ", " separator are returned unchanged.
func FormatDisplayName(name string) string {
	parts := strings.SplitN(name, ", ", 2)
	if len(parts) != 2 {
		return name
	}
	return strings.TrimSpace(parts[1]) + " " + strings.TrimSpace(parts[0])
}

// PlayerID uniquely identifies a player.
type PlayerID = uuid.UUID

// Player represents a member of the dart club.
type Player struct {
	ID          PlayerID   `json:"id"`
	Nr          string     `json:"nr"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Sponsor     string     `json:"sponsor"`
	Address     string     `json:"address"`
	PostalCode  string     `json:"postalCode"`
	City        string     `json:"city"`
	Phone       string     `json:"phone"`
	Mobile      string     `json:"mobile"`
	MemberSince string     `json:"memberSince"`
	Class       string     `json:"class"`
	ListID      *uuid.UUID `json:"listId,omitempty"`
}

// BuddyPreference indicates that two players prefer to play on the same evening.
// The scheduler attempts to assign buddy pairs to the same evenings.
type BuddyPreference struct {
	PlayerID PlayerID
	BuddyID  PlayerID
}
