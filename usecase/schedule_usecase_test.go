package usecase_test

import (
	"context"
	"testing"
	"time"

	"DartScheduler/domain"
	"DartScheduler/usecase"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Minimal stubs for ScheduleUseCase
// ---------------------------------------------------------------------------

type stubScheduleRepo struct {
	sched domain.Schedule
}

func (r *stubScheduleRepo) Save(_ context.Context, _ domain.Schedule) error { return nil }
func (r *stubScheduleRepo) FindLatest(_ context.Context) (domain.Schedule, error) {
	return r.sched, nil
}
func (r *stubScheduleRepo) FindByID(_ context.Context, _ domain.ScheduleID) (domain.Schedule, error) {
	return r.sched, nil
}
func (r *stubScheduleRepo) FindAll(_ context.Context) ([]domain.Schedule, error) {
	return []domain.Schedule{r.sched}, nil
}
func (r *stubScheduleRepo) Delete(_ context.Context, _ domain.ScheduleID) error { return nil }

type stubEveningRepo struct {
	evenings []domain.Evening
}

func (r *stubEveningRepo) Save(_ context.Context, _ domain.Evening, _ domain.ScheduleID) error {
	return nil
}
func (r *stubEveningRepo) FindByID(_ context.Context, _ domain.EveningID) (domain.Evening, error) {
	return domain.Evening{}, domain.ErrNotFound
}
func (r *stubEveningRepo) FindBySchedule(_ context.Context, _ domain.ScheduleID) ([]domain.Evening, error) {
	return r.evenings, nil
}
func (r *stubEveningRepo) Delete(_ context.Context, _ domain.EveningID) error { return nil }
func (r *stubEveningRepo) DeleteBySchedule(_ context.Context, _ domain.ScheduleID) error {
	return nil
}

// hydrateMatchRepo is a match repo stub that tracks which methods were called
// and returns matches keyed by schedule or evening.
type hydrateMatchRepo struct {
	// bySchedule is returned from FindBySchedule.
	bySchedule []domain.Match
	// cancelled is returned from FindCancelledBySchedule.
	cancelled []domain.Match

	findByScheduleCalls   int
	findCancelledCalls    int
}

func (r *hydrateMatchRepo) Save(_ context.Context, _ domain.Match) error              { return nil }
func (r *hydrateMatchRepo) SaveBatch(_ context.Context, _ []domain.Match) error       { return nil }
func (r *hydrateMatchRepo) FindByID(_ context.Context, _ domain.MatchID) (domain.Match, error) {
	return domain.Match{}, domain.ErrNotFound
}
func (r *hydrateMatchRepo) FindByEvening(_ context.Context, _ domain.EveningID) ([]domain.Match, error) {
	return nil, nil
}
func (r *hydrateMatchRepo) FindByPlayer(_ context.Context, _ domain.PlayerID) ([]domain.Match, error) {
	return nil, nil
}
func (r *hydrateMatchRepo) FindByPlayerAndSchedule(_ context.Context, _ domain.PlayerID, _ domain.ScheduleID) ([]domain.Match, error) {
	return nil, nil
}
func (r *hydrateMatchRepo) FindAllPlayed(_ context.Context) ([]domain.Match, error) { return nil, nil }
func (r *hydrateMatchRepo) UpdateResult(_ context.Context, _ domain.Match) error    { return nil }
func (r *hydrateMatchRepo) DeleteByEvening(_ context.Context, _ domain.EveningID) error {
	return nil
}
func (r *hydrateMatchRepo) DeleteBySchedule(_ context.Context, _ domain.ScheduleID) error {
	return nil
}
func (r *hydrateMatchRepo) DeleteByPlayer(_ context.Context, _ domain.PlayerID) error { return nil }

func (r *hydrateMatchRepo) FindBySchedule(_ context.Context, _ domain.ScheduleID) ([]domain.Match, error) {
	r.findByScheduleCalls++
	return r.bySchedule, nil
}
func (r *hydrateMatchRepo) FindCancelledBySchedule(_ context.Context, _ domain.ScheduleID) ([]domain.Match, error) {
	r.findCancelledCalls++
	return r.cancelled, nil
}

// stubPlayerRepo satisfies domain.PlayerRepository with no-ops.
type stubPlayerRepo struct{}

func (r *stubPlayerRepo) Save(_ context.Context, _ domain.Player) error         { return nil }
func (r *stubPlayerRepo) SaveBatch(_ context.Context, _ []domain.Player) error  { return nil }
func (r *stubPlayerRepo) FindByID(_ context.Context, _ domain.PlayerID) (domain.Player, error) {
	return domain.Player{}, domain.ErrNotFound
}
func (r *stubPlayerRepo) FindAll(_ context.Context) ([]domain.Player, error) { return nil, nil }
func (r *stubPlayerRepo) Delete(_ context.Context, _ domain.PlayerID) error   { return nil }
func (r *stubPlayerRepo) DeleteAll(_ context.Context) error                   { return nil }
func (r *stubPlayerRepo) SaveBuddyPreference(_ context.Context, _ domain.BuddyPreference) error {
	return nil
}
func (r *stubPlayerRepo) FindBuddiesForPlayer(_ context.Context, _ domain.PlayerID) ([]domain.PlayerID, error) {
	return nil, nil
}
func (r *stubPlayerRepo) FindAllBuddyPairs(_ context.Context) ([]domain.BuddyPreference, error) {
	return nil, nil
}
func (r *stubPlayerRepo) DeleteBuddiesForPlayer(_ context.Context, _ domain.PlayerID) error {
	return nil
}
func (r *stubPlayerRepo) DeleteAllBuddyPairs(_ context.Context) error { return nil }

// ---------------------------------------------------------------------------
// hydrate tests
// ---------------------------------------------------------------------------

// makeEvening creates a test evening with a given ID and catch-up flag.
func makeEvening(id domain.EveningID, catchUp bool) domain.Evening {
	return domain.Evening{
		ID:               id,
		Number:           1,
		Date:             time.Now(),
		IsCatchUpEvening: catchUp,
	}
}

// makeTestMatch creates a minimal match associated with the given evening.
func makeTestMatch(eveningID domain.EveningID) domain.Match {
	return domain.Match{
		ID:        domain.MatchID(uuid.New()),
		EveningID: eveningID,
		PlayerA:   domain.PlayerID(uuid.New()),
		PlayerB:   domain.PlayerID(uuid.New()),
	}
}

// TestHydrate_MatchesGroupedByEvening verifies that matches returned by
// FindBySchedule are correctly distributed to the right evenings so that
// each evening only sees its own matches.
func TestHydrate_MatchesGroupedByEvening(t *testing.T) {
	ctx := context.Background()

	ev1ID := domain.EveningID(uuid.New())
	ev2ID := domain.EveningID(uuid.New())

	ev1Match1 := makeTestMatch(ev1ID)
	ev1Match2 := makeTestMatch(ev1ID)
	ev2Match1 := makeTestMatch(ev2ID)

	schedID := uuid.New()
	sched := domain.Schedule{ID: schedID}

	matchRepo := &hydrateMatchRepo{
		bySchedule: []domain.Match{ev1Match1, ev1Match2, ev2Match1},
	}
	eveningRepo := &stubEveningRepo{
		evenings: []domain.Evening{
			makeEvening(ev1ID, false),
			makeEvening(ev2ID, false),
		},
	}

	uc := usecase.NewScheduleUseCase(&stubPlayerRepo{}, &stubScheduleRepo{sched: sched}, eveningRepo, matchRepo)

	got, err := uc.GetByID(ctx, schedID)
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}

	if len(got.Evenings) != 2 {
		t.Fatalf("expected 2 evenings, got %d", len(got.Evenings))
	}

	// Find evenings by ID in the result (order is not guaranteed).
	byID := make(map[domain.EveningID]domain.Evening, len(got.Evenings))
	for _, ev := range got.Evenings {
		byID[ev.ID] = ev
	}

	ev1Got, ok := byID[ev1ID]
	if !ok {
		t.Fatal("evening 1 missing from result")
	}
	if len(ev1Got.Matches) != 2 {
		t.Errorf("evening 1: expected 2 matches, got %d", len(ev1Got.Matches))
	}

	ev2Got, ok := byID[ev2ID]
	if !ok {
		t.Fatal("evening 2 missing from result")
	}
	if len(ev2Got.Matches) != 1 {
		t.Errorf("evening 2: expected 1 match, got %d", len(ev2Got.Matches))
	}

	// Verify FindBySchedule was called exactly once (bulk query).
	if matchRepo.findByScheduleCalls != 1 {
		t.Errorf("FindBySchedule should be called once, got %d calls", matchRepo.findByScheduleCalls)
	}
}

