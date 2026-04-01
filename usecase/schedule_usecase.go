package usecase

import (
	"context"
	"fmt"
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
	log.Printf("[Generate] competition=%q season=%q numEvenings=%d startDate=%s intervalDays=%d catchUpNrs=%v skipNrs=%v",
		in.CompetitionName, in.Season, in.NumEvenings, in.StartDate.Format("2006-01-02"),
		in.IntervalDays, in.CatchUpNrs, in.SkipNrs)
	allPlayers, err := uc.players.FindAll(ctx)
	if err != nil {
		return domain.Schedule{}, fmt.Errorf("load players: %w", err)
	}

	// Sponsors (nr contains "-s") are excluded from the playing schedule.
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
		return domain.Schedule{}, fmt.Errorf("load buddy pairs: %w", err)
	}
	log.Printf("[Generate] buddyPairs=%d", len(buddyPairs))

	// Distribute slots into regular, catch-up, and skipped categories.
	catchUpSet := toIntSet(in.CatchUpNrs)
	skipSet := toIntSet(in.SkipNrs)

	type slotInfo struct {
		nr   int
		date time.Time
	}
	var regularSlots, catchUpSlots []slotInfo
	for i := range in.NumEvenings {
		nr := i + 1
		date := in.StartDate.AddDate(0, 0, i*in.IntervalDays)
		switch {
		case catchUpSet[nr]:
			catchUpSlots = append(catchUpSlots, slotInfo{nr, date})
		case skipSet[nr]:
			// skipped slot: date is advanced but no evening is created
		default:
			regularSlots = append(regularSlots, slotInfo{nr, date})
		}
	}

	regularDates := make([]time.Time, len(regularSlots))
	for i, s := range regularSlots {
		regularDates[i] = s.date
	}

	log.Printf("[Generate] slots: regular=%d catchUp=%d skip=%d", len(regularSlots), len(catchUpSlots), len(in.SkipNrs))
	annealCfg := scheduler.DefaultAnnealConfig()
	if in.ProgressFn != nil {
		annealCfg.ProgressFn = in.ProgressFn
	}
	sched, err := scheduler.Generate(scheduler.Input{
		Players:         participants,
		BuddyPairs:      buddyPairs,
		NumEvenings:     len(regularSlots),
		EveningDates:    regularDates,
		CompetitionName: in.CompetitionName,
		Config:          annealCfg,
	})
	if err != nil {
		return domain.Schedule{}, fmt.Errorf("generate schedule: %w", err)
	}
	log.Printf("[Generate] scheduler done: scheduleID=%s evenings=%d", sched.ID, len(sched.Evenings))

	// Re-number evenings to their original slot numbers.
	for i := range sched.Evenings {
		sched.Evenings[i].Number = regularSlots[i].nr
	}

	sched.Season = in.Season
	if err := uc.schedules.Save(ctx, sched); err != nil {
		return domain.Schedule{}, fmt.Errorf("save schedule: %w", err)
	}
	totalMatches := 0
	for _, ev := range sched.Evenings {
		if err := uc.evenings.Save(ctx, ev, sched.ID); err != nil {
			return domain.Schedule{}, fmt.Errorf("save evening %d: %w", ev.Number, err)
		}
		if len(ev.Matches) > 0 {
			if err := uc.matches.SaveBatch(ctx, ev.Matches); err != nil {
				return domain.Schedule{}, fmt.Errorf("save matches for evening %d: %w", ev.Number, err)
			}
			totalMatches += len(ev.Matches)
		}
	}
	log.Printf("[Generate] saved %d regular evenings, %d matches", len(sched.Evenings), totalMatches)

	// Save catch-up evenings (no pre-assigned matches).
	for _, s := range catchUpSlots {
		ev := domain.Evening{
			ID:               uuid.New(),
			Number:           s.nr,
			Date:             s.date,
			IsCatchUpEvening: true,
		}
		if err := uc.evenings.Save(ctx, ev, sched.ID); err != nil {
			return domain.Schedule{}, fmt.Errorf("save catch-up evening %d: %w", s.nr, err)
		}
	}
	log.Printf("[Generate] saved %d catch-up evenings", len(catchUpSlots))

	return sched, nil
}

