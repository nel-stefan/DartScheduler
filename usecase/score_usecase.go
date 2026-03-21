package usecase

import (
	"context"
	"log"
	"sort"

	"DartScheduler/domain"
)

type ScoreUseCase struct {
	matches      domain.MatchRepository
	evenings     domain.EveningRepository
	seasonStats  domain.SeasonPlayerStatRepository
}

func NewScoreUseCase(matches domain.MatchRepository, evenings domain.EveningRepository, seasonStats domain.SeasonPlayerStatRepository) *ScoreUseCase {
	return &ScoreUseCase{matches: matches, evenings: evenings, seasonStats: seasonStats}
}

// Submit records the result for a match.
// ScoreA/ScoreB are computed from the leg winners.
func (uc *ScoreUseCase) Submit(ctx context.Context, in SubmitScoreInput) error {
	match, err := uc.matches.FindByID(ctx, in.MatchID)
	if err != nil {
		return err
	}

	// Compute scores from leg winners
	scoreA, scoreB := 0, 0
	paStr := match.PlayerA.String()
	for _, winner := range []string{in.Leg1Winner, in.Leg2Winner, in.Leg3Winner} {
		if winner == "" {
			continue
		}
		if winner == paStr {
			scoreA++
		} else {
			scoreB++
		}
	}

	match.Leg1Winner = in.Leg1Winner
	match.Leg1Turns = in.Leg1Turns
	match.Leg2Winner = in.Leg2Winner
	match.Leg2Turns = in.Leg2Turns
	match.Leg3Winner = in.Leg3Winner
	match.Leg3Turns = in.Leg3Turns
	match.ReportedBy = in.ReportedBy
	match.RescheduleDate = in.RescheduleDate
	match.SecretaryNr = in.SecretaryNr
	match.CounterNr = in.CounterNr
	match.PlayedDate = in.PlayedDate
	match.PlayerA180s = in.PlayerA180s
	match.PlayerB180s = in.PlayerB180s
	match.PlayerAHighestFinish = in.PlayerAHighestFinish
	match.PlayerBHighestFinish = in.PlayerBHighestFinish
	match.ScoreA = &scoreA
	match.ScoreB = &scoreB
	match.Played = scoreA+scoreB > 0

	log.Printf("[Submit] matchID=%s reportedBy=%q leg1=%q leg2=%q leg3=%q scoreA=%d scoreB=%d played=%v",
		match.ID, match.ReportedBy, match.Leg1Winner, match.Leg2Winner, match.Leg3Winner,
		scoreA, scoreB, match.Played)

	return uc.matches.UpdateResult(ctx, match)
}

// ReportAbsent marks all unplayed matches for playerID in the given evening as reported absent.
func (uc *ScoreUseCase) ReportAbsent(ctx context.Context, eveningID domain.EveningID, playerID domain.PlayerID, reportedBy string) error {
	matches, err := uc.matches.FindByEvening(ctx, eveningID)
	if err != nil {
		return err
	}
	for _, m := range matches {
		if m.PlayerA != playerID && m.PlayerB != playerID {
			continue
		}
		if m.Played {
			continue // leave played matches untouched
		}
		if m.ReportedBy != "" {
			continue // already reported absent by the other player; keep the original reporter
		}
		m.ReportedBy = reportedBy
		if err := uc.matches.UpdateResult(ctx, m); err != nil {
			return err
		}
	}
	log.Printf("[ReportAbsent] eveningID=%s playerID=%s reportedBy=%q", eveningID, playerID, reportedBy)
	return nil
}

