package usecase

import (
	"context"
	"fmt"

	"DartScheduler/domain"
)

type ScoreUseCase struct {
	matches domain.MatchRepository
}

func NewScoreUseCase(matches domain.MatchRepository) *ScoreUseCase {
	return &ScoreUseCase{matches: matches}
}

func (uc *ScoreUseCase) Submit(ctx context.Context, in SubmitScoreInput) error {
	match, err := uc.matches.FindByID(ctx, in.MatchID)
	if err != nil {
		return err
	}
	if match.Played {
		return fmt.Errorf("%w: match %s", domain.ErrMatchAlreadyPlayed, in.MatchID)
	}
	return uc.matches.UpdateScore(ctx, in.MatchID, in.ScoreA, in.ScoreB)
}

func (uc *ScoreUseCase) GetStats(ctx context.Context, players []domain.Player) ([]PlayerStats, error) {
	statsMap := make(map[domain.PlayerID]*PlayerStats, len(players))
	for i := range players {
		statsMap[players[i].ID] = &PlayerStats{Player: players[i]}
	}

	for _, p := range players {
		matches, err := uc.matches.FindByPlayer(ctx, p.ID)
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
