// This file implements the simulated annealing algorithm.
//
// Energy function (lower = better schedule):
//
//	energy = WBuddyHard    × buddy pair mismatches beyond the 1 allowed     (hard)
//	       + WBuddySoft    × buddy pairs with exactly 1 mismatch (allowed)  (soft)
//	       + WMaxViolation × per-evening match counts above the cap          (hard)
//	       + WTripleConsec × evenings in a run of >2 consecutive             (hard)
//	       + WGapViolation × excess gap between a player's active evenings   (hard)
//	       + WExcessTriple × extra 3-match evenings per player (>10%)        (medium)
//	       + WMinMatches   × per-player excess solo evenings (>1 solo)       (hard)
//	       + WSoloSoft     × total solo (player,evening) pairs               (soft)
//	       + WSpread       × excess spread (max−min matches/evening > 5)     (hard)
//	       + WVariance     × variance of total matches per evening           (soft, balance)
package scheduler

import (
	"log"
	"math"
	"math/rand"

	"DartScheduler/domain"
)

const logInterval = 100_000

// AnnealConfig holds all tunable parameters for the simulated annealing algorithm.
// Use DefaultAnnealConfig() for production; override fields in tests for fast runs.
type AnnealConfig struct {
	T0    float64 // initial temperature
	TEnd  float64 // final temperature
	Steps int     // total number of annealing steps

	// MoveFraction is the fraction of steps that use a "move" operation (assign a
	// match to any random evening, including empty ones). This is the primary
	// mechanism for populating empty evenings. The remaining steps use targeted
	// or random swaps.
	MoveFraction float64

	// TargetedFraction is the fraction of non-move steps that use a targeted swap
	// (buddy or consecutive violation) instead of a random swap.
	TargetedFraction float64

	// Weights for the energy function.
	WBuddyHard    float64 // buddy pair has >1 mismatch evening
	WBuddySoft    float64 // buddy pair has exactly 1 mismatch evening (allowed, but prefer 0)
	WMaxViolation float64 // player has more than the allowed matches on one evening
	WTripleConsec float64 // player plays 3+ evenings in a row
	WGapViolation float64 // gap between consecutive active evenings exceeds maxGapBetweenActiveEvenings
	WExcessTriple float64 // >10% of active evenings have 3 matches for a player
	WMinMatches   float64 // player has >1 solo evening (only 1 is allowed)
	WSoloSoft     float64 // any solo evening — nudges toward 0 solo evenings
	WSpread       float64 // spread (max−min matches/evening) exceeds maxEveningSpread (5)
	WVariance     float64 // variance of match counts across evenings

	// ProgressFn is called every logInterval steps with the current step and total.
	// If nil, progress reporting is skipped.
	ProgressFn func(step, total int)
}

// DefaultAnnealConfig returns the production-quality annealing configuration.
// WVariance penalises unequal match counts across evenings; at 500 a ±1 imbalance
// costs ~50 energy, strong enough to drive balancing while remaining softer than
// hard constraints (1000–10000).
func DefaultAnnealConfig() AnnealConfig {
	return AnnealConfig{
		T0:    100.0, // raised so spreading to empty evenings (delta≈+20) is accepted early
		TEnd:  0.001,
		Steps: 1_200_000,

		MoveFraction:     0.20,
		TargetedFraction: 0.50,

		WBuddyHard:    10_000.0,
		WBuddySoft:    200.0,
		WMaxViolation: 10_000.0,
		WTripleConsec: 5_000.0,
		WGapViolation: 5_000.0,
		WExcessTriple: 2_000.0,
		WMinMatches:   5_000.0,
		WSoloSoft:     5.0,
		WSpread:       5_000.0,
		WVariance:     500.0,
	}
}

