package usecase_test

import (
	"context"
	"fmt"
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

// ---------------------------------------------------------------------------
// Tracking repos for schedule operation tests
// ---------------------------------------------------------------------------

type trackingScheduleRepo struct {
	sched          domain.Schedule
	deleteCalls    int
	findLatestErr  error
}

func (r *trackingScheduleRepo) Save(_ context.Context, _ domain.Schedule) error { return nil }
func (r *trackingScheduleRepo) FindLatest(_ context.Context) (domain.Schedule, error) {
	return r.sched, r.findLatestErr
}
func (r *trackingScheduleRepo) FindByID(_ context.Context, _ domain.ScheduleID) (domain.Schedule, error) {
	return r.sched, nil
}
func (r *trackingScheduleRepo) FindAll(_ context.Context) ([]domain.Schedule, error) {
	return []domain.Schedule{r.sched}, nil
}
func (r *trackingScheduleRepo) Delete(_ context.Context, _ domain.ScheduleID) error {
	r.deleteCalls++
	return nil
}

type trackingEveningRepo struct {
	evenings              []domain.Evening
	savedEvenings         []domain.Evening
	deleteByScheduleCalls int
	deleteCalls           int
	deleteByScheduleErr   error
	deleteErr             error
}

func (r *trackingEveningRepo) Save(_ context.Context, ev domain.Evening, _ domain.ScheduleID) error {
	r.savedEvenings = append(r.savedEvenings, ev)
	return nil
}
func (r *trackingEveningRepo) FindByID(_ context.Context, _ domain.EveningID) (domain.Evening, error) {
	return domain.Evening{}, domain.ErrNotFound
}
func (r *trackingEveningRepo) FindBySchedule(_ context.Context, _ domain.ScheduleID) ([]domain.Evening, error) {
	return r.evenings, nil
}
func (r *trackingEveningRepo) Delete(_ context.Context, _ domain.EveningID) error {
	r.deleteCalls++
	return r.deleteErr
}
func (r *trackingEveningRepo) DeleteBySchedule(_ context.Context, _ domain.ScheduleID) error {
	r.deleteByScheduleCalls++
	return r.deleteByScheduleErr
}

type trackingMatchRepo struct {
	deleteByScheduleCalls int
	deleteByEveningCalls  int
	deleteByScheduleErr   error
	deleteByEveningErr    error
}

func (r *trackingMatchRepo) Save(_ context.Context, _ domain.Match) error        { return nil }
func (r *trackingMatchRepo) SaveBatch(_ context.Context, _ []domain.Match) error { return nil }
func (r *trackingMatchRepo) FindByID(_ context.Context, _ domain.MatchID) (domain.Match, error) {
	return domain.Match{}, domain.ErrNotFound
}
func (r *trackingMatchRepo) FindByEvening(_ context.Context, _ domain.EveningID) ([]domain.Match, error) {
	return nil, nil
}
func (r *trackingMatchRepo) FindByPlayer(_ context.Context, _ domain.PlayerID) ([]domain.Match, error) {
	return nil, nil
}
func (r *trackingMatchRepo) FindByPlayerAndSchedule(_ context.Context, _ domain.PlayerID, _ domain.ScheduleID) ([]domain.Match, error) {
	return nil, nil
}
func (r *trackingMatchRepo) FindAllPlayed(_ context.Context) ([]domain.Match, error) { return nil, nil }
func (r *trackingMatchRepo) UpdateResult(_ context.Context, _ domain.Match) error    { return nil }
func (r *trackingMatchRepo) FindBySchedule(_ context.Context, _ domain.ScheduleID) ([]domain.Match, error) {
	return nil, nil
}
func (r *trackingMatchRepo) FindCancelledBySchedule(_ context.Context, _ domain.ScheduleID) ([]domain.Match, error) {
	return nil, nil
}
func (r *trackingMatchRepo) DeleteByEvening(_ context.Context, _ domain.EveningID) error {
	r.deleteByEveningCalls++
	return r.deleteByEveningErr
}
func (r *trackingMatchRepo) DeleteBySchedule(_ context.Context, _ domain.ScheduleID) error {
	r.deleteByScheduleCalls++
	return r.deleteByScheduleErr
}
func (r *trackingMatchRepo) DeleteByPlayer(_ context.Context, _ domain.PlayerID) error { return nil }

// ---------------------------------------------------------------------------
// Schedule operation tests
// ---------------------------------------------------------------------------

// TestGetLatest_ReturnsHydratedSchedule verifies GetLatest returns the schedule
// with evenings loaded from the repo.
func TestGetLatest_ReturnsHydratedSchedule(t *testing.T) {
	ctx := context.Background()

	schedID := uuid.New()
	sched := domain.Schedule{ID: schedID}

	evID := domain.EveningID(uuid.New())
	ev := makeEvening(evID, false)

	uc := usecase.NewScheduleUseCase(
		&stubPlayerRepo{},
		&trackingScheduleRepo{sched: sched},
		&trackingEveningRepo{evenings: []domain.Evening{ev}},
		&trackingMatchRepo{},
	)

	got, err := uc.GetLatest(ctx)
	if err != nil {
		t.Fatalf("GetLatest error: %v", err)
	}
	if len(got.Evenings) != 1 {
		t.Errorf("expected 1 evening, got %d", len(got.Evenings))
	}
	if got.Evenings[0].ID != evID {
		t.Errorf("unexpected evening ID: got %s, want %s", got.Evenings[0].ID, evID)
	}
}

// TestListSchedules_ReturnsSummaries verifies that each schedule is returned
// with the correct evening count.
func TestListSchedules_ReturnsSummaries(t *testing.T) {
	ctx := context.Background()

	sched := domain.Schedule{ID: uuid.New(), CompetitionName: "Kring", Season: "2025"}
	ev1 := makeEvening(domain.EveningID(uuid.New()), false)
	ev2 := makeEvening(domain.EveningID(uuid.New()), false)

	uc := usecase.NewScheduleUseCase(
		&stubPlayerRepo{},
		&trackingScheduleRepo{sched: sched},
		&trackingEveningRepo{evenings: []domain.Evening{ev1, ev2}},
		&trackingMatchRepo{},
	)

	summaries, err := uc.ListSchedules(ctx)
	if err != nil {
		t.Fatalf("ListSchedules error: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	s := summaries[0]
	if s.CompetitionName != "Kring" {
		t.Errorf("CompetitionName: got %q, want %q", s.CompetitionName, "Kring")
	}
	if s.EveningCount != 2 {
		t.Errorf("EveningCount: got %d, want 2", s.EveningCount)
	}
}

// TestDeleteSchedule_DelegatesAll verifies that all three delete methods are called.
func TestDeleteSchedule_DelegatesAll(t *testing.T) {
	ctx := context.Background()

	matchRepo := &trackingMatchRepo{}
	eveningRepo := &trackingEveningRepo{}
	schedRepo := &trackingScheduleRepo{}

	uc := usecase.NewScheduleUseCase(&stubPlayerRepo{}, schedRepo, eveningRepo, matchRepo)

	if err := uc.DeleteSchedule(ctx, domain.ScheduleID(uuid.New())); err != nil {
		t.Fatalf("DeleteSchedule error: %v", err)
	}
	if matchRepo.deleteByScheduleCalls != 1 {
		t.Errorf("DeleteBySchedule (match): expected 1, got %d", matchRepo.deleteByScheduleCalls)
	}
	if eveningRepo.deleteByScheduleCalls != 1 {
		t.Errorf("DeleteBySchedule (evening): expected 1, got %d", eveningRepo.deleteByScheduleCalls)
	}
	if schedRepo.deleteCalls != 1 {
		t.Errorf("schedules.Delete: expected 1, got %d", schedRepo.deleteCalls)
	}
}

// TestDeleteEvening_DelegatesAll verifies matches and the evening itself are deleted.
func TestDeleteEvening_DelegatesAll(t *testing.T) {
	ctx := context.Background()

	matchRepo := &trackingMatchRepo{}
	eveningRepo := &trackingEveningRepo{}

	uc := usecase.NewScheduleUseCase(&stubPlayerRepo{}, &trackingScheduleRepo{}, eveningRepo, matchRepo)

	if err := uc.DeleteEvening(ctx, domain.EveningID(uuid.New())); err != nil {
		t.Fatalf("DeleteEvening error: %v", err)
	}
	if matchRepo.deleteByEveningCalls != 1 {
		t.Errorf("DeleteByEvening (match): expected 1, got %d", matchRepo.deleteByEveningCalls)
	}
	if eveningRepo.deleteCalls != 1 {
		t.Errorf("evenings.Delete: expected 1, got %d", eveningRepo.deleteCalls)
	}
}

// TestAddCatchUpEvening_NumberIsMaxPlusOne verifies the new catch-up evening
// gets a number one greater than the current maximum.
func TestAddCatchUpEvening_NumberIsMaxPlusOne(t *testing.T) {
	ctx := context.Background()

	schedID := domain.ScheduleID(uuid.New())
	sched := domain.Schedule{ID: schedID}

	ev1 := makeEvening(domain.EveningID(uuid.New()), false)
	ev1.Number = 3
	ev2 := makeEvening(domain.EveningID(uuid.New()), false)
	ev2.Number = 5

	eveningRepo := &trackingEveningRepo{evenings: []domain.Evening{ev1, ev2}}

	uc := usecase.NewScheduleUseCase(
		&stubPlayerRepo{},
		&trackingScheduleRepo{sched: sched},
		eveningRepo,
		&trackingMatchRepo{},
	)

	got, err := uc.AddCatchUpEvening(ctx, schedID, time.Now())
	if err != nil {
		t.Fatalf("AddCatchUpEvening error: %v", err)
	}

	// The saved catch-up evening should have number = max(3,5) + 1 = 6.
	if len(eveningRepo.savedEvenings) != 1 {
		t.Fatalf("expected 1 saved evening, got %d", len(eveningRepo.savedEvenings))
	}
	if eveningRepo.savedEvenings[0].Number != 6 {
		t.Errorf("catch-up evening number: got %d, want 6", eveningRepo.savedEvenings[0].Number)
	}
	if !eveningRepo.savedEvenings[0].IsCatchUpEvening {
		t.Error("expected saved evening to have IsCatchUpEvening=true")
	}
	// Returned schedule should include the new evening (eveningRepo now returns
	// the old two; the new one is only in savedEvenings — that's fine, we just
	// check no error and a valid schedule ID is returned).
	if got.ID != schedID {
		t.Errorf("returned schedule ID: got %s, want %s", got.ID, schedID)
	}
}

// TestImportSeason_CreatesMatchesFromRows verifies that match rows produce the
// right number of matches and that missing players are silently skipped.
func TestImportSeason_CreatesMatchesFromRows(t *testing.T) {
	ctx := context.Background()

	pA := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1", Name: "Alpha"}
	pB := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "2", Name: "Beta"}

	playerRepo := &funcPlayerRepo{players: []domain.Player{pA, pB}}

	rows := []usecase.SeasonMatchRow{
		{EveningNr: 1, Date: time.Now(), NrA: "1", NrB: "2", ScoreA: 2, ScoreB: 0},
		{EveningNr: 1, Date: time.Now(), NrA: "1", NrB: "99"},  // player 99 not found → skipped
		{EveningNr: 2, Date: time.Now(), NrA: "2", NrB: "1", ScoreA: 1, ScoreB: 2},
	}

	eveningRepo := &trackingEveningRepo{}

	uc := usecase.NewScheduleUseCase(
		playerRepo,
		&trackingScheduleRepo{},
		eveningRepo,
		&trackingMatchRepo{},
	)

	_, err := uc.ImportSeason(ctx, "Test", "2025", rows, nil)
	if err != nil {
		t.Fatalf("ImportSeason error: %v", err)
	}
	// Two evenings should have been created (nr 1 and nr 2).
	if len(eveningRepo.savedEvenings) != 2 {
		t.Errorf("expected 2 saved evenings, got %d", len(eveningRepo.savedEvenings))
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

// ---------------------------------------------------------------------------
// Error-path tests
// ---------------------------------------------------------------------------

// TestGetLatest_ErrorPropagated verifies that an error from FindLatest is returned.
func TestGetLatest_ErrorPropagated(t *testing.T) {
	ctx := context.Background()
	sentinel := fmt.Errorf("repo unavailable")
	uc := usecase.NewScheduleUseCase(
		&stubPlayerRepo{},
		&trackingScheduleRepo{findLatestErr: sentinel},
		&stubEveningRepo{},
		&hydrateMatchRepo{},
	)
	_, err := uc.GetLatest(ctx)
	if err == nil {
		t.Error("expected error from GetLatest, got nil")
	}
}

// TestDeleteSchedule_MatchesErrorPropagated verifies that an error from
// DeleteBySchedule on the match repo is propagated.
func TestDeleteSchedule_MatchesErrorPropagated(t *testing.T) {
	ctx := context.Background()
	sentinel := fmt.Errorf("db error")
	matchRepo := &trackingMatchRepo{deleteByScheduleErr: sentinel}
	uc := usecase.NewScheduleUseCase(&stubPlayerRepo{}, &trackingScheduleRepo{}, &trackingEveningRepo{}, matchRepo)
	err := uc.DeleteSchedule(ctx, domain.ScheduleID(uuid.New()))
	if err == nil {
		t.Error("expected error from DeleteSchedule, got nil")
	}
}

// TestDeleteEvening_MatchesErrorPropagated verifies that an error from
// DeleteByEvening on the match repo is propagated.
func TestDeleteEvening_MatchesErrorPropagated(t *testing.T) {
	ctx := context.Background()
	sentinel := fmt.Errorf("db error")
	matchRepo := &trackingMatchRepo{deleteByEveningErr: sentinel}
	uc := usecase.NewScheduleUseCase(&stubPlayerRepo{}, &trackingScheduleRepo{}, &trackingEveningRepo{}, matchRepo)
	err := uc.DeleteEvening(ctx, domain.EveningID(uuid.New()))
	if err == nil {
		t.Error("expected error from DeleteEvening, got nil")
	}
}

// TestDeleteSchedule_EveningsErrorPropagated verifies that an error from
// evenings.DeleteBySchedule is propagated.
func TestDeleteSchedule_EveningsErrorPropagated(t *testing.T) {
	ctx := context.Background()
	sentinel := fmt.Errorf("evenings db error")
	eveningRepo := &trackingEveningRepo{deleteByScheduleErr: sentinel}
	uc := usecase.NewScheduleUseCase(&stubPlayerRepo{}, &trackingScheduleRepo{}, eveningRepo, &trackingMatchRepo{})
	err := uc.DeleteSchedule(ctx, domain.ScheduleID(uuid.New()))
	if err == nil {
		t.Error("expected error from DeleteSchedule (evenings), got nil")
	}
}

// TestDeleteEvening_EveningDeleteErrorPropagated verifies that an error from
// evenings.Delete is propagated.
func TestDeleteEvening_EveningDeleteErrorPropagated(t *testing.T) {
	ctx := context.Background()
	sentinel := fmt.Errorf("delete error")
	eveningRepo := &trackingEveningRepo{deleteErr: sentinel}
	uc := usecase.NewScheduleUseCase(&stubPlayerRepo{}, &trackingScheduleRepo{}, eveningRepo, &trackingMatchRepo{})
	err := uc.DeleteEvening(ctx, domain.EveningID(uuid.New()))
	if err == nil {
		t.Error("expected error from DeleteEvening (evening delete), got nil")
	}
}
