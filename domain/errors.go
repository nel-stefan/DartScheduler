package domain

import "errors"

// Sentinel errors recognised by use cases and handlers via errors.Is.
var (
	// ErrNotFound is returned when a requested entity does not exist.
	ErrNotFound = errors.New("not found")

	// ErrAlreadyExists is returned when attempting to create an entity that already exists.
	ErrAlreadyExists = errors.New("already exists")

	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")

	// ErrScheduleConflict is returned when a schedule operation cannot proceed
	// due to conflicting data.
	ErrScheduleConflict = errors.New("schedule conflict")

	// ErrMatchAlreadyPlayed is returned when a score is submitted for a match
	// that has already been played.
	ErrMatchAlreadyPlayed = errors.New("match already played")

	// ErrScheduleConstraintViolation is returned when a generated schedule violates
	// one or more hard constraints and cannot be used.
	ErrScheduleConstraintViolation = errors.New("schema voldoet niet aan harde constraints")
)
