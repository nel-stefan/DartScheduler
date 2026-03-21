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
// infoMatchRepo: extends trackingMatchRepo with per-evening match data
// ---------------------------------------------------------------------------

type infoMatchRepo struct {
	*trackingMatchRepo
	byEvening map[domain.EveningID][]domain.Match
}

func newInfoMatchRepo() *infoMatchRepo {
	return &infoMatchRepo{
		trackingMatchRepo: &trackingMatchRepo{},
		byEvening:         make(map[domain.EveningID][]domain.Match),
	}
}

func (r *infoMatchRepo) FindByEvening(_ context.Context, id domain.EveningID) ([]domain.Match, error) {
	return r.byEvening[id], nil
}

// ---------------------------------------------------------------------------
// stubSeasonStatRepo
// ---------------------------------------------------------------------------

type stubSeasonStatRepo struct {
	stats []domain.SeasonPlayerStat
}

func (r *stubSeasonStatRepo) FindBySchedule(_ context.Context, _ domain.ScheduleID) ([]domain.SeasonPlayerStat, error) {
	return r.stats, nil
}
func (r *stubSeasonStatRepo) Upsert(_ context.Context, _ domain.SeasonPlayerStat) error { return nil }

// ---------------------------------------------------------------------------
// GetInfo tests
// ---------------------------------------------------------------------------

// TestGetInfo_ReturnsMatrix verifies that player×evening match counts are
// correctly built from FindByEvening results.
func TestGetInfo_ReturnsMatrix(t *testing.T) {
	ctx := context.Background()

	schedID := domain.ScheduleID(uuid.New())
	sched := domain.Schedule{ID: schedID}

	pA := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1", Name: "Doe, John"}
	pB := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "2", Name: "Smith, Jane"}

	evID := domain.EveningID(uuid.New())
	ev := domain.Evening{ID: evID, Number: 1, Date: time.Now(), IsCatchUpEvening: false}

	m := domain.Match{
		ID:      domain.MatchID(uuid.New()),
		PlayerA: pA.ID,
		PlayerB: pB.ID,
	}

	playerRepo := &funcPlayerRepo{players: []domain.Player{pA, pB}}
	eveningRepo := &trackingEveningRepo{evenings: []domain.Evening{ev}}
	matchRepo := newInfoMatchRepo()
	matchRepo.byEvening[evID] = []domain.Match{m}

	uc := usecase.NewScheduleUseCase(playerRepo, &trackingScheduleRepo{sched: sched}, eveningRepo, matchRepo)

	result, err := uc.GetInfo(ctx, schedID)
	if err != nil {
		t.Fatalf("GetInfo error: %v", err)
	}

	if len(result.Players) != 2 {
		t.Errorf("Players: got %d, want 2", len(result.Players))
	}
	if len(result.Evenings) != 1 {
		t.Errorf("Evenings: got %d, want 1", len(result.Evenings))
	}
	if len(result.Matrix) != 2 {
		t.Errorf("Matrix: got %d cells, want 2 (one per player)", len(result.Matrix))
	}
	for _, cell := range result.Matrix {
		if cell.Count != 1 {
			t.Errorf("Matrix cell count: got %d, want 1", cell.Count)
		}
	}
}

// TestGetInfo_CatchUpEveningExcludedFromMatrix verifies that catch-up evenings
// do not appear in the matrix or the evenings list.
func TestGetInfo_CatchUpEveningExcludedFromMatrix(t *testing.T) {
	ctx := context.Background()

	schedID := domain.ScheduleID(uuid.New())
	sched := domain.Schedule{ID: schedID}

	pA := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1", Name: "Alpha"}

	regularEvID := domain.EveningID(uuid.New())
	catchUpEvID := domain.EveningID(uuid.New())
	regularEv := domain.Evening{ID: regularEvID, Number: 1, Date: time.Now(), IsCatchUpEvening: false}
	catchUpEv := domain.Evening{ID: catchUpEvID, Number: 2, Date: time.Now(), IsCatchUpEvening: true}

	matchRepo := newInfoMatchRepo()
	// No matches in either evening.

	playerRepo := &funcPlayerRepo{players: []domain.Player{pA}}
	eveningRepo := &trackingEveningRepo{evenings: []domain.Evening{regularEv, catchUpEv}}

	uc := usecase.NewScheduleUseCase(playerRepo, &trackingScheduleRepo{sched: sched}, eveningRepo, matchRepo)

	result, err := uc.GetInfo(ctx, schedID)
	if err != nil {
		t.Fatalf("GetInfo error: %v", err)
	}

	// Only the regular evening should appear.
	if len(result.Evenings) != 1 {
		t.Errorf("Evenings: got %d, want 1 (catch-up excluded)", len(result.Evenings))
	}
	if result.Evenings[0].Number != 1 {
		t.Errorf("Evenings[0].Number: got %d, want 1", result.Evenings[0].Number)
	}
	// Matrix should be empty (no matches).
	if len(result.Matrix) != 0 {
		t.Errorf("Matrix: got %d cells, want 0", len(result.Matrix))
	}
}