// Regenerate reruns the scheduler for an existing schedule, replacing all match
// assignments while keeping the schedule metadata and evening structure intact.
// Only regular (non-catch-up) evenings are affected; their existing IDs and dates
// are reused so no downstream references break.
func (uc *ScheduleUseCase) Regenerate(ctx context.Context, id domain.ScheduleID, progressFn func(step, total int)) (domain.Schedule, error) {
	log.Printf("[Regenerate] scheduleID=%s", id)
	sched, err := uc.schedules.FindByID(ctx, id)
	if err != nil {
		return domain.Schedule{}, fmt.Errorf("load schedule: %w", err)
	}

	evenings, err := uc.evenings.FindBySchedule(ctx, id)
	if err != nil {
		return domain.Schedule{}, fmt.Errorf("load evenings: %w", err)
	}

	// Keep only regular evenings (catch-up evenings have no pre-assigned matches).
	var regularEvenings []domain.Evening
	for _, ev := range evenings {
		if !ev.IsCatchUpEvening {
			regularEvenings = append(regularEvenings, ev)
		}
	}
	sort.Slice(regularEvenings, func(i, j int) bool {
		return regularEvenings[i].Number < regularEvenings[j].Number
	})

	allPlayers, err := uc.players.FindAll(ctx)
	if err != nil {
		return domain.Schedule{}, fmt.Errorf("load players: %w", err)
	}
	participants := make([]domain.Player, 0, len(allPlayers))
	for _, p := range allPlayers {
		if !strings.Contains(strings.ToLower(p.Nr), "-s") {
			participants = append(participants, p)
		}
	}

	buddyPairs, err := uc.players.FindAllBuddyPairs(ctx)
	if err != nil {
		return domain.Schedule{}, fmt.Errorf("load buddy pairs: %w", err)
	}

	regularDates := make([]time.Time, len(regularEvenings))
	for i, ev := range regularEvenings {
		regularDates[i] = ev.Date
	}

	log.Printf("[Regenerate] running scheduler: participants=%d regularEvenings=%d buddyPairs=%d",
		len(participants), len(regularEvenings), len(buddyPairs))
	annealCfg := scheduler.DefaultAnnealConfig()
	if progressFn != nil {
		annealCfg.ProgressFn = progressFn
	}
	newSched, err := scheduler.Generate(scheduler.Input{
		Players:         participants,
		BuddyPairs:      buddyPairs,
		NumEvenings:     len(regularEvenings),
		EveningDates:    regularDates,
		CompetitionName: sched.CompetitionName,
		Config:          annealCfg,
	})
	if err != nil {
		return domain.Schedule{}, fmt.Errorf("generate schedule: %w", err)
	}

	// Delete existing matches for all regular evenings.
	for _, ev := range regularEvenings {
		if err := uc.matches.DeleteByEvening(ctx, ev.ID); err != nil {
			return domain.Schedule{}, fmt.Errorf("delete matches for evening %d: %w", ev.Number, err)
		}
	}

	// Save new matches under the existing evening IDs (newSched.Evenings[i] → regularEvenings[i]).
	totalMatches := 0
	for i, newEv := range newSched.Evenings {
		oldEvID := regularEvenings[i].ID
		for j := range newEv.Matches {
			newEv.Matches[j].ID = uuid.New()
			newEv.Matches[j].EveningID = oldEvID
		}
		if len(newEv.Matches) > 0 {
			if err := uc.matches.SaveBatch(ctx, newEv.Matches); err != nil {
				return domain.Schedule{}, fmt.Errorf("save matches for evening %d: %w", regularEvenings[i].Number, err)
			}
			totalMatches += len(newEv.Matches)
		}
	}
	log.Printf("[Regenerate] done: scheduleID=%s matches=%d", id, totalMatches)
	return uc.hydrate(ctx, sched)
}

func (uc *ScheduleUseCase) RenameSchedule(ctx context.Context, id domain.ScheduleID, name string) error {
	sched, err := uc.schedules.FindByID(ctx, id)
	if err != nil {
		return err
	}
	sched.CompetitionName = name
	return uc.schedules.Save(ctx, sched)
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

// ImportSeason creates a schedule from historically imported match rows and optional catch-up evenings.
// Players are matched by member number from the existing player list.
func (uc *ScheduleUseCase) ImportSeason(ctx context.Context, competitionName, season string, matchRows []SeasonMatchRow, catchUpEvenings []CatchUpEvening) (domain.Schedule, error) {
	log.Printf("[ImportSeason] competition=%q season=%q matchRows=%d catchUpEvenings=%d",
		competitionName, season, len(matchRows), len(catchUpEvenings))
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
				log.Printf("[ImportSeason] skipping match nrA=%q nrB=%q: player(s) not found (okA=%v okB=%v)",
					row.NrA, row.NrB, okA, okB)
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

	// Create catch-up evenings (no pre-assigned matches).
	for _, ie := range catchUpEvenings {
		ev := domain.Evening{
			ID:               uuid.New(),
			Number:           ie.EveningNr,
			Date:             ie.Date,
			IsCatchUpEvening: true,
		}
		if err := uc.evenings.Save(ctx, ev, sched.ID); err != nil {
			return domain.Schedule{}, err
		}
	}

	log.Printf("[ImportSeason] done scheduleID=%s", sched.ID)
	return uc.hydrate(ctx, sched)
}

// DeleteSchedule removes a schedule and all its evenings and matches.
func (uc *ScheduleUseCase) DeleteSchedule(ctx context.Context, id domain.ScheduleID) error {
	log.Printf("[DeleteSchedule] id=%s", id)
	if err := uc.matches.DeleteBySchedule(ctx, id); err != nil {
		return fmt.Errorf("delete matches: %w", err)
	}
	if err := uc.evenings.DeleteBySchedule(ctx, id); err != nil {
		return fmt.Errorf("delete evenings: %w", err)
	}
	if err := uc.schedules.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete schedule: %w", err)
	}
	log.Printf("[DeleteSchedule] done id=%s", id)
	return nil
}

