// Package scheduler generates a dart competition schedule using round-robin pairings
// with simulated annealing optimisation.
//
// Steps:
//  1. roundRobin builds all (N-1) rounds containing N*(N-1)/2 match pairs.
//  2. greedyAssign distributes matches evenly across evenings as the initial assignment.
//  3. anneal optimises the assignment by accepting or rejecting random swaps
//     based on an energy function.
//  4. The optimised assignment is converted into a domain.Schedule.
package scheduler

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"DartScheduler/domain"

	"github.com/google/uuid"
)

// Input holds all parameters needed to generate a schedule.
type Input struct {
	Players         []domain.Player
	BuddyPairs      []domain.BuddyPreference
	NumEvenings     int
	CompetitionName string
	StartDate       time.Time
	IntervalDays    int
	// EveningDates, if set, provides explicit dates for each evening (len must equal NumEvenings).
	// Takes precedence over StartDate+IntervalDays.
	EveningDates []time.Time
	// Config overrides the default annealing parameters (weights, steps, fractions).
	// If Steps == 0, DefaultAnnealConfig() is used.
	Config AnnealConfig
}

// Generate builds a round-robin Schedule with simulated-annealing optimisation.
func Generate(in Input) (domain.Schedule, error) {
	if len(in.Players) < 2 {
		return domain.Schedule{}, fmt.Errorf("%w: need at least 2 players", domain.ErrInvalidInput)
	}
	if in.NumEvenings < 1 {
		return domain.Schedule{}, fmt.Errorf("%w: numEvenings must be ≥ 1", domain.ErrInvalidInput)
	}

	log.Printf("[scheduler.Generate] players=%d numEvenings=%d buddyPairs=%d",
		len(in.Players), in.NumEvenings, len(in.BuddyPairs))

	ids := make([]domain.PlayerID, len(in.Players))
	for i, p := range in.Players {
		ids[i] = p.ID
	}

	// 1. Build full round-robin pairing list.
	rounds := roundRobin(ids)

	// Flatten rounds → matches slice.
	var flatMatches []pair
	for _, r := range rounds {
		flatMatches = append(flatMatches, r...)
	}
	if len(flatMatches) == 0 {
		return domain.Schedule{}, fmt.Errorf("%w: no matches generated", domain.ErrInvalidInput)
	}
	log.Printf("[scheduler.Generate] round-robin done: rounds=%d totalMatches=%d", len(rounds), len(flatMatches))

	// 2. Clustering initial assignment: pair consecutive rounds onto the same evening
	//    so each player starts with ≤1 solo evening instead of spreading 1 match/evening.
	assignment := clusteringAssign(rounds, in.NumEvenings)
	log.Printf("[scheduler.Generate] clustering assignment done: matchesPerEvening≈%d", len(flatMatches)/in.NumEvenings)

	// 3. Simulated annealing optimisation.
	cfg := in.Config
	if cfg.Steps == 0 {
		cfg = DefaultAnnealConfig()
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	log.Printf("[scheduler.Generate] starting annealing: steps=%d", cfg.Steps)
	assignment = anneal(cfg, flatMatches, assignment, in.NumEvenings, in.BuddyPairs, rng)
	log.Printf("[scheduler.Generate] annealing done")

	// 4. Build domain.Schedule from the optimised assignment.
	schedID := uuid.New()
	schedule := domain.Schedule{
		ID:              schedID,
		CompetitionName: in.CompetitionName,
		CreatedAt:       time.Now(),
		Evenings:        make([]domain.Evening, in.NumEvenings),
	}

	for ei := range schedule.Evenings {
		var date time.Time
		if ei < len(in.EveningDates) {
			date = in.EveningDates[ei]
		} else {
			date = in.StartDate.AddDate(0, 0, ei*in.IntervalDays)
		}
		schedule.Evenings[ei] = domain.Evening{
			ID:     uuid.New(),
			Number: ei + 1,
			Date:   date,
		}
	}

	for mi, ei := range assignment {
		m := domain.Match{
			ID:        uuid.New(),
			EveningID: schedule.Evenings[ei].ID,
			PlayerA:   flatMatches[mi].A,
			PlayerB:   flatMatches[mi].B,
		}
		schedule.Evenings[ei].Matches = append(schedule.Evenings[ei].Matches, m)
	}

	log.Printf("[scheduler.Generate] done: scheduleID=%s evenings=%d totalMatches=%d",
		schedule.ID, len(schedule.Evenings), len(flatMatches))
	return schedule, nil
}

// clusteringAssign distributes matches by pairing consecutive rounds onto the same
// evening. Each player appears in every round exactly once, so within a pair of
// rounds every player has exactly 2 matches — eliminating most solo evenings before
// annealing begins. Groups are spaced evenly across numEvenings to avoid introducing
// consecutive-evening violations in the initial assignment.
func clusteringAssign(rounds [][]pair, numEvenings int) []int {
	numRounds := len(rounds)
	if numRounds == 0 {
		return nil
	}

	// Pair consecutive rounds into groups; each group maps to one evening.
	numGroups := (numRounds + 1) / 2

	// Space groups evenly across evenings: group g → evening floor(g*numEvenings/numGroups).
	groupEvening := make([]int, numGroups)
	for g := 0; g < numGroups; g++ {
		groupEvening[g] = g * numEvenings / numGroups
	}

	total := 0
	for _, r := range rounds {
		total += len(r)
	}
	assignment := make([]int, total)

	matchIdx := 0
	for ri, round := range rounds {
		g := ri / 2 // rounds 0,1 → group 0; rounds 2,3 → group 1; etc.
		ei := groupEvening[g]
		for range round {
			assignment[matchIdx] = ei
			matchIdx++
		}
	}
	return assignment
}

// greedyAssign distributes matches across evenings as evenly as possible.
func greedyAssign(matches []pair, numEvenings int) []int {
	assignment := make([]int, len(matches))
	base := len(matches) / numEvenings
	rem := len(matches) % numEvenings
	idx := 0
	for ei := 0; ei < numEvenings; ei++ {
		count := base
		if ei < rem {
			count++
		}
		for j := 0; j < count && idx < len(matches); j++ {
			assignment[idx] = ei
			idx++
		}
	}
	return assignment
}
