package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrScheduleConflict  = errors.New("schedule conflict")
	ErrMatchAlreadyPlayed = errors.New("match already played")
)
