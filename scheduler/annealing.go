// Dit bestand implementeert het simulated annealing algoritme.
//
// Energiefunctie (lagere waarde = beter schema):
//
//	energie = wBuddyHard    × paren zonder gedeelde avond        (hard)
//	        + wMaxViolation × avond-overschrijdingen boven max    (hard)
//	        + wTripleConsec × avonden in reeks > 2 achter elkaar  (hard)
//	        + wExcessTriple × extra 3-wedstrijd-avonden (>10%)    (medium)
//	        + wVariance     × variantie wedstrijden per avond     (soft)
package scheduler

import (
	"math"
	"math/rand"

	"DartScheduler/domain"
)

const (
	t0    = 10.0
	tEnd  = 0.001
	steps = 400_000

	wBuddyHard    = 10_000.0 // buddy pair heeft geen gedeelde avond
	wMaxViolation = 10_000.0 // speler heeft >4 wedstrijden op één avond
	wTripleConsec = 5_000.0  // speler speelt 3+ avonden op rij
	wExcessTriple = 2_000.0  // >10% van actieve avonden hebben 3 wedstrijden
	wVariance     = 1.0

	// Fraction of steps that use a targeted move instead of a random swap.
	targetedFraction = 0.5
)

// anneal improves the assignment (matchIndex → eveningIndex) using simulated annealing.
func anneal(
	matches []pair,
	assignment []int,
	numEvenings int,
	buddyPairs []domain.BuddyPreference,
	rng *rand.Rand,
) []int {
	buddyPlayers := buildBuddyPlayerSet(buddyPairs)
	best := cloneAssignment(assignment)
	current := cloneAssignment(assignment)
	bestEnergy := energy(matches, current, numEvenings, buddyPairs, buddyPlayers)
	currentEnergy := bestEnergy

	playerMatchIdx := buildPlayerMatchIndex(matches)
	alpha := math.Pow(tEnd/t0, 1.0/float64(steps))
	T := t0

	for step := 0; step < steps; step++ {
		var i, j int

		if rng.Float64() < targetedFraction {
			// 50/50 between targeting a buddy violation or a triple-consecutive violation.
			if rng.Intn(2) == 0 {
				i, j = buddyTargetedSwap(matches, current, numEvenings, buddyPairs, playerMatchIdx, rng)
			} else {
				i, j = consecutiveTargetedSwap(matches, current, numEvenings, playerMatchIdx, rng)
			}
		}
		if i < 0 || i == j {
			// Fall back to random swap.
			i = rng.Intn(len(current))
			j = rng.Intn(len(current))
		}

		if current[i] == current[j] {
			T *= alpha
			continue
		}

		current[i], current[j] = current[j], current[i]
		newEnergy := energy(matches, current, numEvenings, buddyPairs, buddyPlayers)
		delta := newEnergy - currentEnergy

		if delta < 0 || rng.Float64() < math.Exp(-delta/T) {
			currentEnergy = newEnergy
			if newEnergy < bestEnergy {
				bestEnergy = newEnergy
				best = cloneAssignment(current)
			}
		} else {
			current[i], current[j] = current[j], current[i]
		}
		T *= alpha
	}
	return best
}

// buddyTargetedSwap finds a buddy pair with a mismatch (one plays on evening ei but
// the other doesn't) and tries to fix it by swapping matches to align them.
func buddyTargetedSwap(
	matches []pair,
	current []int,
	numEvenings int,
	buddyPairs []domain.BuddyPreference,
	playerMatchIdx map[domain.PlayerID][]int,
	rng *rand.Rand,
) (int, int) {
	if len(buddyPairs) == 0 {
		return -1, -1
	}
	sets := eveningPlayerSets(matches, current, numEvenings)

	rng.Shuffle(len(buddyPairs), func(a, b int) { buddyPairs[a], buddyPairs[b] = buddyPairs[b], buddyPairs[a] })
	for _, bp := range buddyPairs {
		a, b := bp.PlayerID, bp.BuddyID

		// Find an evening where A plays but B doesn't (or vice versa).
		eveningIndices := make([]int, numEvenings)
		for i := range eveningIndices {
			eveningIndices[i] = i
		}
		rng.Shuffle(numEvenings, func(x, y int) { eveningIndices[x], eveningIndices[y] = eveningIndices[y], eveningIndices[x] })

		for _, ei := range eveningIndices {
			aPlays := sets[ei][a]
			bPlays := sets[ei][b]
			if aPlays == bPlays {
				continue // already aligned on this evening
			}

			// extra is the player who plays on ei but their buddy doesn't.
			// absent is the player who doesn't play on ei.
			extra, absent := a, b
			if !aPlays {
				extra, absent = b, a
			}

			// Strategy: find a match of `extra` on ei and swap it with a match
			// of `absent` on some other evening — bringing absent to ei.
			extraMatches := playerMatchIdx[extra]
			absentMatches := playerMatchIdx[absent]
			rng.Shuffle(len(extraMatches), func(x, y int) { extraMatches[x], extraMatches[y] = extraMatches[y], extraMatches[x] })
			rng.Shuffle(len(absentMatches), func(x, y int) { absentMatches[x], absentMatches[y] = absentMatches[y], absentMatches[x] })

			for _, mi := range extraMatches {
				if current[mi] != ei {
					continue
				}
				for _, mj := range absentMatches {
					if current[mj] != ei && mi != mj {
						return mi, mj
					}
				}
			}
		}
	}
	return -1, -1
}