// anneal improves the assignment (matchIndex → eveningIndex) using simulated annealing.
func anneal(
	cfg AnnealConfig,
	matches []pair,
	assignment []int,
	numEvenings int,
	buddyPairs []domain.BuddyPreference,
	rng *rand.Rand,
) []int {
	buddyPlayers := buildBuddyPlayerSet(buddyPairs)
	best := cloneAssignment(assignment)
	current := cloneAssignment(assignment)
	bestEnergy := energy(cfg, matches, current, numEvenings, buddyPairs, buddyPlayers)
	currentEnergy := bestEnergy

	playerMatchIdx := buildPlayerMatchIndex(matches)
	alpha := math.Pow(cfg.TEnd/cfg.T0, 1.0/float64(cfg.Steps))
	T := cfg.T0

	// Report progress ~100 times per run regardless of step count so the progress
	// bar stays smooth even when workers run fewer steps than logInterval.
	progressInterval := cfg.Steps / 100
	if progressInterval < 1 {
		progressInterval = 1
	}

	for step := 0; step < cfg.Steps; step++ {
		if step%logInterval == 0 {
			log.Printf("[anneal] step=%d/%d T=%.4f currentEnergy=%.1f bestEnergy=%.1f",
				step, cfg.Steps, T, currentEnergy, bestEnergy)
		}
		if cfg.ProgressFn != nil && step%progressInterval == 0 {
			cfg.ProgressFn(step, cfg.Steps)
		}

		// Move operation: assign one match to any random evening (can populate empties).
		if rng.Float64() < cfg.MoveFraction {
			mi := rng.Intn(len(current))
			oldEvening := current[mi]
			newEvening := rng.Intn(numEvenings)
			if oldEvening != newEvening {
				current[mi] = newEvening
				newEnergy := energy(cfg, matches, current, numEvenings, buddyPairs, buddyPlayers)
				delta := newEnergy - currentEnergy
				if delta < 0 || rng.Float64() < math.Exp(-delta/T) {
					currentEnergy = newEnergy
					if newEnergy < bestEnergy {
						bestEnergy = newEnergy
						best = cloneAssignment(current)
					}
				} else {
					current[mi] = oldEvening
				}
			}
			T *= alpha
			continue
		}

		// Swap operation: exchange two match evening assignments.
		var i, j int

		if rng.Float64() < cfg.TargetedFraction {
			// Rotate among four targeted strategies: buddy, consecutive, gap, solo.
			switch rng.Intn(4) {
			case 0:
				i, j = buddyTargetedSwap(matches, current, numEvenings, buddyPairs, playerMatchIdx, rng)
			case 1:
				i, j = consecutiveTargetedSwap(matches, current, numEvenings, playerMatchIdx, rng)
			case 2:
				i, j = gapTargetedSwap(matches, current, numEvenings, playerMatchIdx, rng)
			case 3:
				i, j = soloTargetedSwap(matches, current, numEvenings, playerMatchIdx, rng)
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
		newEnergy := energy(cfg, matches, current, numEvenings, buddyPairs, buddyPlayers)
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
	log.Printf("[anneal] done: finalBestEnergy=%.1f", bestEnergy)
	if cfg.ProgressFn != nil {
		cfg.ProgressFn(cfg.Steps, cfg.Steps)
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

	shuffledPairs := make([]domain.BuddyPreference, len(buddyPairs))
	copy(shuffledPairs, buddyPairs)
	rng.Shuffle(len(shuffledPairs), func(a, b int) { shuffledPairs[a], shuffledPairs[b] = shuffledPairs[b], shuffledPairs[a] })
	for _, bp := range shuffledPairs {
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
			extraMatches := make([]int, len(playerMatchIdx[extra]))
			copy(extraMatches, playerMatchIdx[extra])
			absentMatches := make([]int, len(playerMatchIdx[absent]))
			copy(absentMatches, playerMatchIdx[absent])
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

// gapTargetedSwap finds a player with a gap > maxGapBetweenActiveEvenings and tries to
// move one of their matches into the gap by swapping it with a match currently sitting
// inside the gap window.
func gapTargetedSwap(
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
		// Find a gap > maxGapBetweenActiveEvenings.
		lastActive := -1
		for ei := 0; ei < numEvenings; ei++ {
			if sets[ei][pid] {
				if lastActive >= 0 && ei-lastActive > maxGapBetweenActiveEvenings {
					// Gap (lastActive, ei) is too large. Find a match of pid outside this
					// gap and swap it with any match currently in the gap window.
					gapStart := lastActive + 1
					gapEnd := ei - 1

					mList := playerMatchIdx[pid]
					rng.Shuffle(len(mList), func(a, b int) { mList[a], mList[b] = mList[b], mList[a] })
					for _, mi := range mList {
						// Pick a match of pid that is outside the gap window.
						if current[mi] >= gapStart && current[mi] <= gapEnd {
							continue // already inside gap — we want something outside
						}
						// Swap with any match whose evening falls inside the gap window.
						for _, mj := range randomSample(len(current), 30, rng) {
							if current[mj] >= gapStart && current[mj] <= gapEnd && mi != mj {
								return mi, mj
							}
						}
					}
				}
				lastActive = ei
			}
		}
	}
	return -1, -1
}

// soloTargetedSwap finds a player with ≥2 solo evenings and consolidates two of them:
// it swaps the player's match on soloA with a non-player match on soloB.
// After the swap the player has 0 matches on soloA and 2 on soloB (eliminating both solos).
func soloTargetedSwap(
	matches []pair,
	current []int,
	numEvenings int,
	playerMatchIdx map[domain.PlayerID][]int,
	rng *rand.Rand,
) (int, int) {
	counts := playerCountsPerEvening(matches, current, numEvenings)

	// Build evening → match indices for efficient lookup.
	eveningMatches := make([][]int, numEvenings)
	for mi, ei := range current {
		eveningMatches[ei] = append(eveningMatches[ei], mi)
	}

	// Build per-player list of solo evenings.
	playerSolos := make(map[domain.PlayerID][]int)
	for ei := 0; ei < numEvenings; ei++ {
		for pid, c := range counts[ei] {
			if c == 1 {
				playerSolos[pid] = append(playerSolos[pid], ei)
			}
		}
	}

	// Collect players with ≥2 solo evenings.
	candidates := make([]domain.PlayerID, 0, len(playerSolos))
	for pid, solos := range playerSolos {
		if len(solos) >= 2 {
			candidates = append(candidates, pid)
		}
	}
	if len(candidates) == 0 {
		return -1, -1
	}
	rng.Shuffle(len(candidates), func(a, b int) { candidates[a], candidates[b] = candidates[b], candidates[a] })

	for _, pid := range candidates {
		solos := playerSolos[pid]
		rng.Shuffle(len(solos), func(a, b int) { solos[a], solos[b] = solos[b], solos[a] })

		for si := 0; si < len(solos); si++ {
			soloA := solos[si]

			// Find pid's match on soloA.
			matchA := -1
			for _, mi := range playerMatchIdx[pid] {
				if current[mi] == soloA {
					matchA = mi
					break
				}
			}
			if matchA < 0 {
				continue
			}

			// Find soloB with a non-pid match to swap with.
			for sj := 0; sj < len(solos); sj++ {
				if sj == si {
					continue
				}
				soloB := solos[sj]
				mList := eveningMatches[soloB]
				for _, mj := range mList {
					if matches[mj].A != pid && matches[mj].B != pid {
						return matchA, mj
					}
				}
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

func energy(cfg AnnealConfig, matches []pair, assignment []int, numEvenings int, buddyPairs []domain.BuddyPreference, buddyPlayers map[domain.PlayerID]bool) float64 {
	buddyHard := float64(countBuddyHardViolations(matches, assignment, buddyPairs, numEvenings))
	buddySoft := float64(countBuddySoftViolations(matches, assignment, buddyPairs, numEvenings))
	maxV := float64(countMaxViolations(matches, assignment, numEvenings, buddyPlayers))
	tripleV := float64(countTripleConsecutiveViolations(matches, assignment, numEvenings))
	gapV := float64(countGapViolations(matches, assignment, numEvenings))
	excessV := float64(countExcessTripleMatchViolations(matches, assignment, numEvenings))
	minV := float64(countMinMatchViolations(matches, assignment, numEvenings))
	soloV := float64(countSoloEvenings(matches, assignment, numEvenings))
	spreadV := float64(countSpreadViolation(assignment, numEvenings))
	va := varianceMatchesPerEvening(assignment, numEvenings)
	return cfg.WBuddyHard*buddyHard + cfg.WBuddySoft*buddySoft +
		cfg.WMaxViolation*maxV + cfg.WTripleConsec*tripleV + cfg.WGapViolation*gapV +
		cfg.WExcessTriple*excessV + cfg.WMinMatches*minV + cfg.WSoloSoft*soloV +
		cfg.WSpread*spreadV + cfg.WVariance*va
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

func cloneAssignment(a []int) []int {
	b := make([]int, len(a))
	copy(b, a)
	return b
}
