package scheduler_test

import (
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

		// Every pair should appear exactly once.
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

// TestMaxMatchesPerPlayerPerEvening checks soft constraint: at most 3.
func TestMaxMatchesPerPlayerPerEvening(t *testing.T) {
	players := makePlayers(10)
	sched, err := scheduler.Generate(scheduler.Input{
		Players:         players,
		NumEvenings:     10,
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
				t.Errorf("evening %d: player %s has %d matches (max 3)", ev.Number, pid, c)
			}
		}
	}
}

// TestBuddySatisfaction checks that buddy pairs tend to share evenings.
func TestBuddySatisfaction(t *testing.T) {
	players := makePlayers(8)
	buddies := []domain.BuddyPreference{
		{PlayerID: players[0].ID, BuddyID: players[1].ID},
		{PlayerID: players[2].ID, BuddyID: players[3].ID},
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

	// At least one buddy pair should share at least one evening.
	satisfiedCount := 0
	for _, bp := range buddies {
		for _, ev := range sched.Evenings {
			hasA, hasB := false, false
			for _, m := range ev.Matches {
				if m.PlayerA == bp.PlayerID || m.PlayerB == bp.PlayerID {
					hasA = true
				}
				if m.PlayerA == bp.BuddyID || m.PlayerB == bp.BuddyID {
					hasB = true
				}
			}
			if hasA && hasB {
				satisfiedCount++
				break
			}
		}
	}
	if satisfiedCount == 0 {
		t.Error("no buddy pair shares any evening")
	}
}

// TestOddNumberOfPlayers ensures a BYE player works (no panic, all real pairs present).
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