// TestHydrate_EveningWithNoMatchesGetsEmptySlice verifies that an evening
// that has no matches in the repository receives a nil/empty matches slice
// rather than matches belonging to another evening.
func TestHydrate_EveningWithNoMatchesGetsEmptySlice(t *testing.T) {
	ctx := context.Background()

	ev1ID := domain.EveningID(uuid.New())
	ev2ID := domain.EveningID(uuid.New()) // this evening has no matches

	ev1Match := makeTestMatch(ev1ID)

	schedID := uuid.New()
	sched := domain.Schedule{ID: schedID}

	matchRepo := &hydrateMatchRepo{
		bySchedule: []domain.Match{ev1Match},
	}
	eveningRepo := &stubEveningRepo{
		evenings: []domain.Evening{
			makeEvening(ev1ID, false),
			makeEvening(ev2ID, false),
		},
	}

	uc := usecase.NewScheduleUseCase(&stubPlayerRepo{}, &stubScheduleRepo{sched: sched}, eveningRepo, matchRepo)

	got, err := uc.GetByID(ctx, schedID)
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}

	byID := make(map[domain.EveningID]domain.Evening, len(got.Evenings))
	for _, ev := range got.Evenings {
		byID[ev.ID] = ev
	}

	ev2Got := byID[ev2ID]
	if len(ev2Got.Matches) != 0 {
		t.Errorf("evening with no matches should have 0 matches, got %d", len(ev2Got.Matches))
	}
}

