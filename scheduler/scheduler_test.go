package scheduler_test

import (
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
// Hard constraint: buddy pairs MUST share ALL evenings
// ---------------------------------------------------------------------------

func TestBuddyPairsMustShareAllEvenings(t *testing.T) {
	// Use a smaller case so the annealing converges in time.
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

	// Each unique buddy pair must play together on EVERY evening one of them plays.
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

		for i, ep := range eveningPlayers {
			aPlays := ep[bp.PlayerID]
			bPlays := ep[bp.BuddyID]
			if aPlays != bPlays {
				t.Errorf("buddy pair %s — %s: evening %d: one plays (%v) but the other doesn't (%v)",
					bp.PlayerID, bp.BuddyID, sched.Evenings[i].Number, aPlays, bPlays)
			}
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

	// 1. Buddy pairs share ALL evenings (neither plays without the other).
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
		for i, ep := range eveningPlayers {
			aPlays := ep[bp.PlayerID]
			bPlays := ep[bp.BuddyID]
			if aPlays != bPlays {
				t.Errorf("(combined) buddy pair %s — %s: evening %d: one plays (%v) but the other doesn't (%v)",
					bp.PlayerID, bp.BuddyID, sched.Evenings[i].Number, aPlays, bPlays)
			}
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
	_, err := scheduler.Generate(scheduler.Input{
		Players:         players,
		NumEvenings:     5,
		CompetitionName: "Test",
		StartDate:       time.Now(),
		IntervalDays:    7,
	})
	if err != nil {
		t.Fatalf("odd players: %v", err)
	}
}
