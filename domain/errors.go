package domain

import "errors"

// Schildwachtfouten die door use cases en handlers worden herkend via errors.Is.
var (
	// ErrNotFound wordt teruggegeven wanneer een gevraagde entiteit niet bestaat.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists wordt teruggegeven bij een poging een entiteit aan te maken die al bestaat.
	ErrAlreadyExists = errors.New("already exists")

	// ErrInvalidInput wordt teruggegeven wanneer invoervalidatie mislukt.
	ErrInvalidInput = errors.New("invalid input")

	// ErrScheduleConflict wordt teruggegeven wanneer een schema-operatie niet mogelijk is
	// vanwege conflicterende gegevens.
	ErrScheduleConflict = errors.New("schedule conflict")

	// ErrMatchAlreadyPlayed wordt teruggegeven wanneer een score wordt ingevoerd
	// voor een wedstrijd die al is gespeeld.
	ErrMatchAlreadyPlayed = errors.New("match already played")
)
