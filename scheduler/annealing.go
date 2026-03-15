// Dit bestand implementeert het simulated annealing algoritme dat de
// wedstrijdverdeling over avonden optimaliseert.
//
// Energiefunctie:
//
//	energie = wViolation × overtredingen
//	        + wBuddy    × (1 − buddy_tevredenheid)
//	        + wVariance × variantie_wedstrijden_per_avond
//
// Een overtreding treedt op wanneer een speler meer dan maxMatchesPerPlayerPerEvening
// wedstrijden op één avond heeft. Buddy-tevredenheid meet de fractie van
// buddy-koppels die op minstens één avond samen spelen.
package scheduler

import (
	"math"
	"math/rand"

	"DartScheduler/domain"
)

const (
	t0    = 10.0
	tEnd  = 0.001
	steps = 50_000

	wViolation = 1000.0
	wBuddy     = 100.0
	wVariance  = 1.0
)

// anneal improves the assignment (matchIndex → eveningIndex) in-place using
// simulated annealing and returns the best assignment found.
func anneal(
	matches []pair,
	assignment []int,
	numEvenings int,
	buddyPairs []domain.BuddyPreference,
	rng *rand.Rand,
) []int {
	best := cloneAssignment(assignment)
	current := cloneAssignment(assignment)
	bestEnergy := energy(matches, current, numEvenings, buddyPairs)
	currentEnergy := bestEnergy

	alpha := math.Pow(tEnd/t0, 1.0/float64(steps))
	T := t0

	for step := 0; step < steps; step++ {
		// Neighbour: swap two random matches into different evenings.
		i := rng.Intn(len(current))
		j := rng.Intn(len(current))
		if current[i] == current[j] {
			T *= alpha
			continue
		}
		current[i], current[j] = current[j], current[i]

		newEnergy := energy(matches, current, numEvenings, buddyPairs)
		delta := newEnergy - currentEnergy

		if delta < 0 || rng.Float64() < math.Exp(-delta/T) {
			currentEnergy = newEnergy
			if newEnergy < bestEnergy {
				bestEnergy = newEnergy
				best = cloneAssignment(current)
			}
		} else {
			// Revert.
			current[i], current[j] = current[j], current[i]
		}
		T *= alpha
	}
	return best
}

func energy(matches []pair, assignment []int, numEvenings int, buddyPairs []domain.BuddyPreference) float64 {
	v := float64(countViolations(matches, assignment, numEvenings))
	bs := buddySatisfaction(matches, assignment, buddyPairs, numEvenings)
	va := varianceMatchesPerEvening(assignment, numEvenings)
	return wViolation*v + wBuddy*(1-bs) + wVariance*va
}

func cloneAssignment(a []int) []int {
	b := make([]int, len(a))
	copy(b, a)
	return b
}
