// Package scheduler genereert een dartcompetitie-schema op basis van round-robin
// met simulated annealing optimalisatie.
//
// Werkwijze:
//  1. roundRobin bouwt alle (N-1) rondes met N*(N-1)/2 wedstrijdparen.
//  2. greedyAssign verdeelt de wedstrijden gelijkmatig over de avonden als startpunt.
//  3. anneal optimaliseert de verdeling door willekeurige verwisselingen te accepteren
//     of te verwerpen op basis van een energiefunctie.
//  4. De geoptimaliseerde toewijzing wordt omgezet naar domain.Schedule.
package scheduler

import (
	"fmt"
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
}

// Generate builds a round-robin Schedule with simulated-annealing optimisation.
func Generate(in Input) (domain.Schedule, error) {
	if len(in.Players) < 2 {
		return domain.Schedule{}, fmt.Errorf("%w: need at least 2 players", domain.ErrInvalidInput)
	}
	if in.NumEvenings < 1 {
		return domain.Schedule{}, fmt.Errorf("%w: numEvenings must be ≥ 1", domain.ErrInvalidInput)
	}

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

	// 2. Greedy initial assignment: distribute matches across evenings evenly.
	assignment := greedyAssign(flatMatches, in.NumEvenings)

	// 3. Simulated annealing optimisation.
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	assignment = anneal(flatMatches, assignment, in.NumEvenings, in.BuddyPairs, rng)

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

	return schedule, nil
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