// consecutiveTargetedSwap finds a player in a 3+-consecutive run and moves one of
// their matches to break the run.
func consecutiveTargetedSwap(
	matches []pair,
	current []int,
	numEvenings int,
	playerMatchIdx map[domain.PlayerID][]int,
	rng *rand.Rand,
) (int, int) {
	sets := eveningPlayerSets(matches, current, numEvenings)

	allPlayers := make([]domain.PlayerID, 0, len(playerMatchIdx))
	for pid := range playerMatchIdx {
		allPlayers = append(allPlayers, pid)
	}
	rng.Shuffle(len(allPlayers), func(a, b int) { allPlayers[a], allPlayers[b] = allPlayers[b], allPlayers[a] })

	for _, pid := range allPlayers {
		// Find a run of >2 consecutive evenings.
		run := 0
		runStart := -1
		for ei := 0; ei < numEvenings; ei++ {
			if sets[ei][pid] {
				run++
				if run == 1 {
					runStart = ei
				}
				if run > maxConsecutiveEvenings {
					// Evening ei is the violation point; try to move one of pid's matches on ei.
					mList := playerMatchIdx[pid]
					rng.Shuffle(len(mList), func(a, b int) { mList[a], mList[b] = mList[b], mList[a] })
					for _, mi := range mList {
						if current[mi] != ei {
							continue
						}
						// Swap with any match on an evening not in the run.
						for _, mj := range randomSample(len(current), 30, rng) {
							ej := current[mj]
							// Avoid swapping into the run [runStart, ei].
							if ej < runStart || ej > ei {
								if mi != mj {
									return mi, mj
								}
							}
						}
					}
					_ = runStart
				}
			} else {
				run = 0
				runStart = -1
			}
		}
	}
	return -1, -1
}

// buildPlayerMatchIndex maps each player ID to the list of match indices they participate in.
func buildPlayerMatchIndex(matches []pair) map[domain.PlayerID][]int {
	idx := make(map[domain.PlayerID][]int)
	for mi, m := range matches {
		idx[m.A] = append(idx[m.A], mi)
		idx[m.B] = append(idx[m.B], mi)
	}
	return idx
}

// randomSample returns up to n random indices in [0, total).
func randomSample(total, n int, rng *rand.Rand) []int {
	if total < n {
		n = total
	}
	out := make([]int, n)
	for i := range out {
		out[i] = rng.Intn(total)
	}
	return out
}

func energy(matches []pair, assignment []int, numEvenings int, buddyPairs []domain.BuddyPreference, buddyPlayers map[domain.PlayerID]bool) float64 {
	buddyV := float64(countBuddyHardViolations(matches, assignment, buddyPairs, numEvenings))
	maxV := float64(countMaxViolations(matches, assignment, numEvenings, buddyPlayers))
	tripleV := float64(countTripleConsecutiveViolations(matches, assignment, numEvenings))
	excessV := float64(countExcessTripleMatchViolations(matches, assignment, numEvenings))
	va := varianceMatchesPerEvening(assignment, numEvenings)
	return wBuddyHard*buddyV + wMaxViolation*maxV + wTripleConsec*tripleV + wExcessV*excessV + wVariance*va
}

// buildBuddyPlayerSet returns the set of all player IDs that appear in any buddy pair.
func buildBuddyPlayerSet(buddyPairs []domain.BuddyPreference) map[domain.PlayerID]bool {
	s := make(map[domain.PlayerID]bool, len(buddyPairs)*2)
	for _, bp := range buddyPairs {
		s[bp.PlayerID] = true
		s[bp.BuddyID] = true
	}
	return s
}

const wExcessV = wExcessTriple // alias for readability in energy()

func cloneAssignment(a []int) []int {
	b := make([]int, len(a))
	copy(b, a)
	return b
}