// TestGetInfo_BuddyPairsIncludeSharedEvenings verifies that buddy pairs list
// the evenings they share.
func TestGetInfo_BuddyPairsIncludeSharedEvenings(t *testing.T) {
	ctx := context.Background()

	schedID := domain.ScheduleID(uuid.New())
	sched := domain.Schedule{ID: schedID}

	pA := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1", Name: "Doe, John"}
	pB := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "2", Name: "Smith, Jane"}

	ev1ID := domain.EveningID(uuid.New())
	ev2ID := domain.EveningID(uuid.New())
	ev1 := domain.Evening{ID: ev1ID, Number: 1, Date: time.Now(), IsCatchUpEvening: false}
	ev2 := domain.Evening{ID: ev2ID, Number: 2, Date: time.Now(), IsCatchUpEvening: false}

	// Match in ev1: pA vs pB — they share evening 1.
	m1 := domain.Match{ID: domain.MatchID(uuid.New()), PlayerA: pA.ID, PlayerB: pB.ID}
	// Match in ev2: only pA plays — they don't share evening 2.
	pC := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "3", Name: "Other"}
	m2 := domain.Match{ID: domain.MatchID(uuid.New()), PlayerA: pA.ID, PlayerB: pC.ID}

	playerRepo := &funcPlayerRepo{
		players: []domain.Player{pA, pB, pC},
		buddies: []domain.BuddyPreference{
			{PlayerID: pA.ID, BuddyID: pB.ID},
			{PlayerID: pB.ID, BuddyID: pA.ID},
		},
	}
	eveningRepo := &trackingEveningRepo{evenings: []domain.Evening{ev1, ev2}}
	matchRepo := newInfoMatchRepo()
	matchRepo.byEvening[ev1ID] = []domain.Match{m1}
	matchRepo.byEvening[ev2ID] = []domain.Match{m2}

	uc := usecase.NewScheduleUseCase(playerRepo, &trackingScheduleRepo{sched: sched}, eveningRepo, matchRepo)

	result, err := uc.GetInfo(ctx, schedID)
	if err != nil {
		t.Fatalf("GetInfo error: %v", err)
	}

	if len(result.BuddyPairs) != 1 {
		t.Fatalf("BuddyPairs: got %d, want 1", len(result.BuddyPairs))
	}
	bp := result.BuddyPairs[0]
	if len(bp.EveningNrs) != 1 || bp.EveningNrs[0] != 1 {
		t.Errorf("BuddyPair shared evenings: got %v, want [1]", bp.EveningNrs)
	}
}

// TestGetInfo_NoPlayers verifies an empty but valid result for a schedule with no players.
func TestGetInfo_NoPlayers(t *testing.T) {
	ctx := context.Background()

	schedID := domain.ScheduleID(uuid.New())
	sched := domain.Schedule{ID: schedID}

	uc := usecase.NewScheduleUseCase(
		&funcPlayerRepo{},
		&trackingScheduleRepo{sched: sched},
		&trackingEveningRepo{},
		newInfoMatchRepo(),
	)

	result, err := uc.GetInfo(ctx, schedID)
	if err != nil {
		t.Fatalf("GetInfo error: %v", err)
	}
	if len(result.Players) != 0 || len(result.Evenings) != 0 || len(result.Matrix) != 0 {
		t.Errorf("expected empty result, got players=%d evenings=%d matrix=%d",
			len(result.Players), len(result.Evenings), len(result.Matrix))
	}
}

