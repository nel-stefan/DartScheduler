package scheduler

import "DartScheduler/domain"

// byePlayerID is a sentinel used when N is odd.
var byePlayerID = domain.PlayerID{}

// roundRobin generates the canonical circle-method round-robin pairings.
// Returns (rounds, isBye map) where rounds[r] is the list of (A,B) pairs for round r.
// If N is odd a BYE player (zero UUID) is added; the player paired with BYE has no opponent.
func roundRobin(players []domain.PlayerID) [][]pair {
	ps := make([]domain.PlayerID, len(players))
	copy(ps, players)

	if len(ps)%2 != 0 {
		ps = append(ps, byePlayerID)
	}
	n := len(ps)
	rounds := make([][]pair, n-1)

	for r := range rounds {
		half := n / 2
		round := make([]pair, 0, half)
		for i := 0; i < half; i++ {
			a, b := ps[i], ps[n-1-i]
			if a != byePlayerID && b != byePlayerID {
				round = append(round, pair{a, b})
			}
		}
		rounds[r] = round
		// Rotate positions 1..n-1; position 0 stays fixed.
		last := ps[n-1]
		copy(ps[2:], ps[1:n-1])
		ps[1] = last
	}
	return rounds
}

type pair struct{ A, B domain.PlayerID }
