package usecase

import (
	"context"
	"log"

	"DartScheduler/domain"
)

type ScoreUseCase struct {
	matches domain.MatchRepository
}

func NewScoreUseCase(matches domain.MatchRepository) *ScoreUseCase {
	return &ScoreUseCase{matches: matches}
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
	match.ScoreA = &scoreA
	match.ScoreB = &scoreB
	match.Played = scoreA+scoreB > 0

	log.Printf("[Submit] matchID=%s reportedBy=%q leg1=%q leg2=%q leg3=%q scoreA=%d scoreB=%d played=%v",
		match.ID, match.ReportedBy, match.Leg1Winner, match.Leg2Winner, match.Leg3Winner,
		scoreA, scoreB, match.Played)

	return uc.matches.UpdateResult(ctx, match)
}

// GetDutyStats counts secretary+counter appearances per player from played matches.
func (uc *ScoreUseCase) GetDutyStats(ctx context.Context, players []domain.Player, scheduleID *domain.ScheduleID) ([]DutyStats, error) {
	var matchList []domain.Match
	var err error
	if scheduleID != nil {
		// collect all played matches for this schedule via all players (dedup by match ID)
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
	counts := make(map[string]int, len(players))
	for _, m := range matchList {
		if m.SecretaryNr != "" {
			counts[m.SecretaryNr]++
		}
		if m.CounterNr != "" {
			counts[m.CounterNr]++
		}
	}
	out := make([]DutyStats, 0, len(players))
	for _, p := range players {
		out = append(out, DutyStats{Player: p, Count: counts[p.Nr]})
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
		}
	}

	out := make([]PlayerStats, 0, len(players))
	for _, p := range players {
		out = append(out, *statsMap[p.ID])
	}
	return out, nil
}
