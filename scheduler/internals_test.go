// Whitebox tests for internal scheduler functions.
// Uses package scheduler (not scheduler_test) to access unexported types and functions.
package scheduler

import (
	"testing"

	"DartScheduler/domain"

	"github.com/google/uuid"
)

func newPID() domain.PlayerID { return domain.PlayerID(uuid.New()) }

// ---------------------------------------------------------------------------
// countMaxViolations
// ---------------------------------------------------------------------------

func TestCountMaxViolations_NoViolations(t *testing.T) {
	p1, p2, p3, p4 := newPID(), newPID(), newPID(), newPID()
	matches := []pair{{p1, p2}, {p3, p4}}
	assignment := []int{0, 1}
	v := countMaxViolations(matches, assignment, 2, map[domain.PlayerID]bool{})
	if v != 0 {
		t.Errorf("want 0 violations, got %d", v)
	}
}

func TestCountMaxViolations_RegularPlayerViolation(t *testing.T) {
	p1, p2, p3, p4 := newPID(), newPID(), newPID(), newPID()
	// p1 has 4 matches on evening 0 → exceeds limit of 3 by 1
	matches := []pair{{p1, p2}, {p1, p3}, {p1, p4}, {p1, p2}}
	assignment := []int{0, 0, 0, 0}
	v := countMaxViolations(matches, assignment, 1, map[domain.PlayerID]bool{})
	if v != 1 {
		t.Errorf("want 1 violation (4 matches, limit 3), got %d", v)
	}
}

func TestCountMaxViolations_BuddyPlayerAtLimit(t *testing.T) {
	p1, p2, p3, p4 := newPID(), newPID(), newPID(), newPID()
	// Buddy player p1 has 4 matches on evening 0 → exactly at the buddy limit of 4, no violation
	matches := []pair{{p1, p2}, {p1, p3}, {p1, p4}, {p1, p2}}
	assignment := []int{0, 0, 0, 0}
	buddyPlayers := map[domain.PlayerID]bool{p1: true}
	v := countMaxViolations(matches, assignment, 1, buddyPlayers)
	if v != 0 {
		t.Errorf("want 0 violations for buddy player at limit of 4, got %d", v)
	}
}

func TestCountMaxViolations_BuddyPlayerExcess(t *testing.T) {
	p1, p2, p3, p4, p5 := newPID(), newPID(), newPID(), newPID(), newPID()
	// Buddy player p1 has 5 matches on evening 0 → exceeds buddy limit of 4 by 1
	matches := []pair{{p1, p2}, {p1, p3}, {p1, p4}, {p1, p5}, {p1, p2}}
	assignment := []int{0, 0, 0, 0, 0}
	buddyPlayers := map[domain.PlayerID]bool{p1: true}
	v := countMaxViolations(matches, assignment, 1, buddyPlayers)
	if v != 1 {
		t.Errorf("want 1 violation for buddy player with 5 matches, got %d", v)
	}
}

// ---------------------------------------------------------------------------
// countBuddyHardViolations
// ---------------------------------------------------------------------------

func TestCountBuddyHardViolations_NoViolations(t *testing.T) {
	p1, p2, p3, p4 := newPID(), newPID(), newPID(), newPID()
	// p1 and p2 both play on the same evening → no mismatch
	matches := []pair{{p1, p3}, {p2, p4}}
	assignment := []int{0, 0}
	buddyPairs := []domain.BuddyPreference{{PlayerID: p1, BuddyID: p2}}
	v := countBuddyHardViolations(matches, assignment, buddyPairs, 1)
	if v != 0 {
		t.Errorf("want 0 violations, got %d", v)
	}
}

