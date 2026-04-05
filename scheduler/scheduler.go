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
	"math"
	"math/rand"
	"runtime"
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
	if len(in.Players) == 0 {
		return domain.Schedule{}, fmt.Errorf("%w: er zijn geen spelers geïmporteerd — importeer eerst spelers via het Beheer-scherm", domain.ErrInvalidInput)
	}
	if len(in.Players) < 2 {
		return domain.Schedule{}, fmt.Errorf("%w: minimaal 2 spelers vereist, maar er is slechts 1 speler geïmporteerd", domain.ErrInvalidInput)
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

	// 2. Initial assignment.
	//
	// When numEvenings ≤ numRounds: clusteringAssign groups consecutive rounds onto the
	// same evening so each player immediately has ≥2 matches per active evening.
	//
	// When numEvenings > numRounds: clusteringAssign would leave many evenings empty
	// (only numRounds/2 evenings populated). Instead use greedyAssign for a balanced
	// spread across all evenings; soloTargetedSwap then consolidates per-player matches.
	var assignment []int
	if in.NumEvenings <= len(rounds) {
		assignment = clusteringAssign(rounds, in.NumEvenings)
	} else {
		assignment = greedyAssign(flatMatches, in.NumEvenings)
	}
	log.Printf("[scheduler.Generate] initial assignment done: matchesPerEvening≈%d", len(flatMatches)/in.NumEvenings)

	// 3. Simulated annealing optimisation — run one independent worker per CPU core,
	//    each with a different random seed and Steps/numWorkers steps, then pick the
	//    best result. This gives ~numWorkers× speedup with no quality loss because
	//    multiple restarts often find better solutions than a single long run.
	cfg := in.Config
	if cfg.Steps == 0 {
		cfg = DefaultAnnealConfig()
	}

	// Scale total steps with problem size: larger schedules need more annealing
	// iterations to converge. Minimum 10K steps per match pair ensures adequate
	// convergence for N=20, E=30 (190 matches → 1.9M steps) without affecting
	// smaller schedules (N=12, E=11 → 660K < 1.2M default, so no scaling).
	const minStepsPerMatch = 10_000
	if minScaled := minStepsPerMatch * len(flatMatches); minScaled > cfg.Steps {
		cfg.Steps = minScaled
	}

	// Each worker needs enough steps to redistribute matches from the clustered
	// initial assignment to all evenings. Below ~300K steps convergence becomes
	// unreliable, leaving some evenings with 0 matches.
	const minStepsPerWorker = 300_000

	numWorkers := runtime.NumCPU()
	if numWorkers < 1 {
		numWorkers = 1
	}
	stepsPerWorker := cfg.Steps / numWorkers
	if stepsPerWorker < minStepsPerWorker {
		stepsPerWorker = minStepsPerWorker
		numWorkers = cfg.Steps / stepsPerWorker
		if numWorkers < 1 {
			numWorkers = 1
		}
	}
	log.Printf("[scheduler.Generate] starting annealing: workers=%d stepsPerWorker=%d totalSteps=%d",
		numWorkers, stepsPerWorker, cfg.Steps)

	type workerResult struct {
		assignment []int
		e          float64
	}
	resultCh := make(chan workerResult, numWorkers)
	buddyPlayers := buildBuddyPlayerSet(in.BuddyPairs)

	for w := range numWorkers {
		workerCfg := cfg
		workerCfg.Steps = stepsPerWorker
		workerCfg.ProgressFn = nil
		// Worker 0 reports progress scaled so it spans the full original Steps range.
		if w == 0 && cfg.ProgressFn != nil {
			fn := cfg.ProgressFn
			scale := numWorkers
			workerCfg.ProgressFn = func(step, _ int) { fn(step*scale, 0) }
		}
		rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(w)*997))
		init := cloneAssignment(assignment)
		go func(wCfg AnnealConfig, init []int, rng *rand.Rand) {
			best := anneal(wCfg, flatMatches, init, in.NumEvenings, in.BuddyPairs, rng)
			e := energy(wCfg, flatMatches, best, in.NumEvenings, in.BuddyPairs, buddyPlayers)
			resultCh <- workerResult{best, e}
		}(workerCfg, init, rng)
	}

	bestEnergy := math.Inf(1)
	for range numWorkers {
		r := <-resultCh
		if r.e < bestEnergy {
			bestEnergy = r.e
			assignment = r.assignment
		}
	}
	log.Printf("[scheduler.Generate] annealing done: bestEnergy=%.1f workers=%d", bestEnergy, numWorkers)

	// Safety pass: if any regular evening still has 0 matches (can happen when
	// the annealing doesn't fully converge from the clustered initial state),
	// redistribute one match at a time from the most-loaded evening.
	assignment = fillEmptyEvenings(assignment, in.NumEvenings)

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

// fillEmptyEvenings ensures every evening has at least one match by moving one match
// at a time from the most-loaded evening to each empty evening.
func fillEmptyEvenings(assignment []int, numEvenings int) []int {
	result := cloneAssignment(assignment)
	for {
		counts := make([]int, numEvenings)
		for _, ei := range result {
			counts[ei]++
		}
		// Find first empty evening.
		emptyEvening := -1
		for ei, c := range counts {
			if c == 0 {
				emptyEvening = ei
				break
			}
		}
		if emptyEvening < 0 {
			return result // all evenings have at least one match
		}
		// Move one match from the most-loaded evening to the empty one.
		maxEvening := 0
		for ei, c := range counts {
			if c > counts[maxEvening] {
				maxEvening = ei
			}
		}
		for mi, ei := range result {
			if ei == maxEvening {
				result[mi] = emptyEvening
				break
			}
		}
	}
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