// ---------------------------------------------------------------------------
// GetStats with seasonStats overlay
// ---------------------------------------------------------------------------

// TestGetStats_SeasonStatsOverlay verifies that 180s and HighestFinish values
// are taken from the SeasonPlayerStatRepository when available.
func TestGetStats_SeasonStatsOverlay(t *testing.T) {
	ctx := context.Background()

	pA := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1"}

	schedID := domain.ScheduleID(uuid.New())
	seasonStat := domain.SeasonPlayerStat{
		ScheduleID:    schedID,
		PlayerID:      pA.ID,
		OneEighties:   5,
		HighestFinish: 140,
	}

	repo := newMatchesByPlayerRepo()
	statRepo := &stubSeasonStatRepo{stats: []domain.SeasonPlayerStat{seasonStat}}

	uc := usecase.NewScoreUseCase(repo, nil, statRepo)
	stats, err := uc.GetStats(ctx, []domain.Player{pA}, &schedID)
	if err != nil {
		t.Fatalf("GetStats error: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}
	s := stats[0]
	if s.OneEighties != 5 {
		t.Errorf("OneEighties: got %d, want 5", s.OneEighties)
	}
	if s.HighestFinish != 140 {
		t.Errorf("HighestFinish: got %d, want 140", s.HighestFinish)
	}
}

// ---------------------------------------------------------------------------
// ImportSeason with catch-up evenings and leg-winner derivation
// ---------------------------------------------------------------------------

// TestImportSeason_WithCatchUpEvenings verifies that catch-up evenings are
// stored alongside regular match evenings.
func TestImportSeason_WithCatchUpEvenings(t *testing.T) {
	ctx := context.Background()

	pA := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1"}
	pB := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "2"}
	playerRepo := &funcPlayerRepo{players: []domain.Player{pA, pB}}

	rows := []usecase.SeasonMatchRow{
		{EveningNr: 1, Date: time.Now(), NrA: "1", NrB: "2", ScoreA: 2, ScoreB: 0},
	}
	catchUps := []usecase.CatchUpEvening{
		{EveningNr: 2, Date: time.Now()},
	}

	eveningRepo := &trackingEveningRepo{}
	uc := usecase.NewScheduleUseCase(
		playerRepo,
		&trackingScheduleRepo{},
		eveningRepo,
		&trackingMatchRepo{},
	)

	_, err := uc.ImportSeason(ctx, "Test", "2025", rows, catchUps)
	if err != nil {
		t.Fatalf("ImportSeason error: %v", err)
	}

	// Two evenings: one from rows, one catch-up.
	if len(eveningRepo.savedEvenings) != 2 {
		t.Errorf("expected 2 saved evenings (1 regular + 1 catch-up), got %d", len(eveningRepo.savedEvenings))
	}
	// Last saved evening should be the catch-up.
	last := eveningRepo.savedEvenings[len(eveningRepo.savedEvenings)-1]
	if !last.IsCatchUpEvening {
		t.Error("expected last saved evening to be a catch-up evening")
	}
}

// TestImportSeason_LegWinnerDerivation verifies that when ScoreA/ScoreB are
// zero but leg winners are set, the score is derived from the leg results.
func TestImportSeason_LegWinnerDerivation(t *testing.T) {
	ctx := context.Background()

	pA := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1", Name: "Alpha"}
	pB := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "2", Name: "Beta"}
	playerRepo := &funcPlayerRepo{players: []domain.Player{pA, pB}}

	// ScoreA/ScoreB == 0 but leg winners are set → derive from legs.
	rows := []usecase.SeasonMatchRow{
		{
			EveningNr:  1,
			Date:       time.Now(),
			NrA:        "1",
			NrB:        "2",
			Leg1Winner: "1",   // pA wins leg 1
			Leg2Winner: "1",   // pA wins leg 2
		},
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
	if len(eveningRepo.savedEvenings) != 1 {
		t.Errorf("expected 1 saved evening, got %d", len(eveningRepo.savedEvenings))
	}
}