func TestCountBuddyHardViolations_Mismatch(t *testing.T) {
	p1, p2, p3, p4 := newPID(), newPID(), newPID(), newPID()
	// p1 plays evening 0; p2 plays evening 1 → mismatch on both evenings = 2 violations
	matches := []pair{{p1, p3}, {p2, p4}}
	assignment := []int{0, 1}
	buddyPairs := []domain.BuddyPreference{{PlayerID: p1, BuddyID: p2}}
	v := countBuddyHardViolations(matches, assignment, buddyPairs, 2)
	if v != 2 {
		t.Errorf("want 2 violations, got %d", v)
	}
}

func TestCountBuddyHardViolations_BidirectionalCountedOnce(t *testing.T) {
	p1, p2, p3, p4 := newPID(), newPID(), newPID(), newPID()
	// Bidirectional buddy pair (p1→p2 and p2→p1): each unique pair counted only once
	matches := []pair{{p1, p3}, {p2, p4}}
	assignment := []int{0, 1}
	buddyPairs := []domain.BuddyPreference{
		{PlayerID: p1, BuddyID: p2},
		{PlayerID: p2, BuddyID: p1},
	}
	v := countBuddyHardViolations(matches, assignment, buddyPairs, 2)
	if v != 2 {
		t.Errorf("want 2 violations (bidirectional pair counted once), got %d", v)
	}
}

func TestCountBuddyHardViolations_NoBuddyPairs(t *testing.T) {
	p1, p2 := newPID(), newPID()
	matches := []pair{{p1, p2}}
	assignment := []int{0}
	v := countBuddyHardViolations(matches, assignment, nil, 1)
	if v != 0 {
		t.Errorf("want 0 violations with no buddy pairs, got %d", v)
	}
}

// ---------------------------------------------------------------------------
// countTripleConsecutiveViolations
// ---------------------------------------------------------------------------

func TestCountTripleConsecutiveViolations_NoViolations(t *testing.T) {
	p1, p2 := newPID(), newPID()
	// p1 plays exactly 2 consecutive evenings → within the limit
	matches := []pair{{p1, p2}, {p1, p2}}
	assignment := []int{0, 1}
	v := countTripleConsecutiveViolations(matches, assignment, 3)
	if v != 0 {
		t.Errorf("want 0 violations for run of 2, got %d", v)
	}
}

func TestCountTripleConsecutiveViolations_RunOf3(t *testing.T) {
	p1, p2, p3, p4 := newPID(), newPID(), newPID(), newPID()
	// p1 plays 3 consecutive evenings against different opponents → only p1 has the run
	// run exceeds limit of 2 at position 2 → 1 violation from p1
	matches := []pair{{p1, p2}, {p1, p3}, {p1, p4}}
	assignment := []int{0, 1, 2}
	v := countTripleConsecutiveViolations(matches, assignment, 3)
	if v != 1 {
		t.Errorf("want 1 violation for run of 3, got %d", v)
	}
}

func TestCountTripleConsecutiveViolations_RunOf4(t *testing.T) {
	p1, p2, p3, p4, p5 := newPID(), newPID(), newPID(), newPID(), newPID()
	// p1 plays 4 consecutive evenings against different opponents → only p1 has the run
	// violations at positions 2 and 3 → 2 violations from p1
	matches := []pair{{p1, p2}, {p1, p3}, {p1, p4}, {p1, p5}}
	assignment := []int{0, 1, 2, 3}
	v := countTripleConsecutiveViolations(matches, assignment, 4)
	if v != 2 {
		t.Errorf("want 2 violations for run of 4, got %d", v)
	}
}

func TestCountTripleConsecutiveViolations_BreakResets(t *testing.T) {
	p1, p2, p3, p4, p5 := newPID(), newPID(), newPID(), newPID(), newPID()
	// p1 plays evenings 0,1 then skips 2, then plays 3,4 → two separate runs of 2, no violations
	matches := []pair{{p1, p2}, {p1, p3}, {p1, p4}, {p1, p5}}
	assignment := []int{0, 1, 3, 4}
	v := countTripleConsecutiveViolations(matches, assignment, 5)
	if v != 0 {
		t.Errorf("want 0 violations when runs are broken by gap, got %d", v)
	}
}

