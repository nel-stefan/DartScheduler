package scheduler

import "DartScheduler/domain"

const maxMatchesPerPlayerPerEvening = 3

// countViolations returns the number of "player plays more than maxMatchesPerPlayerPerEvening
// matches in a single evening" violations for the given assignment.
// assignment maps matchIndex → eveningIndex.
func countViolations(matches []pair, assignment []int, numEvenings int) int {
	// counts[evening][player] = number of matches
	counts := make([]map[domain.PlayerID]int, numEvenings)
	for i := range counts {
		counts[i] = make(map[domain.PlayerID]int)
	}
	for mi, ei := range assignment {
		m := matches[mi]
		counts[ei][m.A]++
		counts[ei][m.B]++
	}
	violations := 0
	for _, ev := range counts {
		for _, c := range ev {
			if c > maxMatchesPerPlayerPerEvening {
				violations += c - maxMatchesPerPlayerPerEvening
			}
		}
	}
	return violations
}

// buddySatisfaction returns the fraction of buddy pairs that appear in the same evening
// (at least once). Value in [0,1]; higher is better.
func buddySatisfaction(matches []pair, assignment []int, buddyPairs []domain.BuddyPreference, numEvenings int) float64 {
	if len(buddyPairs) == 0 {
		return 1.0
	}
	// Build per-evening sets of players.
	eveningPlayers := make([]map[domain.PlayerID]struct{}, numEvenings)
	for i := range eveningPlayers {
		eveningPlayers[i] = make(map[domain.PlayerID]struct{})
	}
	for mi, ei := range assignment {
		m := matches[mi]
		eveningPlayers[ei][m.A] = struct{}{}
		eveningPlayers[ei][m.B] = struct{}{}
	}

	satisfied := 0
	for _, bp := range buddyPairs {
		for _, ep := range eveningPlayers {
			_, hasP := ep[bp.PlayerID]
			_, hasB := ep[bp.BuddyID]
			if hasP && hasB {
				satisfied++
				break
			}
		}
	}
	return float64(satisfied) / float64(len(buddyPairs))
}

// varianceMatchesPerEvening returns the variance of the number of matches per evening.
func varianceMatchesPerEvening(assignment []int, numEvenings int) float64 {
	counts := make([]float64, numEvenings)
	for _, ei := range assignment {
		counts[ei]++
	}
	mean := float64(len(assignment)) / float64(numEvenings)
	var variance float64
	for _, c := range counts {
		d := c - mean
		variance += d * d
	}
	return variance / float64(numEvenings)
}