func (uc *ScheduleUseCase) DeleteEvening(ctx context.Context, id domain.EveningID) error {
	log.Printf("[DeleteEvening] id=%s", id)
	if err := uc.matches.DeleteByEvening(ctx, id); err != nil {
		return fmt.Errorf("delete matches for evening: %w", err)
	}
	if err := uc.evenings.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete evening: %w", err)
	}
	log.Printf("[DeleteEvening] done id=%s", id)
	return nil
}

// AddCatchUpEvening appends a catch-up evening to the given schedule.
// No matches are pre-assigned; the evening dynamically shows all unplayed matches
// with a non-empty ReportedBy from earlier evenings (see hydrate).
func (uc *ScheduleUseCase) AddCatchUpEvening(ctx context.Context, scheduleID domain.ScheduleID, date time.Time) (domain.Schedule, error) {
	log.Printf("[AddCatchUpEvening] scheduleID=%s date=%s", scheduleID, date.Format("2006-01-02"))
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
	catchUpEv := domain.Evening{
		ID:               uuid.New(),
		Number:           maxNr + 1,
		Date:             date,
		IsCatchUpEvening: true,
	}
	if err := uc.evenings.Save(ctx, catchUpEv, scheduleID); err != nil {
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

	// Only regular (non-catch-up) evenings have pre-assigned matches.
	var regularEvenings []domain.Evening
	for _, ev := range evenings {
		if !ev.IsCatchUpEvening {
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
			PlayerAName: domain.FormatDisplayName(pA.Name),
			PlayerBID:   pB.ID.String(),
			PlayerBNr:   pB.Nr,
			PlayerBName: domain.FormatDisplayName(pB.Name),
			EveningIDs:  ids,
			EveningNrs:  nrs,
		})
	}

	playerItems := make([]PlayerInfoItem, len(allPlayers))
	for i, p := range allPlayers {
		playerItems[i] = PlayerInfoItem{ID: p.ID.String(), Nr: p.Nr, Name: domain.FormatDisplayName(p.Name), Email: p.Email}
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
// It uses two queries total: one for all regular matches, one for cancelled
// matches used by catch-up evenings.
func (uc *ScheduleUseCase) hydrate(ctx context.Context, sched domain.Schedule) (domain.Schedule, error) {
	evenings, err := uc.evenings.FindBySchedule(ctx, sched.ID)
	if err != nil {
		return sched, err
	}

	// Fetch all regular matches for this schedule in a single query.
	allMatches, err := uc.matches.FindBySchedule(ctx, sched.ID)
	if err != nil {
		return sched, fmt.Errorf("load matches for schedule %s: %w", sched.ID, err)
	}

	// Group matches by evening_id.
	matchesByEvening := make(map[domain.EveningID][]domain.Match, len(evenings))
	for _, m := range allMatches {
		matchesByEvening[m.EveningID] = append(matchesByEvening[m.EveningID], m)
	}

	// Lazily load cancelled matches only once if there are catch-up evenings.
	var cancelledMatches []domain.Match
	cancelledLoaded := false

	for i, ev := range evenings {
		if ev.IsCatchUpEvening {
			if !cancelledLoaded {
				log.Printf("[hydrate] loading cancelled matches for schedule %s", sched.ID)
				cancelledMatches, err = uc.matches.FindCancelledBySchedule(ctx, sched.ID)
				if err != nil {
					return sched, fmt.Errorf("load cancelled matches for schedule %s: %w", sched.ID, err)
				}
				log.Printf("[hydrate] %d cancelled matches found", len(cancelledMatches))
				cancelledLoaded = true
			}
			evenings[i].Matches = cancelledMatches
		} else {
			evenings[i].Matches = matchesByEvening[ev.ID]
		}
	}
	sched.Evenings = evenings
	return sched, nil
}