// ---------------------------------------------------------------------------
// countExcessTripleMatchViolations
// ---------------------------------------------------------------------------

func TestCountExcessTripleMatchViolations_NoViolations(t *testing.T) {
	p1, p2, p3 := newPID(), newPID(), newPID()
	// p1 active on 10 evenings; exactly 1 evening with 3 matches → allowed = ceil(0.1*10)=1, no excess
	var matches []pair
	var assignment []int
	matches = append(matches, pair{p1, p2}, pair{p1, p3}, pair{p1, p2}) // evening 0: 3 matches
	assignment = append(assignment, 0, 0, 0)
	for i := 1; i < 10; i++ { // evenings 1-9: 1 match each
		matches = append(matches, pair{p1, p2})
		assignment = append(assignment, i)
	}
	v := countExcessTripleMatchViolations(matches, assignment, 10)
	if v != 0 {
		t.Errorf("want 0 violations (1 triple out of 10 active, allowed=1), got %d", v)
	}
}

func TestCountExcessTripleMatchViolations_ExcessViolation(t *testing.T) {
	p1, p2, p3 := newPID(), newPID(), newPID()
	// p1 active on 10 evenings; 2 evenings with 3 matches → allowed=1, excess=1
	var matches []pair
	var assignment []int
	for _, ei := range []int{0, 1} { // evenings 0 and 1: 3 matches each
		matches = append(matches, pair{p1, p2}, pair{p1, p3}, pair{p1, p2})
		assignment = append(assignment, ei, ei, ei)
	}
	for i := 2; i < 10; i++ { // evenings 2-9: 1 match each
		matches = append(matches, pair{p1, p2})
		assignment = append(assignment, i)
	}
	v := countExcessTripleMatchViolations(matches, assignment, 10)
	if v != 1 {
		t.Errorf("want 1 violation (2 triples out of 10 active, allowed=1), got %d", v)
	}
}

// ---------------------------------------------------------------------------
// varianceMatchesPerEvening
// ---------------------------------------------------------------------------

func TestVarianceMatchesPerEvening_Zero(t *testing.T) {
	// Perfectly balanced: 1 match per evening → variance = 0
	v := varianceMatchesPerEvening([]int{0, 1}, 2)
	if v != 0 {
		t.Errorf("want variance 0 for balanced assignment, got %f", v)
	}
}

func TestVarianceMatchesPerEvening_NonZero(t *testing.T) {
	// 2 matches on evening 0, 0 on evening 1 → mean=1, variance=((2-1)²+(0-1)²)/2=1
	v := varianceMatchesPerEvening([]int{0, 0}, 2)
	if v != 1.0 {
		t.Errorf("want variance 1.0 for unbalanced assignment, got %f", v)
	}
}

func TestVarianceMatchesPerEvening_AllSameEvening(t *testing.T) {
	// 4 matches all on evening 0 of 4 evenings → counts=[4,0,0,0], mean=1, var=((3²+1+1+1)/4)=3
	v := varianceMatchesPerEvening([]int{0, 0, 0, 0}, 4)
	if v != 3.0 {
		t.Errorf("want variance 3.0, got %f", v)
	}
}

// ---------------------------------------------------------------------------
// greedyAssign
// ---------------------------------------------------------------------------

func TestGreedyAssign_EvenDistribution(t *testing.T) {
	// 6 matches across 3 evenings: each gets exactly 2
	assignment := greedyAssign(make([]pair, 6), 3)
	counts := make([]int, 3)
	for _, ei := range assignment {
		counts[ei]++
	}
	for i, c := range counts {
		if c != 2 {
			t.Errorf("evening %d: want 2 matches, got %d", i, c)
		}
	}
}

