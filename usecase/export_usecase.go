package usecase

import (
	"context"
	"io"

	"DartScheduler/domain"
)

// Exporter abstracts both Excel and PDF exporting.
type Exporter interface {
	Export(ctx context.Context, sched domain.Schedule, players []domain.Player, w io.Writer) error
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
	for i, ev := range evenings {
		ms, err := uc.matches.FindByEvening(ctx, ev.ID)
		if err != nil {
			return err
		}
		evenings[i].Matches = ms
	}
	sched.Evenings = evenings

	players, err := uc.players.FindAll(ctx)
	if err != nil {
		return err
	}

	return exp.Export(ctx, sched, players, w)
}
