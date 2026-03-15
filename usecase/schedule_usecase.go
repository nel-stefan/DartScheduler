package usecase

import (
	"context"
	"log"
	"sort"
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
	log.Printf("[Generate] competition=%q season=%q numEvenings=%d startDate=%s intervalDays=%d inhaalNrs=%v vrijeNrs=%v",
		in.CompetitionName, in.Season, in.NumEvenings, in.StartDate.Format("2006-01-02"),
		in.IntervalDays, in.InhaalNrs, in.VrijeNrs)
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
	log.Printf("[Generate] players loaded: total=%d participants=%d (sponsors excluded)",
		len(allPlayers), len(participants))

	buddyPairs, err := uc.players.FindAllBuddyPairs(ctx)
	if err != nil {
		return domain.Schedule{}, err
	}
	log.Printf("[Generate] buddyPairs=%d", len(buddyPairs))

	// Verdeel de slots over normaal, inhaal en vrij.
	inhaalSet := toIntSet(in.InhaalNrs)
	vrijSet := toIntSet(in.VrijeNrs)

	type slotInfo struct {
		nr   int
		date time.Time
	}
	var regularSlots, inhaalSlots []slotInfo
	for i := range in.NumEvenings {
		nr := i + 1
		date := in.StartDate.AddDate(0, 0, i*in.IntervalDays)
		switch {
		case inhaalSet[nr]:
			inhaalSlots = append(inhaalSlots, slotInfo{nr, date})
		case vrijSet[nr]:
			// vrije avond: datum wordt overgeslagen, geen evening aangemaakt
		default:
			regularSlots = append(regularSlots, slotInfo{nr, date})
		}
	}

	regularDates := make([]time.Time, len(regularSlots))
	for i, s := range regularSlots {
		regularDates[i] = s.date
	}

	log.Printf("[Generate] slots: regular=%d inhaal=%d vrij=%d", len(regularSlots), len(inhaalSlots), len(in.VrijeNrs))
	sched, err := scheduler.Generate(scheduler.Input{
		Players:         participants,
		BuddyPairs:      buddyPairs,
		NumEvenings:     len(regularSlots),
		EveningDates:    regularDates,
		CompetitionName: in.CompetitionName,
	})
	if err != nil {
		return domain.Schedule{}, err
	}
	log.Printf("[Generate] scheduler done: scheduleID=%s evenings=%d", sched.ID, len(sched.Evenings))

	// Hernum evenings naar hun slot-nummer.
	for i := range sched.Evenings {
		sched.Evenings[i].Number = regularSlots[i].nr
	}

	sched.Season = in.Season
	if err := uc.schedules.Save(ctx, sched); err != nil {
		return domain.Schedule{}, err
	}
	totalMatches := 0
	for _, ev := range sched.Evenings {
		if err := uc.evenings.Save(ctx, ev, sched.ID); err != nil {
			return domain.Schedule{}, err
		}
		if len(ev.Matches) > 0 {
			if err := uc.matches.SaveBatch(ctx, ev.Matches); err != nil {
				return domain.Schedule{}, err
			}
			totalMatches += len(ev.Matches)
		}
	}
	log.Printf("[Generate] saved %d regular evenings, %d matches", len(sched.Evenings), totalMatches)

	// Sla inhaalavonden op.
	for _, s := range inhaalSlots {
		ev := domain.Evening{
			ID:            uuid.New(),
			Number:        s.nr,
			Date:          s.date,
			IsInhaalAvond: true,
		}
		if err := uc.evenings.Save(ctx, ev, sched.ID); err != nil {
			return domain.Schedule{}, err
		}
	}
	log.Printf("[Generate] saved %d inhaalavonden", len(inhaalSlots))

	return sched, nil
}

func toIntSet(nrs []int) map[int]bool {
	s := make(map[int]bool, len(nrs))
	for _, n := range nrs {
		s[n] = true
	}
	return s
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
	log.Printf("[ImportSeason] competition=%q season=%q matchRows=%d inhaalEvenings=%d",
		competitionName, season, len(matchRows), len(inhaalEvenings))
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

	log.Printf("[ImportSeason] done scheduleID=%s", sched.ID)
	return uc.hydrate(ctx, sched)
}