func TestGreedyAssign_UnevenDistribution(t *testing.T) {
	// 7 matches across 3 evenings: first evening gets 3, others get 2
	assignment := greedyAssign(make([]pair, 7), 3)
	counts := make([]int, 3)
	for _, ei := range assignment {
		counts[ei]++
	}
	total := 0
	for _, c := range counts {
		total += c
		if c < 2 || c > 3 {
			t.Errorf("evening count %d outside expected range [2,3]", c)
		}
	}
	if total != 7 {
		t.Errorf("want 7 total assignments, got %d", total)
	}
}

func TestGreedyAssign_AllMatchesCovered(t *testing.T) {
	n := 15
	assignment := greedyAssign(make([]pair, n), 4)
	if len(assignment) != n {
		t.Errorf("want assignment length %d, got %d", n, len(assignment))
	}
	for i, ei := range assignment {
		if ei < 0 || ei >= 4 {
			t.Errorf("assignment[%d]=%d out of range [0,4)", i, ei)
		}
	}
}

// ---------------------------------------------------------------------------
// countMinMatchViolations
// ---------------------------------------------------------------------------

func TestCountMinMatchViolations_NoViolations(t *testing.T) {
	p1, p2, p3, p4 := newPID(), newPID(), newPID(), newPID()
	// Each player has 2 matches on evening 0 → no solo evenings → no violations.
	matches := []pair{{p1, p2}, {p3, p4}, {p1, p3}, {p2, p4}}
	assignment := []int{0, 0, 0, 0}
	v := countMinMatchViolations(matches, assignment, 1)
	if v != 0 {
		t.Errorf("want 0 violations, got %d", v)
	}
}

func TestCountMinMatchViolations_OneSoloAllowed(t *testing.T) {
	p1, p2 := newPID(), newPID()
	// p1 and p2 each have exactly 1 solo evening (ev0) and 2 matches on ev1.
	// 1 solo evening is allowed → 0 violations.
	matches := []pair{{p1, p2}, {p1, p2}, {p1, p2}}
	assignment := []int{0, 1, 1}
	v := countMinMatchViolations(matches, assignment, 2)
	if v != 0 {
		t.Errorf("want 0 violations (1 solo allowed), got %d", v)
	}
}

func TestCountMinMatchViolations_ExcessSoloEvenings(t *testing.T) {
	p1, p2 := newPID(), newPID()
	// p1 and p2 each appear on 3 separate evenings with 1 match each.
	// 3 solo evenings − 1 allowed = 2 violations per player → 4 total.
	matches := []pair{{p1, p2}, {p1, p2}, {p1, p2}}
	assignment := []int{0, 1, 2}
	v := countMinMatchViolations(matches, assignment, 3)
	if v != 4 {
		t.Errorf("want 4 violations (2 excess solo evenings × 2 players), got %d", v)
	}
}

func TestCountMinMatchViolations_MixedPlayers(t *testing.T) {
	p1, p2, p3 := newPID(), newPID(), newPID()
	// p1: 2 solo evenings (ev0, ev1) → 1 violation
	// p2: 1 solo evening (ev0) → 0 violations
	// p3: 2 matches on ev1, 1 on ev2 → 1 solo evening → 0 violations
	matches := []pair{{p1, p2}, {p1, p3}, {p3, p2}, {p1, p3}}
	// ev0: p1 vs p2 (p1:1, p2:1)
	// ev1: p1 vs p3 + p3 vs p2 (p1:1, p3:2, p2:1 — but p2 already on ev0 solo)
	// ev2: p1 vs p3 (p1:1, p3:1)
	assignment := []int{0, 1, 1, 2}
	// p1: ev0→1, ev1→1, ev2→1 → 3 solo evenings → 2 violations
	// p2: ev0→1, ev1→1 → 2 solo evenings → 1 violation
	// p3: ev1→2, ev2→1 → 1 solo evening → 0 violations
	v := countMinMatchViolations(matches, assignment, 3)
	if v != 3 {
		t.Errorf("want 3 violations, got %d", v)
	}
}
