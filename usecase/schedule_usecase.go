package usecase

import (
	"context"
	"strings"
	"time"

	"DartScheduler/domain"
	"DartScheduler/scheduler"

	"github.com/google/uuid"
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

	// Sponsoren (nr bevat "-s") doen niet mee in het speelschema.
	participants := make([]domain.Player, 0, len(allPlayers))
	for _, p := range allPlayers {
		if !strings.Contains(strings.ToLower(p.Nr), "-s") {
			participants = append(participants, p)
		}
	}

	buddyPairs, err := uc.players.FindAllBuddyPairs(ctx)
	if err != nil {
		return domain.Schedule{}, err
	}

	sched, err := scheduler.Generate(scheduler.Input{
		Players:         participants,
		BuddyPairs:      buddyPairs,
		NumEvenings:     in.NumEvenings,
		CompetitionName: in.CompetitionName,
		StartDate:       in.StartDate,
		IntervalDays:    in.IntervalDays,
	})
	if err != nil {
		return domain.Schedule{}, err
	}

	sched.Season = in.Season
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

func (uc *ScheduleUseCase) ListSchedules(ctx context.Context) ([]SeasonSummary, error) {
	schedules, err := uc.schedules.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]SeasonSummary, len(schedules))
	for i, s := range schedules {
		evenings, err := uc.evenings.FindBySchedule(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		out[i] = SeasonSummary{
			ID:              s.ID.String(),
			CompetitionName: s.CompetitionName,
			Season:          s.Season,
			CreatedAt:       s.CreatedAt,
			EveningCount:    len(evenings),
		}
	}
	return out, nil
}

// ImportSeason creates a schedule from historically imported match rows and optional inhaalavonden.
// Players are matched by nr from the existing player list.
func (uc *ScheduleUseCase) ImportSeason(ctx context.Context, competitionName, season string, matchRows []SeasonMatchRow, inhaalEvenings []InhaalEvening) (domain.Schedule, error) {
	allPlayers, err := uc.players.FindAll(ctx)
	if err != nil {
		return domain.Schedule{}, err
	}
	// Build player lookup by nr
	byNr := make(map[string]domain.Player, len(allPlayers))
	for _, p := range allPlayers {
		byNr[p.Nr] = p
	}

	// Group rows by evening number
	eveningOrder := []int{}
	eveningDates := map[int]time.Time{}
	matchesByEvening := map[int][]SeasonMatchRow{}
	for _, row := range matchRows {
		if _, seen := eveningDates[row.EveningNr]; !seen {
			eveningOrder = append(eveningOrder, row.EveningNr)
			eveningDates[row.EveningNr] = row.Date
		}
		matchesByEvening[row.EveningNr] = append(matchesByEvening[row.EveningNr], row)
	}

	sched := domain.Schedule{
		ID:              uuid.New(),
		CompetitionName: competitionName,
		Season:          season,
		CreatedAt:       time.Now(),
	}
	if err := uc.schedules.Save(ctx, sched); err != nil {
		return domain.Schedule{}, err
	}

	for _, evNr := range eveningOrder {
		ev := domain.Evening{
			ID:     uuid.New(),
			Number: evNr,
			Date:   eveningDates[evNr],
		}
		if err := uc.evenings.Save(ctx, ev, sched.ID); err != nil {
			return domain.Schedule{}, err
		}

		var matches []domain.Match
		for _, row := range matchesByEvening[evNr] {
			pA, okA := byNr[row.NrA]
			pB, okB := byNr[row.NrB]
			if !okA || !okB {
				continue // skip if players not in system
			}
			scoreA, scoreB := row.ScoreA, row.ScoreB
			// derive score from legs if not provided
			if scoreA == 0 && scoreB == 0 && (row.Leg1Winner != "" || row.Leg2Winner != "") {
				for _, w := range []string{row.Leg1Winner, row.Leg2Winner, row.Leg3Winner} {
					if w == "" {
						continue
					}
					if w == pA.Nr || strings.EqualFold(w, pA.Name) {
						scoreA++
					} else {
						scoreB++
					}
				}
			}
			played := scoreA > 0 || scoreB > 0
			m := domain.Match{
				ID:          uuid.New(),
				EveningID:   ev.ID,
				PlayerA:     pA.ID,
				PlayerB:     pB.ID,
				Played:      played,
				Leg1Winner:  row.Leg1Winner,
				Leg1Turns:   row.Leg1Turns,
				Leg2Winner:  row.Leg2Winner,
				Leg2Turns:   row.Leg2Turns,
				Leg3Winner:  row.Leg3Winner,
				Leg3Turns:   row.Leg3Turns,
				SecretaryNr: row.Secretary,
				CounterNr:   row.Counter,
			}
			if played {
				m.ScoreA = &scoreA
				m.ScoreB = &scoreB
			}
			matches = append(matches, m)
		}
		if len(matches) > 0 {
			if err := uc.matches.SaveBatch(ctx, matches); err != nil {
				return domain.Schedule{}, err
			}
		}
	}

	// Create inhaalavonden (no pre-assigned matches).
	for _, ie := range inhaalEvenings {
		ev := domain.Evening{
			ID:            uuid.New(),
			Number:        ie.EveningNr,
			Date:          ie.Date,
			IsInhaalAvond: true,
		}
		if err := uc.evenings.Save(ctx, ev, sched.ID); err != nil {
			return domain.Schedule{}, err
		}
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