// AddInhaalAvond creates an inhaalavond for the given schedule and date.
// Matches are not moved — the inhaalavond dynamically shows all unplayed
// matches with a non-empty ReportedBy from earlier evenings (see hydrate).
func (uc *ScheduleUseCase) DeleteSchedule(ctx context.Context, id domain.ScheduleID) error {
	log.Printf("[DeleteSchedule] id=%s", id)
	if err := uc.matches.DeleteBySchedule(ctx, id); err != nil {
		return err
	}
	if err := uc.evenings.DeleteBySchedule(ctx, id); err != nil {
		return err
	}
	if err := uc.schedules.Delete(ctx, id); err != nil {
		return err
	}
	log.Printf("[DeleteSchedule] done id=%s", id)
	return nil
}

func (uc *ScheduleUseCase) DeleteEvening(ctx context.Context, id domain.EveningID) error {
	log.Printf("[DeleteEvening] id=%s", id)
	if err := uc.matches.DeleteByEvening(ctx, id); err != nil {
		return err
	}
	if err := uc.evenings.Delete(ctx, id); err != nil {
		return err
	}
	log.Printf("[DeleteEvening] done id=%s", id)
	return nil
}

func (uc *ScheduleUseCase) AddInhaalAvond(ctx context.Context, scheduleID domain.ScheduleID, date time.Time) (domain.Schedule, error) {
	log.Printf("[AddInhaalAvond] scheduleID=%s date=%s", scheduleID, date.Format("2006-01-02"))
	evenings, err := uc.evenings.FindBySchedule(ctx, scheduleID)
	if err != nil {
		return domain.Schedule{}, err
	}
	maxNr := 0
	for _, ev := range evenings {
		if ev.Number > maxNr {
			maxNr = ev.Number
		}
	}
	inhaalEv := domain.Evening{
		ID:            uuid.New(),
		Number:        maxNr + 1,
		Date:          date,
		IsInhaalAvond: true,
	}
	if err := uc.evenings.Save(ctx, inhaalEv, scheduleID); err != nil {
		return domain.Schedule{}, err
	}
	sched, err := uc.schedules.FindByID(ctx, scheduleID)
	if err != nil {
		return domain.Schedule{}, err
	}
	return uc.hydrate(ctx, sched)
}

