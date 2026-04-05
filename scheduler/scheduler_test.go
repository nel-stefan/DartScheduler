package scheduler_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"DartScheduler/domain"
	"DartScheduler/scheduler"

	"github.com/google/uuid"
)

func makePlayers(n int) []domain.Player {
	players := make([]domain.Player, n)
	for i := range players {
		players[i] = domain.Player{ID: uuid.New(), Name: "P" + string(rune('A'+i))}
	}
	return players
}

// fastConfig returns an annealing config with fewer steps for use in tests.
// Hard constraints still converge for small schedules (≤12 players).
func fastConfig() scheduler.AnnealConfig {
	cfg := scheduler.DefaultAnnealConfig()
	cfg.Steps = 200_000
	return cfg
}

// ---------------------------------------------------------------------------
// Input validation errors
// ---------------------------------------------------------------------------

func TestGenerateTooFewPlayersError(t *testing.T) {
	_, err := scheduler.Generate(scheduler.Input{
		Players:     makePlayers(1),
		NumEvenings: 1,
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("want ErrInvalidInput for 1 player, got %v", err)
	}
}

func TestGenerateZeroPlayersError(t *testing.T) {
	_, err := scheduler.Generate(scheduler.Input{
		Players:     makePlayers(0),
		NumEvenings: 1,
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("want ErrInvalidInput for 0 players, got %v", err)
	}
}

func TestGenerateZeroEveningsError(t *testing.T) {
	_, err := scheduler.Generate(scheduler.Input{
		Players:     makePlayers(4),
		NumEvenings: 0,
	})
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("want ErrInvalidInput for 0 evenings, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// EveningDates parameter
// ---------------------------------------------------------------------------

func TestEveningDatesUsed(t *testing.T) {
	players := makePlayers(4)
	dates := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	sched, err := scheduler.Generate(scheduler.Input{
		Players:         players,
		NumEvenings:     3,
		CompetitionName: "Test",
		EveningDates:    dates,
		Config:          fastConfig(),
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if len(sched.Evenings) != 3 {
		t.Fatalf("want 3 evenings, got %d", len(sched.Evenings))
	}
	for i, ev := range sched.Evenings {
		if !ev.Date.Equal(dates[i]) {
			t.Errorf("evening %d: want date %v, got %v", i, dates[i], ev.Date)
		}
	}
}

func TestEveningNumbersAreSequential(t *testing.T) {
	players := makePlayers(4)
	sched, err := scheduler.Generate(scheduler.Input{
		Players:         players,
		NumEvenings:     3,
		CompetitionName: "Test",
		StartDate:       time.Now(),
		IntervalDays:    7,
		Config:          fastConfig(),
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	for i, ev := range sched.Evenings {
		if ev.Number != i+1 {
			t.Errorf("evening %d: want Number=%d, got %d", i, i+1, ev.Number)
		}
	}
}

func TestAllMatchesHaveValidEveningID(t *testing.T) {
	players := makePlayers(6)
	sched, err := scheduler.Generate(scheduler.Input{
		Players:         players,
		NumEvenings:     5,
		CompetitionName: "Test",
		StartDate:       time.Now(),
		IntervalDays:    7,
		Config:          fastConfig(),
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	eveningIDs := make(map[domain.EveningID]bool, len(sched.Evenings))
	for _, ev := range sched.Evenings {
		eveningIDs[ev.ID] = true
	}
	for _, ev := range sched.Evenings {
		for _, m := range ev.Matches {
			if !eveningIDs[m.EveningID] {
				t.Errorf("match %s has EveningID %s not in schedule", m.ID, m.EveningID)
			}
			if m.EveningID != ev.ID {
				t.Errorf("match stored in evening %s but EveningID=%s", ev.ID, m.EveningID)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Round-robin correctness
// ---------------------------------------------------------------------------

// TestEveryPairAppearsExactlyOnce verifies the round-robin property.
func TestEveryPairAppearsExactlyOnce(t *testing.T) {
	for _, n := range []int{4, 6, 8, 10} {
		players := makePlayers(n)
		sched, err := scheduler.Generate(scheduler.Input{
			Players:         players,
			NumEvenings:     n - 1,
			CompetitionName: "Test",
			StartDate:       time.Now(),
			IntervalDays:    7,
			Config:          fastConfig(),
		})
		if err != nil {
			t.Fatalf("n=%d: Generate error: %v", n, err)
		}

		type pair struct{ a, b domain.PlayerID }
		seen := map[pair]int{}
		for _, ev := range sched.Evenings {
			for _, m := range ev.Matches {
				key := pair{m.PlayerA, m.PlayerB}
				if m.PlayerA.String() > m.PlayerB.String() {
					key = pair{m.PlayerB, m.PlayerA}
				}
				seen[key]++
			}
		}

		expected := n * (n - 1) / 2
		if len(seen) != expected {
			t.Errorf("n=%d: expected %d unique pairs, got %d", n, expected, len(seen))
		}
		for k, c := range seen {
			if c != 1 {
				t.Errorf("n=%d: pair %v appears %d times", n, k, c)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Hard constraint: max matches per player per evening (ceiling = 4)
// ---------------------------------------------------------------------------

func TestMaxMatchesPerPlayerPerEvening(t *testing.T) {
	// No buddy pairs: all players are regular → max 3 matches per evening.
	players := makePlayers(12)
	sched, err := scheduler.Generate(scheduler.Input{
		Players:         players,
		NumEvenings:     11,
		CompetitionName: "Test",
		StartDate:       time.Now(),
		IntervalDays:    7,
		Config:          fastConfig(),
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, ev := range sched.Evenings {
		counts := map[domain.PlayerID]int{}
		for _, m := range ev.Matches {
			counts[m.PlayerA]++
			counts[m.PlayerB]++
		}
		for pid, c := range counts {
			if c > 3 {
				t.Errorf("evening %d: non-buddy player %s has %d matches (hard max 3)", ev.Number, pid, c)
			}
		}
	}
}

func TestBuddyPlayersAllowedFourMatchesNonBuddyNotAllowed(t *testing.T) {
	players := makePlayers(8)
	buddies := []domain.BuddyPreference{
		{PlayerID: players[0].ID, BuddyID: players[1].ID},
		{PlayerID: players[1].ID, BuddyID: players[0].ID},
	}
	buddySet := map[domain.PlayerID]bool{players[0].ID: true, players[1].ID: true}

	sched, err := scheduler.Generate(scheduler.Input{
		Players:         players,
		BuddyPairs:      buddies,
		NumEvenings:     7,
		CompetitionName: "Test",
		StartDate:       time.Now(),
		IntervalDays:    7,
		Config:          fastConfig(),
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, ev := range sched.Evenings {
		counts := map[domain.PlayerID]int{}
		for _, m := range ev.Matches {
			counts[m.PlayerA]++
			counts[m.PlayerB]++
		}
		for pid, c := range counts {
			if buddySet[pid] {
				if c > 4 {
					t.Errorf("evening %d: buddy player %s has %d matches (max 4)", ev.Number, pid, c)
				}
			} else {
				if c > 3 {
					t.Errorf("evening %d: non-buddy player %s has %d matches (max 3)", ev.Number, pid, c)
				}
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Hard constraint: buddy pairs have at most 1 mismatch evening
// ---------------------------------------------------------------------------

func TestBuddyPairsAtMostOneMismatch(t *testing.T) {
	// 8 players × 7 evenings → 28 matches, ~4/evening, each player active ~4 evenings.
	players := makePlayers(8)
	buddies := []domain.BuddyPreference{
		{PlayerID: players[0].ID, BuddyID: players[1].ID},
		{PlayerID: players[1].ID, BuddyID: players[0].ID}, // bidirectional
		{PlayerID: players[2].ID, BuddyID: players[3].ID},
		{PlayerID: players[3].ID, BuddyID: players[2].ID},
	}
	sched, err := scheduler.Generate(scheduler.Input{
		Players:         players,
		BuddyPairs:      buddies,
		NumEvenings:     7,
		CompetitionName: "Test",
		StartDate:       time.Now(),
		IntervalDays:    7,
		Config:          fastConfig(),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Build per-evening player set.
	eveningPlayers := make([]map[domain.PlayerID]bool, len(sched.Evenings))
	for i, ev := range sched.Evenings {
		eveningPlayers[i] = make(map[domain.PlayerID]bool)
		for _, m := range ev.Matches {
			eveningPlayers[i][m.PlayerA] = true
			eveningPlayers[i][m.PlayerB] = true
		}
	}

	// Each unique buddy pair may have at most 1 evening where one plays without the other.
	checked := map[[2]string]bool{}
	for _, bp := range buddies {
		a, b := bp.PlayerID.String(), bp.BuddyID.String()
		key := [2]string{a, b}
		if a > b {
			key = [2]string{b, a}
		}
		if checked[key] {
			continue
		}
		checked[key] = true

		mismatches := 0
		for _, ep := range eveningPlayers {
			if ep[bp.PlayerID] != ep[bp.BuddyID] {
				mismatches++
			}
		}
		if mismatches > 1 {
			t.Errorf("buddy pair %s — %s: %d mismatch evenings (hard max 1)",
				bp.PlayerID, bp.BuddyID, mismatches)
		}
	}
}

// ---------------------------------------------------------------------------
// Hard constraint: no player plays more than 2 consecutive evenings
// ---------------------------------------------------------------------------

func TestNoMoreThanTwoConsecutiveEvenings(t *testing.T) {
	players := makePlayers(12)
	sched, err := scheduler.Generate(scheduler.Input{
		Players:         players,
		NumEvenings:     11,
		CompetitionName: "Test",
		StartDate:       time.Now(),
		IntervalDays:    7,
		Config:          fastConfig(),
	})
	if err != nil {
		t.Fatal(err)
	}

	eveningPlayers := make([]map[domain.PlayerID]bool, len(sched.Evenings))
	for i, ev := range sched.Evenings {
		eveningPlayers[i] = make(map[domain.PlayerID]bool)
		for _, m := range ev.Matches {
			eveningPlayers[i][m.PlayerA] = true
			eveningPlayers[i][m.PlayerB] = true
		}
	}

	allPlayers := make(map[domain.PlayerID]bool)
	for _, ep := range eveningPlayers {
		for pid := range ep {
			allPlayers[pid] = true
		}
	}

	for pid := range allPlayers {
		run := 0
		for i, ep := range eveningPlayers {
			if ep[pid] {
				run++
				if run > 2 {
					t.Errorf("player %s plays 3+ consecutive evenings (run ends at evening index %d)", pid, i)
				}
			} else {
				run = 0
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Soft constraint: at most 10% of active evenings have 3 matches per player
// ---------------------------------------------------------------------------

func TestAtMostTenPercentTripleMatchEvenings(t *testing.T) {
	players := makePlayers(12)
	sched, err := scheduler.Generate(scheduler.Input{
		Players:         players,
		NumEvenings:     11,
		CompetitionName: "Test",
		StartDate:       time.Now(),
		IntervalDays:    7,
		Config:          fastConfig(),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Count active evenings and triple-match evenings per player.
	active := map[domain.PlayerID]int{}
	triple := map[domain.PlayerID]int{}
	for _, ev := range sched.Evenings {
		counts := map[domain.PlayerID]int{}
		for _, m := range ev.Matches {
			counts[m.PlayerA]++
			counts[m.PlayerB]++
		}
		for pid, c := range counts {
			active[pid]++
			if c >= 3 {
				triple[pid]++
			}
		}
	}

	for pid, act := range active {
		tri := triple[pid]
		allowed := int(math.Ceil(0.1 * float64(act)))
		if tri > allowed {
			t.Errorf("player %s: %d/%d active evenings have 3 matches (allowed %d, >10%%)",
				pid, tri, act, allowed)
		}
	}
}

// ---------------------------------------------------------------------------
// Hard constraint with buddy pairs: combined test (all hard rules)
// ---------------------------------------------------------------------------

func TestAllHardConstraintsWithBuddies(t *testing.T) {
	// 10 players × 9 evenings: manageable size for full hard-constraint verification.
	players := makePlayers(10)
	buddies := []domain.BuddyPreference{
		{PlayerID: players[0].ID, BuddyID: players[1].ID},
		{PlayerID: players[1].ID, BuddyID: players[0].ID},
		{PlayerID: players[2].ID, BuddyID: players[3].ID},
		{PlayerID: players[3].ID, BuddyID: players[2].ID},
	}
	sched, err := scheduler.Generate(scheduler.Input{
		Players:         players,
		BuddyPairs:      buddies,
		NumEvenings:     9,
		CompetitionName: "Test",
		StartDate:       time.Now(),
		IntervalDays:    7,
		Config:          fastConfig(),
	})
	if err != nil {
		t.Fatal(err)
	}

	eveningPlayers := make([]map[domain.PlayerID]bool, len(sched.Evenings))
	for i, ev := range sched.Evenings {
		eveningPlayers[i] = make(map[domain.PlayerID]bool)
		for _, m := range ev.Matches {
			eveningPlayers[i][m.PlayerA] = true
			eveningPlayers[i][m.PlayerB] = true
		}
	}

	// 1. Buddy pairs have at most 1 mismatch evening.
	checked := map[[2]string]bool{}
	for _, bp := range buddies {
		a, b := bp.PlayerID.String(), bp.BuddyID.String()
		key := [2]string{a, b}
		if a > b {
			key = [2]string{b, a}
		}
		if checked[key] {
			continue
		}
		checked[key] = true
		mismatches := 0
		for _, ep := range eveningPlayers {
			if ep[bp.PlayerID] != ep[bp.BuddyID] {
				mismatches++
			}
		}
		if mismatches > 1 {
			t.Errorf("(combined) buddy pair %s — %s: %d mismatch evenings (hard max 1)",
				bp.PlayerID, bp.BuddyID, mismatches)
		}
	}

	// 2. No player plays 3+ consecutive evenings.
	allPlayers := make(map[domain.PlayerID]bool)
	for _, ep := range eveningPlayers {
		for pid := range ep {
			allPlayers[pid] = true
		}
	}
	for pid := range allPlayers {
		run := 0
		for i, ep := range eveningPlayers {
			if ep[pid] {
				run++
				if run > 2 {
					t.Errorf("(combined) player %s plays 3+ consecutive evenings at index %d", pid, i)
				}
			} else {
				run = 0
			}
		}
	}

	// 3. Max matches per player per evening: buddy players ≤ 4, others ≤ 3.
	buddySet := map[domain.PlayerID]bool{}
	for _, bp := range buddies {
		buddySet[bp.PlayerID] = true
		buddySet[bp.BuddyID] = true
	}
	for _, ev := range sched.Evenings {
		cnt := map[domain.PlayerID]int{}
		for _, m := range ev.Matches {
			cnt[m.PlayerA]++
			cnt[m.PlayerB]++
		}
		for pid, c := range cnt {
			limit := 3
			if buddySet[pid] {
				limit = 4
			}
			if c > limit {
				t.Errorf("(combined) evening %d: player %s has %d matches (max %d)", ev.Number, pid, c, limit)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Odd number of players (BYE handling)
// ---------------------------------------------------------------------------

func TestOddNumberOfPlayers(t *testing.T) {
	players := makePlayers(5)
	// 5 players / 5 evenings forces 1 match/active evening for every player (infeasible for
	// the solo-evening hard constraint). Use 2 evenings: avg 5 matches/evening allows ≥2 per player.
	_, err := scheduler.Generate(scheduler.Input{
		Players:         players,
		NumEvenings:     2,
		CompetitionName: "Test",
		StartDate:       time.Now(),
		IntervalDays:    7,
		Config:          fastConfig(),
	})
	if err != nil {
		t.Fatalf("odd players: %v", err)
	}
}

// TestAtMostOneSoloEveningPerPlayer verifies the hard constraint: each player
// may have at most 1 evening with exactly 1 match. Uses the realistic scenario
// of 20 players across 30 evenings (N-1=19 matches per player; max 9 active
// evenings each, so clustering into 2-match evenings is mathematically feasible).
// Skipped in short mode because it runs the full default annealing (~1.2M steps).
func TestAtMostOneSoloEveningPerPlayer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large integration test in short mode")
	}
	players := make([]domain.Player, 20)
	for i := range players {
		players[i] = domain.Player{ID: domain.PlayerID(uuid.New())}
	}
	sched, err := scheduler.Generate(scheduler.Input{
		Players:      players,
		NumEvenings:  30,
		StartDate:    time.Now(),
		IntervalDays: 7,
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Count solo evenings per player.
	soloPerPlayer := make(map[domain.PlayerID]int)
	for _, ev := range sched.Evenings {
		matchCount := make(map[domain.PlayerID]int)
		for _, m := range ev.Matches {
			matchCount[m.PlayerA]++
			matchCount[m.PlayerB]++
		}
		for pid, c := range matchCount {
			if c == 1 {
				soloPerPlayer[pid]++
			}
		}
	}

	for pid, solo := range soloPerPlayer {
		if solo > 1 {
			t.Errorf("player %v has %d solo evenings (max 1 allowed)", pid, solo)
		}
	}
}

// ---------------------------------------------------------------------------
// Hard constraint: max gap between a player's consecutive active evenings
// ---------------------------------------------------------------------------

func TestMaxGapBetweenActiveEvenings(t *testing.T) {
	// 12 players × 11 evenings: every player plays on every evening, so gap = 1 always.
	// This also covers the realistic scenario where the constraint is trivially satisfied.
	players := makePlayers(12)
	sched, err := scheduler.Generate(scheduler.Input{
		Players:         players,
		NumEvenings:     11,
		CompetitionName: "Test",
		StartDate:       time.Now(),
		IntervalDays:    7,
		Config:          fastConfig(),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Build per-player list of active evening indices (sorted).
	eveningIndex := make(map[domain.EveningID]int, len(sched.Evenings))
	for i, ev := range sched.Evenings {
		eveningIndex[ev.ID] = i
	}

	activeEvenings := make(map[domain.PlayerID][]int)
	for _, ev := range sched.Evenings {
		seen := make(map[domain.PlayerID]bool)
		for _, m := range ev.Matches {
			for _, pid := range []domain.PlayerID{m.PlayerA, m.PlayerB} {
				if !seen[pid] {
					activeEvenings[pid] = append(activeEvenings[pid], eveningIndex[ev.ID])
					seen[pid] = true
				}
			}
		}
	}

	for pid, indices := range activeEvenings {
		for i := 1; i < len(indices); i++ {
			gap := indices[i] - indices[i-1]
			if gap > 4 { // maxGapBetweenActiveEvenings = 4
				t.Errorf("player %s: gap of %d between evening indices %d and %d (max 4)",
					pid, gap, indices[i-1], indices[i])
			}
		}
	}
}
