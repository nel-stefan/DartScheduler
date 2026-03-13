package usecase

import (
	"context"

	"DartScheduler/domain"
	"DartScheduler/scheduler"
)

type ScheduleUseCase struct {
	players   domain.PlayerRepository
	schedules domain.ScheduleRepository
	evenings  domain.EveningRepository
	matches   domain.MatchRepository
}

func NewScheduleUseCase(
	players domain.PlayerRepository,
	schedules domain.ScheduleRepository,
	evenings domain.EveningRepository,
	matches domain.MatchRepository,
) *ScheduleUseCase {
	return &ScheduleUseCase{
		players:   players,
		schedules: schedules,
		evenings:  evenings,
		matches:   matches,
	}
}

func (uc *ScheduleUseCase) Generate(ctx context.Context, in GenerateScheduleInput) (domain.Schedule, error) {
	allPlayers, err := uc.players.FindAll(ctx)
	if err != nil {
		return domain.Schedule{}, err
	}
	buddyPairs, err := uc.players.FindAllBuddyPairs(ctx)
	if err != nil {
		return domain.Schedule{}, err
	}

	sched, err := scheduler.Generate(scheduler.Input{
		Players:         allPlayers,
		BuddyPairs:      buddyPairs,
		NumEvenings:     in.NumEvenings,
		CompetitionName: in.CompetitionName,
		StartDate:       in.StartDate,
		IntervalDays:    in.IntervalDays,
	})
	if err != nil {
		return domain.Schedule{}, err
	}

	if err := uc.schedules.Save(ctx, sched); err != nil {
		return domain.Schedule{}, err
	}
	for _, ev := range sched.Evenings {
		if err := uc.evenings.Save(ctx, ev, sched.ID); err != nil {
			return domain.Schedule{}, err
		}
		if len(ev.Matches) > 0 {
			if err := uc.matches.SaveBatch(ctx, ev.Matches); err != nil {
				return domain.Schedule{}, err
			}
		}
	}
	return sched, nil
}

func (uc *ScheduleUseCase) GetLatest(ctx context.Context) (domain.Schedule, error) {
	sched, err := uc.schedules.FindLatest(ctx)
	if err != nil {
		return domain.Schedule{}, err
	}
	return uc.hydrate(ctx, sched)
}

func (uc *ScheduleUseCase) GetByID(ctx context.Context, id domain.ScheduleID) (domain.Schedule, error) {
	sched, err := uc.schedules.FindByID(ctx, id)
	if err != nil {
		return domain.Schedule{}, err
	}
	return uc.hydrate(ctx, sched)
}

// hydrate loads evenings (with matches) into the schedule.
func (uc *ScheduleUseCase) hydrate(ctx context.Context, sched domain.Schedule) (domain.Schedule, error) {
	evenings, err := uc.evenings.FindBySchedule(ctx, sched.ID)
	if err != nil {
		return sched, err
	}
	for i, ev := range evenings {
		matches, err := uc.matches.FindByEvening(ctx, ev.ID)
		if err != nil {
			return sched, err
		}
		evenings[i].Matches = matches
	}
	sched.Evenings = evenings
	return sched, nil
}
