package scheduler

import (
	"math"

	"DartScheduler/domain"
)

const (
	maxMatchesPerPlayerPerEvening      = 3    // hard ceiling for regular players
	maxMatchesPerBuddyPlayerPerEvening = 4    // buddies need extra slack to align on shared evenings
	maxConsecutiveEvenings             = 2    // playing 3+ evenings in a row is forbidden
	maxTripleFraction                  = 0.10 // at most 10% of active evenings may have 3 matches
)

// playerCountsPerEvening returns counts[eveningIdx][playerID] = matchCount.
func playerCountsPerEvening(matches []pair, assignment []int, numEvenings int) []map[domain.PlayerID]int {
	counts := make([]map[domain.PlayerID]int, numEvenings)
	for i := range counts {
		counts[i] = make(map[domain.PlayerID]int)
	}
	for mi, ei := range assignment {
		m := matches[mi]
		counts[ei][m.A]++
		counts[ei][m.B]++
	}
	return counts
}

// eveningPlayerSets returns a bool-set of players per evening.
func eveningPlayerSets(matches []pair, assignment []int, numEvenings int) []map[domain.PlayerID]bool {
	sets := make([]map[domain.PlayerID]bool, numEvenings)
	for i := range sets {
		sets[i] = make(map[domain.PlayerID]bool)
	}
	for mi, ei := range assignment {
		m := matches[mi]
		sets[ei][m.A] = true
		sets[ei][m.B] = true
	}
	return sets
}

// countMaxViolations counts (excess matches) per player per evening above the hard ceiling.
// Buddy players may have up to maxMatchesPerBuddyPlayerPerEvening; others are capped at maxMatchesPerPlayerPerEvening.
func countMaxViolations(matches []pair, assignment []int, numEvenings int, buddyPlayers map[domain.PlayerID]bool) int {
	counts := playerCountsPerEvening(matches, assignment, numEvenings)
	violations := 0
	for _, ev := range counts {
		for pid, c := range ev {
			limit := maxMatchesPerPlayerPerEvening
			if buddyPlayers[pid] {
				limit = maxMatchesPerBuddyPlayerPerEvening
			}
			if c > limit {
				violations += c - limit
			}
		}
	}
	return violations
}

// countBuddyHardViolations counts per-evening mismatches for each buddy pair:
// every evening where one plays but the other doesn't is one violation.
// Duplicate (bidirectional) buddy preferences are counted only once per pair.
func countBuddyHardViolations(matches []pair, assignment []int, buddyPairs []domain.BuddyPreference, numEvenings int) int {
	if len(buddyPairs) == 0 {
		return 0
	}
	sets := eveningPlayerSets(matches, assignment, numEvenings)

	type pairKey [2]domain.PlayerID
	seen := make(map[pairKey]bool)
	violations := 0
	for _, bp := range buddyPairs {
		a, b := bp.PlayerID, bp.BuddyID
		key := pairKey{a, b}
		revKey := pairKey{b, a}
		if seen[key] || seen[revKey] {
			continue
		}
		seen[key] = true
		for _, s := range sets {
			if s[a] != s[b] { // one plays, the other doesn't
				violations++
			}
		}
	}
	return violations
}

// countTripleConsecutiveViolations counts per-player evening appearances beyond the
// maxConsecutiveEvenings limit. A run of length k (k > maxConsecutiveEvenings)
// contributes k - maxConsecutiveEvenings violations.
func countTripleConsecutiveViolations(matches []pair, assignment []int, numEvenings int) int {
	sets := eveningPlayerSets(matches, assignment, numEvenings)

	// Collect all players.
	allPlayers := make(map[domain.PlayerID]bool)
	for _, s := range sets {
		for pid := range s {
			allPlayers[pid] = true
		}
	}

	violations := 0
	for pid := range allPlayers {
		run := 0
		for ei := 0; ei < numEvenings; ei++ {
			if sets[ei][pid] {
				run++
				if run > maxConsecutiveEvenings {
					violations++
				}
			} else {
				run = 0
			}
		}
	}
	return violations
}

// countExcessTripleMatchViolations counts extra 3-match evenings per player beyond
// the allowed 10% of their active evenings. Each extra triple is one violation.
func countExcessTripleMatchViolations(matches []pair, assignment []int, numEvenings int) int {
	counts := playerCountsPerEvening(matches, assignment, numEvenings)

	activeEvenings := make(map[domain.PlayerID]int)
	tripleEvenings := make(map[domain.PlayerID]int)
	for _, ev := range counts {
		for pid, c := range ev {
			activeEvenings[pid]++
			if c >= 3 {
				tripleEvenings[pid]++
			}
		}
	}

	violations := 0
	for pid, active := range activeEvenings {
		allowed := int(math.Ceil(maxTripleFraction * float64(active)))
		triple := tripleEvenings[pid]
		if triple > allowed {
			violations += triple - allowed
		}
	}
	return violations
}

// varianceMatchesPerEvening returns the variance of total matches per evening.
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