// GetDutyStats counts secretary+counter appearances per player from played matches,
// and returns per-match detail including the evening number and opponent names.
func (uc *ScoreUseCase) GetDutyStats(ctx context.Context, players []domain.Player, scheduleID *domain.ScheduleID) ([]DutyStats, error) {
	var matchList []domain.Match
	var err error
	if scheduleID != nil {
		seen := make(map[domain.MatchID]struct{})
		for _, p := range players {
			pm, e := uc.matches.FindByPlayerAndSchedule(ctx, p.ID, *scheduleID)
			if e != nil {
				return nil, e
			}
			for _, m := range pm {
				if m.Played {
					if _, ok := seen[m.ID]; !ok {
						seen[m.ID] = struct{}{}
						matchList = append(matchList, m)
					}
				}
			}
		}
	} else {
		matchList, err = uc.matches.FindAllPlayed(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Build lookup maps.
	playerByID := make(map[domain.PlayerID]domain.Player, len(players))
	for _, p := range players {
		playerByID[p.ID] = p
	}

	// Collect evening numbers from the evenings repo if scheduleID is set.
	eveningNr := make(map[domain.EveningID]int)
	if scheduleID != nil {
		evenings, e := uc.evenings.FindBySchedule(ctx, *scheduleID)
		if e == nil {
			for _, ev := range evenings {
				eveningNr[ev.ID] = ev.Number
			}
		}
	}

	toDutyMatch := func(m domain.Match) DutyMatch {
		pA := playerByID[m.PlayerA]
		pB := playerByID[m.PlayerB]
		return DutyMatch{
			EveningNr:   eveningNr[m.EveningID],
			PlayerANr:   pA.Nr,
			PlayerAName: pA.Name,
			PlayerBNr:   pB.Nr,
			PlayerBName: pB.Name,
		}
	}

	type dutyAccum struct {
		secMatches []DutyMatch
		cntMatches []DutyMatch
	}
	accum := make(map[string]*dutyAccum)
	for _, m := range matchList {
		if m.SecretaryNr != "" {
			a := accum[m.SecretaryNr]
			if a == nil {
				a = &dutyAccum{}
				accum[m.SecretaryNr] = a
			}
			a.secMatches = append(a.secMatches, toDutyMatch(m))
		}
		if m.CounterNr != "" {
			a := accum[m.CounterNr]
			if a == nil {
				a = &dutyAccum{}
				accum[m.CounterNr] = a
			}
			a.cntMatches = append(a.cntMatches, toDutyMatch(m))
		}
	}

	sortByEvening := func(ms []DutyMatch) {
		sort.Slice(ms, func(i, j int) bool { return ms[i].EveningNr < ms[j].EveningNr })
	}

	out := make([]DutyStats, 0, len(players))
	for _, p := range players {
		a := accum[p.Nr]
		var sec, cnt []DutyMatch
		if a != nil {
			sec = a.secMatches
			cnt = a.cntMatches
			sortByEvening(sec)
			sortByEvening(cnt)
		}
		out = append(out, DutyStats{
			Player:           p,
			SecretaryCount:   len(sec),
			CounterCount:     len(cnt),
			Count:            len(sec) + len(cnt),
			SecretaryMatches: sec,
			CounterMatches:   cnt,
		})
	}
	return out, nil
}

// GetStats computes standings statistics for the given players.
// Only played matches (Played == true) are counted.
func (uc *ScoreUseCase) GetStats(ctx context.Context, players []domain.Player, scheduleID *domain.ScheduleID) ([]PlayerStats, error) {
	statsMap := make(map[domain.PlayerID]*PlayerStats, len(players))
	for i := range players {
		statsMap[players[i].ID] = &PlayerStats{Player: players[i]}
	}

	for _, p := range players {
		var matches []domain.Match
		var err error
		if scheduleID != nil {
			matches, err = uc.matches.FindByPlayerAndSchedule(ctx, p.ID, *scheduleID)
		} else {
			matches, err = uc.matches.FindByPlayer(ctx, p.ID)
		}
		if err != nil {
			return nil, err
		}
		var wonLegTurns []int
		for _, m := range matches {
			if !m.Played || m.ScoreA == nil || m.ScoreB == nil {
				continue
			}
			st, ok := statsMap[p.ID]
			if !ok {
				continue
			}
			st.Played++
			var myScore, oppScore int
			if m.PlayerA == p.ID {
				myScore, oppScore = *m.ScoreA, *m.ScoreB
			} else {
				myScore, oppScore = *m.ScoreB, *m.ScoreA
			}
			st.PointsFor += myScore
			st.PointsAgainst += oppScore
			switch {
			case myScore > oppScore:
				st.Wins++
			case myScore < oppScore:
				st.Losses++
			default:
				st.Draws++
			}
				pid := p.ID.String()
			for _, leg := range []struct {
				w string
				t int
			}{{m.Leg1Winner, m.Leg1Turns}, {m.Leg2Winner, m.Leg2Turns}, {m.Leg3Winner, m.Leg3Turns}} {
				if leg.w == pid && leg.t > 0 {
					wonLegTurns = append(wonLegTurns, leg.t)
				}
			}
		}
		if st, ok := statsMap[p.ID]; ok && len(wonLegTurns) > 0 {
			minT := wonLegTurns[0]
			sum := 0
			for _, t := range wonLegTurns {
				sum += t
				if t < minT {
					minT = t
				}
			}
			st.MinTurns = minT
			st.AvgTurns = float64(sum) / float64(len(wonLegTurns))
			st.AvgScorePerTurn = 501.0 / st.AvgTurns
		}
	}

	// Overlay season-level 180s / highest-finish when available.
	if uc.seasonStats != nil && scheduleID != nil {
		sStats, err := uc.seasonStats.FindBySchedule(ctx, *scheduleID)
		if err != nil {
			return nil, err
		}
		sMap := make(map[domain.PlayerID]domain.SeasonPlayerStat, len(sStats))
		for _, s := range sStats {
			sMap[s.PlayerID] = s
		}
		for pid, st := range statsMap {
			if ss, ok := sMap[pid]; ok {
				st.OneEighties = ss.OneEighties
				st.HighestFinish = ss.HighestFinish
			}
		}
	}

	out := make([]PlayerStats, 0, len(players))
	for _, p := range players {
		out = append(out, *statsMap[p.ID])
	}
	return out, nil
}