// GetInfo returns analytics for a schedule: player×evening matrix, buddy pair shared evenings.
func (uc *ScheduleUseCase) GetInfo(ctx context.Context, scheduleID domain.ScheduleID) (ScheduleInfoResult, error) {
	log.Printf("[GetInfo] scheduleID=%s", scheduleID)
	allPlayers, err := uc.players.FindAll(ctx)
	if err != nil {
		return ScheduleInfoResult{}, err
	}
	playerMap := make(map[domain.PlayerID]domain.Player, len(allPlayers))
	for _, p := range allPlayers {
		playerMap[p.ID] = p
	}

	evenings, err := uc.evenings.FindBySchedule(ctx, scheduleID)
	if err != nil {
		return ScheduleInfoResult{}, err
	}

	// Only regular (non-inhaal) evenings have pre-assigned matches.
	var regularEvenings []domain.Evening
	for _, ev := range evenings {
		if !ev.IsInhaalAvond {
			regularEvenings = append(regularEvenings, ev)
		}
	}

	type cellKey struct {
		playerID  domain.PlayerID
		eveningID domain.EveningID
	}
	cellCount := make(map[cellKey]int)
	for _, ev := range regularEvenings {
		matches, err := uc.matches.FindByEvening(ctx, ev.ID)
		if err != nil {
			return ScheduleInfoResult{}, err
		}
		for _, m := range matches {
			cellCount[cellKey{m.PlayerA, ev.ID}]++
			cellCount[cellKey{m.PlayerB, ev.ID}]++
		}
	}

	matrix := make([]MatrixCellItem, 0, len(cellCount))
	for k, count := range cellCount {
		matrix = append(matrix, MatrixCellItem{
			PlayerID:  k.playerID.String(),
			EveningID: k.eveningID.String(),
			Count:     count,
		})
	}

	// Build set: player → evenings they play in.
	playerEvenings := make(map[domain.PlayerID]map[domain.EveningID]bool)
	for k := range cellCount {
		if playerEvenings[k.playerID] == nil {
			playerEvenings[k.playerID] = make(map[domain.EveningID]bool)
		}
		playerEvenings[k.playerID][k.eveningID] = true
	}

	// Evening number lookup for sorting.
	eveningNrMap := make(map[domain.EveningID]int, len(regularEvenings))
	for _, ev := range regularEvenings {
		eveningNrMap[ev.ID] = ev.Number
	}

	buddyPairs, err := uc.players.FindAllBuddyPairs(ctx)
	if err != nil {
		return ScheduleInfoResult{}, err
	}

	seenPairs := make(map[string]bool)
	buddyPairItems := make([]BuddyPairItem, 0)
	for _, bp := range buddyPairs {
		aID, bID := bp.PlayerID, bp.BuddyID
		key := aID.String() + ":" + bID.String()
		revKey := bID.String() + ":" + aID.String()
		if seenPairs[key] || seenPairs[revKey] {
			continue
		}
		seenPairs[key] = true

		pA, okA := playerMap[aID]
		pB, okB := playerMap[bID]
		if !okA || !okB {
			continue
		}

		var sharedEIDs []domain.EveningID
		for eid := range playerEvenings[aID] {
			if playerEvenings[bID][eid] {
				sharedEIDs = append(sharedEIDs, eid)
			}
		}
		sort.Slice(sharedEIDs, func(i, j int) bool {
			return eveningNrMap[sharedEIDs[i]] < eveningNrMap[sharedEIDs[j]]
		})

		ids := make([]string, len(sharedEIDs))
		nrs := make([]int, len(sharedEIDs))
		for i, eid := range sharedEIDs {
			ids[i] = eid.String()
			nrs[i] = eveningNrMap[eid]
		}

		buddyPairItems = append(buddyPairItems, BuddyPairItem{
			PlayerAID:   pA.ID.String(),
			PlayerANr:   pA.Nr,
			PlayerAName: pA.Name,
			PlayerBID:   pB.ID.String(),
			PlayerBNr:   pB.Nr,
			PlayerBName: pB.Name,
			EveningIDs:  ids,
			EveningNrs:  nrs,
		})
	}

	playerItems := make([]PlayerInfoItem, len(allPlayers))
	for i, p := range allPlayers {
		playerItems[i] = PlayerInfoItem{ID: p.ID.String(), Nr: p.Nr, Name: p.Name}
	}

	eveningItems := make([]EveningInfoItem, len(regularEvenings))
	for i, ev := range regularEvenings {
		eveningItems[i] = EveningInfoItem{
			ID:     ev.ID.String(),
			Number: ev.Number,
			Date:   ev.Date.Format("2006-01-02"),
		}
	}

	log.Printf("[GetInfo] scheduleID=%s players=%d evenings=%d matrixCells=%d buddyPairs=%d",
		scheduleID, len(playerItems), len(eveningItems), len(matrix), len(buddyPairItems))
	return ScheduleInfoResult{
		Players:    playerItems,
		Evenings:   eveningItems,
		Matrix:     matrix,
		BuddyPairs: buddyPairItems,
	}, nil
}

// hydrate loads evenings (with matches) into the schedule.
func (uc *ScheduleUseCase) hydrate(ctx context.Context, sched domain.Schedule) (domain.Schedule, error) {
	evenings, err := uc.evenings.FindBySchedule(ctx, sched.ID)
	if err != nil {
		return sched, err
	}
	for i, ev := range evenings {
		var matches []domain.Match
		if ev.IsInhaalAvond {
			log.Printf("[hydrate] inhaalavond ev=%s → querying cancelled matches", ev.ID)
			matches, err = uc.matches.FindCancelledBySchedule(ctx, sched.ID)
			log.Printf("[hydrate] inhaalavond ev=%s → %d cancelled matches found", ev.ID, len(matches))
		} else {
			matches, err = uc.matches.FindByEvening(ctx, ev.ID)
		}
		if err != nil {
			return sched, err
		}
		evenings[i].Matches = matches
	}
	sched.Evenings = evenings
	return sched, nil
}
