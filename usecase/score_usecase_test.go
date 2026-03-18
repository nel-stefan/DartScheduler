package usecase_test

import (
	"context"
	"testing"

	"DartScheduler/domain"
	"DartScheduler/usecase"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// In-memory MatchRepository stub (only the methods used by ScoreUseCase)
// ---------------------------------------------------------------------------

type stubMatchRepo struct {
	matches map[domain.MatchID]domain.Match
}

func newStubMatchRepo(matches []domain.Match) *stubMatchRepo {
	r := &stubMatchRepo{matches: make(map[domain.MatchID]domain.Match, len(matches))}
	for _, m := range matches {
		r.matches[m.ID] = m
	}
	return r
}

func (r *stubMatchRepo) FindByEvening(_ context.Context, eveningID domain.EveningID) ([]domain.Match, error) {
	var out []domain.Match
	for _, m := range r.matches {
		if m.EveningID == eveningID {
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *stubMatchRepo) UpdateResult(_ context.Context, m domain.Match) error {
	r.matches[m.ID] = m
	return nil
}

func (r *stubMatchRepo) FindByID(_ context.Context, id domain.MatchID) (domain.Match, error) {
	m, ok := r.matches[id]
	if !ok {
		return domain.Match{}, domain.ErrNotFound
	}
	return m, nil
}

// Unused methods — satisfy the interface with no-ops.
func (r *stubMatchRepo) Save(_ context.Context, m domain.Match) error                                          { return nil }
func (r *stubMatchRepo) SaveBatch(_ context.Context, _ []domain.Match) error                                   { return nil }
func (r *stubMatchRepo) FindByPlayer(_ context.Context, _ domain.PlayerID) ([]domain.Match, error)             { return nil, nil }
func (r *stubMatchRepo) FindByPlayerAndSchedule(_ context.Context, _ domain.PlayerID, _ domain.ScheduleID) ([]domain.Match, error) {
	return nil, nil
}
func (r *stubMatchRepo) FindAllPlayed(_ context.Context) ([]domain.Match, error) { return nil, nil }
func (r *stubMatchRepo) FindBySchedule(_ context.Context, _ domain.ScheduleID) ([]domain.Match, error) {
	return nil, nil
}
func (r *stubMatchRepo) FindCancelledBySchedule(_ context.Context, _ domain.ScheduleID) ([]domain.Match, error) {
	return nil, nil
}
func (r *stubMatchRepo) DeleteByEvening(_ context.Context, _ domain.EveningID) error   { return nil }
func (r *stubMatchRepo) DeleteBySchedule(_ context.Context, _ domain.ScheduleID) error { return nil }
func (r *stubMatchRepo) DeleteByPlayer(_ context.Context, _ domain.PlayerID) error     { return nil }

// ---------------------------------------------------------------------------
// ReportAbsent tests
// ---------------------------------------------------------------------------

func makeMatch(eveningID domain.EveningID, pA, pB domain.PlayerID, played bool) domain.Match {
	m := domain.Match{
		ID:        domain.MatchID(uuid.New()),
		EveningID: eveningID,
		PlayerA:   pA,
		PlayerB:   pB,
		Played:    played,
	}
	if played {
		s := 2
		m.ScoreA = &s
		m.ScoreB = new(int)
	}
	return m
}

// TestReportAbsent_OnlyUnplayedMatchesAreMarked verifies that already-played
// matches are left untouched when a player is reported absent.
func TestReportAbsent_OnlyUnplayedMatchesAreMarked(t *testing.T) {
	ctx := context.Background()

	eveningID := domain.EveningID(uuid.New())
	player := domain.PlayerID(uuid.New())
	opponent1 := domain.PlayerID(uuid.New())
	opponent2 := domain.PlayerID(uuid.New())
	unrelated := domain.PlayerID(uuid.New())

	playedMatch := makeMatch(eveningID, player, opponent1, true)
	unplayedMatch := makeMatch(eveningID, player, opponent2, false)
	unrelatedMatch := makeMatch(eveningID, unrelated, opponent1, false)

	repo := newStubMatchRepo([]domain.Match{playedMatch, unplayedMatch, unrelatedMatch})
	uc := usecase.NewScoreUseCase(repo, nil)

	if err := uc.ReportAbsent(ctx, domain.EveningID(eveningID), player, "test"); err != nil {
		t.Fatalf("ReportAbsent error: %v", err)
	}

	// Played match must remain unchanged.
	got := repo.matches[playedMatch.ID]
	if got.ReportedBy != "" {
		t.Errorf("played match should not be modified, got ReportedBy=%q", got.ReportedBy)
	}

	// Unplayed match for the player must be marked.
	got = repo.matches[unplayedMatch.ID]
	if got.ReportedBy != "test" {
		t.Errorf("unplayed match should have ReportedBy=%q, got %q", "test", got.ReportedBy)
	}

	// Unrelated match (different player) must not be touched.
	got = repo.matches[unrelatedMatch.ID]
	if got.ReportedBy != "" {
		t.Errorf("unrelated match should not be modified, got ReportedBy=%q", got.ReportedBy)
	}
}

// TestReportAbsent_PlayerBAlsoMarked verifies the match is found when the
// absent player is PlayerB rather than PlayerA.
func TestReportAbsent_PlayerBAlsoMarked(t *testing.T) {
	ctx := context.Background()

	eveningID := domain.EveningID(uuid.New())
	playerA := domain.PlayerID(uuid.New())
	playerB := domain.PlayerID(uuid.New())

	m := makeMatch(eveningID, playerA, playerB, false)
	repo := newStubMatchRepo([]domain.Match{m})
	uc := usecase.NewScoreUseCase(repo, nil)

	if err := uc.ReportAbsent(ctx, eveningID, playerB, "reporter"); err != nil {
		t.Fatalf("ReportAbsent error: %v", err)
	}

	got := repo.matches[m.ID]
	if got.ReportedBy != "reporter" {
		t.Errorf("match where absent player is PlayerB should be marked, got ReportedBy=%q", got.ReportedBy)
	}
}

// TestReportAbsent_DoesNotOverwriteExistingReporter verifies that when a match
// was already reported absent by one player, the original reporter is preserved
// if the opponent also calls ReportAbsent afterwards.
func TestReportAbsent_DoesNotOverwriteExistingReporter(t *testing.T) {
	ctx := context.Background()

	eveningID := domain.EveningID(uuid.New())
	playerA := domain.PlayerID(uuid.New())
	playerB := domain.PlayerID(uuid.New())

	m := makeMatch(eveningID, playerA, playerB, false)
	m.ReportedBy = "eerste afmelder" // already reported by playerA
	repo := newStubMatchRepo([]domain.Match{m})
	uc := usecase.NewScoreUseCase(repo, nil)

	// playerB now also reports absent — should not overwrite
	if err := uc.ReportAbsent(ctx, eveningID, playerB, "tweede afmelder"); err != nil {
		t.Fatalf("ReportAbsent error: %v", err)
	}

	got := repo.matches[m.ID]
	if got.ReportedBy != "eerste afmelder" {
		t.Errorf("original reporter should be preserved, got ReportedBy=%q", got.ReportedBy)
	}
}

// TestReportAbsent_NoMatchesOnEvening verifies no error is returned when the
// player has no matches on the given evening.
func TestReportAbsent_NoMatchesOnEvening(t *testing.T) {
	ctx := context.Background()

	eveningID := domain.EveningID(uuid.New())
	player := domain.PlayerID(uuid.New())

	repo := newStubMatchRepo(nil)
	uc := usecase.NewScoreUseCase(repo, nil)

	if err := uc.ReportAbsent(ctx, eveningID, player, "test"); err != nil {
		t.Errorf("expected no error for empty evening, got %v", err)
	}
}