// TestHydrate_CatchUpEveningGetsCancelledMatches verifies that a catch-up
// evening receives the cancelled matches and not the regular match pool.
func TestHydrate_CatchUpEveningGetsCancelledMatches(t *testing.T) {
	ctx := context.Background()

	regularEveningID := domain.EveningID(uuid.New())
	catchUpEveningID := domain.EveningID(uuid.New())

	regularMatch := makeTestMatch(regularEveningID)
	cancelledMatch := makeTestMatch(regularEveningID) // originally from a regular evening

	schedID := uuid.New()
	sched := domain.Schedule{ID: schedID}

	matchRepo := &hydrateMatchRepo{
		bySchedule: []domain.Match{regularMatch},
		cancelled:  []domain.Match{cancelledMatch},
	}
	eveningRepo := &stubEveningRepo{
		evenings: []domain.Evening{
			makeEvening(regularEveningID, false),
			makeEvening(catchUpEveningID, true),
		},
	}

	uc := usecase.NewScheduleUseCase(&stubPlayerRepo{}, &stubScheduleRepo{sched: sched}, eveningRepo, matchRepo)

	got, err := uc.GetByID(ctx, schedID)
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}

	byID := make(map[domain.EveningID]domain.Evening, len(got.Evenings))
	for _, ev := range got.Evenings {
		byID[ev.ID] = ev
	}

	catchUpGot := byID[catchUpEveningID]
	if len(catchUpGot.Matches) != 1 || catchUpGot.Matches[0].ID != cancelledMatch.ID {
		t.Errorf("catch-up evening: expected cancelled match, got %v", catchUpGot.Matches)
	}

	regularGot := byID[regularEveningID]
	if len(regularGot.Matches) != 1 || regularGot.Matches[0].ID != regularMatch.ID {
		t.Errorf("regular evening: expected regular match, got %v", regularGot.Matches)
	}
}

// TestHydrate_MultipleCatchUpEveningsQueryCancelledOnce verifies that when
// there are multiple catch-up evenings, FindCancelledBySchedule is only called
// once (lazy-load / single-query behaviour).
func TestHydrate_MultipleCatchUpEveningsQueryCancelledOnce(t *testing.T) {
	ctx := context.Background()

	catchUp1ID := domain.EveningID(uuid.New())
	catchUp2ID := domain.EveningID(uuid.New())

	cancelledMatch := makeTestMatch(domain.EveningID(uuid.New()))

	schedID := uuid.New()
	sched := domain.Schedule{ID: schedID}

	matchRepo := &hydrateMatchRepo{
		bySchedule: nil,
		cancelled:  []domain.Match{cancelledMatch},
	}
	eveningRepo := &stubEveningRepo{
		evenings: []domain.Evening{
			makeEvening(catchUp1ID, true),
			makeEvening(catchUp2ID, true),
		},
	}

	uc := usecase.NewScheduleUseCase(&stubPlayerRepo{}, &stubScheduleRepo{sched: sched}, eveningRepo, matchRepo)

	if _, err := uc.GetByID(ctx, schedID); err != nil {
		t.Fatalf("GetByID error: %v", err)
	}

	if matchRepo.findCancelledCalls != 1 {
		t.Errorf("FindCancelledBySchedule should be called exactly once for multiple catch-up evenings, got %d calls", matchRepo.findCancelledCalls)
	}
}
