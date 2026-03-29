package usecase

import (
	"context"
	"io"
	"time"

	"DartScheduler/domain"
)

// Exporter abstracts both Excel and PDF exporting.
type Exporter interface {
	Export(ctx context.Context, sched domain.Schedule, players []domain.Player, w io.Writer) error
}

// EveningExporter abstracts single-evening Excel export.
type EveningExporter interface {
	ExportEvening(ctx context.Context, sched domain.Schedule, ev domain.Evening, players []domain.Player, w io.Writer) error
}

type ExportUseCase struct {
	schedules domain.ScheduleRepository
	evenings  domain.EveningRepository
	matches   domain.MatchRepository
	players   domain.PlayerRepository
}

func NewExportUseCase(
	schedules domain.ScheduleRepository,
	evenings domain.EveningRepository,
	matches domain.MatchRepository,
	players domain.PlayerRepository,
) *ExportUseCase {
	return &ExportUseCase{
		schedules: schedules,
		evenings:  evenings,
		matches:   matches,
		players:   players,
	}
}

func (uc *ExportUseCase) Export(ctx context.Context, exp Exporter, w io.Writer) error {
	sched, err := uc.schedules.FindLatest(ctx)
	if err != nil {
		return err
	}
	evenings, err := uc.evenings.FindBySchedule(ctx, sched.ID)
	if err != nil {
		return err
	}
	allMatches, err := uc.matches.FindBySchedule(ctx, sched.ID)
	if err != nil {
		return err
	}
	byEvening := make(map[domain.EveningID][]domain.Match, len(evenings))
	for _, m := range allMatches {
		byEvening[m.EveningID] = append(byEvening[m.EveningID], m)
	}
	for i, ev := range evenings {
		evenings[i].Matches = byEvening[ev.ID]
	}
	sched.Evenings = evenings

	players, err := uc.players.FindAll(ctx)
	if err != nil {
		return err
	}

	return exp.Export(ctx, sched, players, w)
}

// EveningDate returns the date of the evening with the given ID.
func (uc *ExportUseCase) EveningDate(ctx context.Context, eveningID domain.EveningID) (time.Time, error) {
	ev, err := uc.evenings.FindByID(ctx, eveningID)
	if err != nil {
		return time.Time{}, err
	}
	return ev.Date, nil
}

// ExportEvening exports a single evening's matches in wedstrijdformulier format.
func (uc *ExportUseCase) ExportEvening(ctx context.Context, exp EveningExporter, eveningID domain.EveningID, w io.Writer) error {
	sched, err := uc.schedules.FindLatest(ctx)
	if err != nil {
		return err
	}
	evenings, err := uc.evenings.FindBySchedule(ctx, sched.ID)
	if err != nil {
		return err
	}
	var targetEvening domain.Evening
	found := false
	for _, ev := range evenings {
		if ev.ID == eveningID {
			targetEvening = ev
			found = true
			break
		}
	}
	if !found {
		return domain.ErrNotFound
	}
	var matches []domain.Match
	if targetEvening.IsCatchUpEvening {
		matches, err = uc.matches.FindCancelledBySchedule(ctx, sched.ID)
	} else {
		matches, err = uc.matches.FindByEvening(ctx, targetEvening.ID)
	}
	if err != nil {
		return err
	}
	targetEvening.Matches = matches

	// For a regular evening, add one synthetic evening holding all cancelled
	// matches so the exporter can render them as an "Afgemeld" extra tab.
	if !targetEvening.IsCatchUpEvening {
		all, err := uc.matches.FindCancelledBySchedule(ctx, sched.ID)
		if err != nil {
			return err
		}
		// Exclude matches that were planned for this evening — those belong on
		// the main sheet and should not appear on the Afgemeld tab.
		var cancelled []domain.Match
		for _, m := range all {
			if m.EveningID != targetEvening.ID {
				cancelled = append(cancelled, m)
			}
		}
		if len(cancelled) > 0 {
			sched.Evenings = []domain.Evening{{IsCatchUpEvening: true, Matches: cancelled}}
		} else {
			sched.Evenings = nil
		}
	}

	players, err := uc.players.FindAll(ctx)
	if err != nil {
		return err
	}
	return exp.ExportEvening(ctx, sched, targetEvening, players, w)
}
