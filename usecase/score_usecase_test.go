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
	uc := usecase.NewScoreUseCase(repo, nil, nil)

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
	uc := usecase.NewScoreUseCase(repo, nil, nil)

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
	uc := usecase.NewScoreUseCase(repo, nil, nil)

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
	uc := usecase.NewScoreUseCase(repo, nil, nil)

	if err := uc.ReportAbsent(ctx, eveningID, player, "test"); err != nil {
		t.Errorf("expected no error for empty evening, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Submit tests
// ---------------------------------------------------------------------------

// TestSubmit_TwoLegWin verifies that a 2-0 win is recorded correctly.
func TestSubmit_TwoLegWin(t *testing.T) {
	ctx := context.Background()

	pA := domain.PlayerID(uuid.New())
	pB := domain.PlayerID(uuid.New())
	m := domain.Match{
		ID:      domain.MatchID(uuid.New()),
		PlayerA: pA,
		PlayerB: pB,
	}
	repo := newStubMatchRepo([]domain.Match{m})
	uc := usecase.NewScoreUseCase(repo, nil, nil)

	err := uc.Submit(ctx, usecase.SubmitScoreInput{
		MatchID:    m.ID,
		Leg1Winner: pA.String(),
		Leg1Turns:  15,
		Leg2Winner: pA.String(),
		Leg2Turns:  18,
	})
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}

	got := repo.matches[m.ID]
	if got.ScoreA == nil || *got.ScoreA != 2 {
		t.Errorf("ScoreA: got %v, want 2", got.ScoreA)
	}
	if got.ScoreB == nil || *got.ScoreB != 0 {
		t.Errorf("ScoreB: got %v, want 0", got.ScoreB)
	}
	if !got.Played {
		t.Error("Played should be true")
	}
}

// TestSubmit_ThreeLegSplit verifies a 2-1 result with three legs.
func TestSubmit_ThreeLegSplit(t *testing.T) {
	ctx := context.Background()

	pA := domain.PlayerID(uuid.New())
	pB := domain.PlayerID(uuid.New())
	m := domain.Match{
		ID:      domain.MatchID(uuid.New()),
		PlayerA: pA,
		PlayerB: pB,
	}
	repo := newStubMatchRepo([]domain.Match{m})
	uc := usecase.NewScoreUseCase(repo, nil, nil)

	err := uc.Submit(ctx, usecase.SubmitScoreInput{
		MatchID:    m.ID,
		Leg1Winner: pA.String(),
		Leg1Turns:  14,
		Leg2Winner: pB.String(),
		Leg2Turns:  16,
		Leg3Winner: pA.String(),
		Leg3Turns:  20,
	})
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}

	got := repo.matches[m.ID]
	if *got.ScoreA != 2 {
		t.Errorf("ScoreA: got %d, want 2", *got.ScoreA)
	}
	if *got.ScoreB != 1 {
		t.Errorf("ScoreB: got %d, want 1", *got.ScoreB)
	}
}

// TestSubmit_AdminFieldsPersisted verifies secretary, counter and 180s fields are saved.
func TestSubmit_AdminFieldsPersisted(t *testing.T) {
	ctx := context.Background()

	pA := domain.PlayerID(uuid.New())
	pB := domain.PlayerID(uuid.New())
	m := domain.Match{
		ID:      domain.MatchID(uuid.New()),
		PlayerA: pA,
		PlayerB: pB,
	}
	repo := newStubMatchRepo([]domain.Match{m})
	uc := usecase.NewScoreUseCase(repo, nil, nil)

	err := uc.Submit(ctx, usecase.SubmitScoreInput{
		MatchID:              m.ID,
		Leg1Winner:           pA.String(),
		Leg1Turns:            17,
		Leg2Winner:           pA.String(),
		Leg2Turns:            19,
		SecretaryNr:          "5",
		CounterNr:            "7",
		PlayerA180s:          2,
		PlayerB180s:          1,
		PlayerAHighestFinish: 120,
		PlayerBHighestFinish: 60,
		ReportedBy:           "5 Jan",
	})
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}

	got := repo.matches[m.ID]
	if got.SecretaryNr != "5" {
		t.Errorf("SecretaryNr: got %q, want %q", got.SecretaryNr, "5")
	}
	if got.CounterNr != "7" {
		t.Errorf("CounterNr: got %q, want %q", got.CounterNr, "7")
	}
	if got.PlayerA180s != 2 {
		t.Errorf("PlayerA180s: got %d, want 2", got.PlayerA180s)
	}
	if got.PlayerAHighestFinish != 120 {
		t.Errorf("PlayerAHighestFinish: got %d, want 120", got.PlayerAHighestFinish)
	}
	if got.ReportedBy != "5 Jan" {
		t.Errorf("ReportedBy: got %q, want %q", got.ReportedBy, "5 Jan")
	}
}

// ---------------------------------------------------------------------------
// matchesByPlayerRepo: match repo that stores matches per player for stats tests
// ---------------------------------------------------------------------------

type matchesByPlayerRepo struct {
	data map[domain.PlayerID][]domain.Match
}

func newMatchesByPlayerRepo() *matchesByPlayerRepo {
	return &matchesByPlayerRepo{data: make(map[domain.PlayerID][]domain.Match)}
}

func (r *matchesByPlayerRepo) addForPlayer(pid domain.PlayerID, m domain.Match) {
	r.data[pid] = append(r.data[pid], m)
}

func (r *matchesByPlayerRepo) FindByPlayer(_ context.Context, pid domain.PlayerID) ([]domain.Match, error) {
	return r.data[pid], nil
}
func (r *matchesByPlayerRepo) FindByPlayerAndSchedule(_ context.Context, pid domain.PlayerID, _ domain.ScheduleID) ([]domain.Match, error) {
	return r.data[pid], nil
}
func (r *matchesByPlayerRepo) FindAllPlayed(_ context.Context) ([]domain.Match, error) {
	seen := make(map[domain.MatchID]struct{})
	var all []domain.Match
	for _, ms := range r.data {
		for _, m := range ms {
			if _, ok := seen[m.ID]; !ok {
				seen[m.ID] = struct{}{}
				all = append(all, m)
			}
		}
	}
	return all, nil
}
func (r *matchesByPlayerRepo) Save(_ context.Context, _ domain.Match) error        { return nil }
func (r *matchesByPlayerRepo) SaveBatch(_ context.Context, _ []domain.Match) error { return nil }
func (r *matchesByPlayerRepo) FindByID(_ context.Context, _ domain.MatchID) (domain.Match, error) {
	return domain.Match{}, domain.ErrNotFound
}
func (r *matchesByPlayerRepo) FindByEvening(_ context.Context, _ domain.EveningID) ([]domain.Match, error) {
	return nil, nil
}
func (r *matchesByPlayerRepo) UpdateResult(_ context.Context, _ domain.Match) error { return nil }
func (r *matchesByPlayerRepo) FindBySchedule(_ context.Context, _ domain.ScheduleID) ([]domain.Match, error) {
	return nil, nil
}
func (r *matchesByPlayerRepo) FindCancelledBySchedule(_ context.Context, _ domain.ScheduleID) ([]domain.Match, error) {
	return nil, nil
}
func (r *matchesByPlayerRepo) DeleteByEvening(_ context.Context, _ domain.EveningID) error {
	return nil
}
func (r *matchesByPlayerRepo) DeleteBySchedule(_ context.Context, _ domain.ScheduleID) error {
	return nil
}
func (r *matchesByPlayerRepo) DeleteByPlayer(_ context.Context, _ domain.PlayerID) error { return nil }

// ---------------------------------------------------------------------------
// GetStats tests
// ---------------------------------------------------------------------------

// TestGetStats_WinAndLoss verifies wins, losses, and point totals are counted correctly.
func TestGetStats_WinAndLoss(t *testing.T) {
	ctx := context.Background()

	pA := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1", Name: "A"}
	pB := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "2", Name: "B"}

	scoreA, scoreB := 2, 0
	m := domain.Match{
		ID:      domain.MatchID(uuid.New()),
		PlayerA: pA.ID,
		PlayerB: pB.ID,
		Played:  true,
		ScoreA:  &scoreA,
		ScoreB:  &scoreB,
	}

	repo := newMatchesByPlayerRepo()
	repo.addForPlayer(pA.ID, m)
	repo.addForPlayer(pB.ID, m)

	uc := usecase.NewScoreUseCase(repo, nil, nil)
	stats, err := uc.GetStats(ctx, []domain.Player{pA, pB}, nil)
	if err != nil {
		t.Fatalf("GetStats error: %v", err)
	}

	byID := make(map[domain.PlayerID]usecase.PlayerStats)
	for _, s := range stats {
		byID[s.Player.ID] = s
	}

	aStats := byID[pA.ID]
	if aStats.Wins != 1 || aStats.Losses != 0 || aStats.Played != 1 {
		t.Errorf("player A: wins=%d losses=%d played=%d, want 1/0/1", aStats.Wins, aStats.Losses, aStats.Played)
	}
	if aStats.PointsFor != 2 || aStats.PointsAgainst != 0 {
		t.Errorf("player A points: for=%d against=%d, want 2/0", aStats.PointsFor, aStats.PointsAgainst)
	}

	bStats := byID[pB.ID]
	if bStats.Wins != 0 || bStats.Losses != 1 {
		t.Errorf("player B: wins=%d losses=%d, want 0/1", bStats.Wins, bStats.Losses)
	}
}

// TestGetStats_DrawCounted verifies that 1–1 results increment the Draws counter.
func TestGetStats_DrawCounted(t *testing.T) {
	ctx := context.Background()

	pA := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1"}
	pB := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "2"}

	s1, s2 := 1, 1
	m := domain.Match{
		ID:      domain.MatchID(uuid.New()),
		PlayerA: pA.ID,
		PlayerB: pB.ID,
		Played:  true,
		ScoreA:  &s1,
		ScoreB:  &s2,
	}

	repo := newMatchesByPlayerRepo()
	repo.addForPlayer(pA.ID, m)
	repo.addForPlayer(pB.ID, m)

	uc := usecase.NewScoreUseCase(repo, nil, nil)
	stats, err := uc.GetStats(ctx, []domain.Player{pA, pB}, nil)
	if err != nil {
		t.Fatalf("GetStats error: %v", err)
	}
	for _, s := range stats {
		if s.Draws != 1 {
			t.Errorf("player %s: draws=%d, want 1", s.Player.Nr, s.Draws)
		}
	}
}

// TestGetStats_TurnStats verifies MinTurns and AvgTurns are computed from won legs.
func TestGetStats_TurnStats(t *testing.T) {
	ctx := context.Background()

	pA := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1"}
	pB := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "2"}

	paStr := pA.ID.String()
	scoreA, scoreB := 2, 1
	m := domain.Match{
		ID:         domain.MatchID(uuid.New()),
		PlayerA:    pA.ID,
		PlayerB:    pB.ID,
		Played:     true,
		ScoreA:     &scoreA,
		ScoreB:     &scoreB,
		Leg1Winner: paStr,
		Leg1Turns:  12,
		Leg2Winner: pB.ID.String(),
		Leg2Turns:  15,
		Leg3Winner: paStr,
		Leg3Turns:  18,
	}

	repo := newMatchesByPlayerRepo()
	repo.addForPlayer(pA.ID, m)

	uc := usecase.NewScoreUseCase(repo, nil, nil)
	stats, err := uc.GetStats(ctx, []domain.Player{pA}, nil)
	if err != nil {
		t.Fatalf("GetStats error: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}
	s := stats[0]
	if s.MinTurns != 12 {
		t.Errorf("MinTurns: got %d, want 12", s.MinTurns)
	}
	wantAvg := 15.0 // (12 + 18) / 2
	if s.AvgTurns != wantAvg {
		t.Errorf("AvgTurns: got %v, want %v", s.AvgTurns, wantAvg)
	}
}

// TestGetStats_UnplayedMatchesIgnored verifies that matches with Played==false are skipped.
func TestGetStats_UnplayedMatchesIgnored(t *testing.T) {
	ctx := context.Background()

	pA := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1"}

	m := domain.Match{
		ID:      domain.MatchID(uuid.New()),
		PlayerA: pA.ID,
		Played:  false,
	}

	repo := newMatchesByPlayerRepo()
	repo.addForPlayer(pA.ID, m)

	uc := usecase.NewScoreUseCase(repo, nil, nil)
	stats, err := uc.GetStats(ctx, []domain.Player{pA}, nil)
	if err != nil {
		t.Fatalf("GetStats error: %v", err)
	}
	if stats[0].Played != 0 {
		t.Errorf("expected Played=0 for unplayed match, got %d", stats[0].Played)
	}
}

// TestGetStats_WithScheduleID verifies scheduleID path uses FindByPlayerAndSchedule.
func TestGetStats_WithScheduleID(t *testing.T) {
	ctx := context.Background()

	pA := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1"}

	schedID := domain.ScheduleID(uuid.New())
	scoreA, scoreB := 2, 0
	m := domain.Match{
		ID:      domain.MatchID(uuid.New()),
		PlayerA: pA.ID,
		Played:  true,
		ScoreA:  &scoreA,
		ScoreB:  &scoreB,
	}

	repo := newMatchesByPlayerRepo()
	repo.addForPlayer(pA.ID, m)

	uc := usecase.NewScoreUseCase(repo, nil, nil)
	stats, err := uc.GetStats(ctx, []domain.Player{pA}, &schedID)
	if err != nil {
		t.Fatalf("GetStats error: %v", err)
	}
	if stats[0].Wins != 1 {
		t.Errorf("expected 1 win, got %d", stats[0].Wins)
	}
}

// ---------------------------------------------------------------------------
// GetDutyStats tests
// ---------------------------------------------------------------------------

// TestGetDutyStats_SecretaryAndCounter verifies secretary/counter counts per player.
func TestGetDutyStats_SecretaryAndCounter(t *testing.T) {
	ctx := context.Background()

	pA := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1", Name: "Alpha"}
	pB := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "2", Name: "Beta"}

	evID := domain.EveningID(uuid.New())
	scoreA, scoreB := 2, 0
	m := domain.Match{
		ID:          domain.MatchID(uuid.New()),
		EveningID:   evID,
		PlayerA:     pA.ID,
		PlayerB:     pB.ID,
		Played:      true,
		ScoreA:      &scoreA,
		ScoreB:      &scoreB,
		SecretaryNr: "1", // pA's nr
		CounterNr:   "2", // pB's nr
	}

	repo := newMatchesByPlayerRepo()
	repo.addForPlayer(pA.ID, m)
	repo.addForPlayer(pB.ID, m)

	uc := usecase.NewScoreUseCase(repo, &stubEveningRepo{}, nil)
	stats, err := uc.GetDutyStats(ctx, []domain.Player{pA, pB}, nil)
	if err != nil {
		t.Fatalf("GetDutyStats error: %v", err)
	}

	byNr := make(map[string]usecase.DutyStats)
	for _, s := range stats {
		byNr[s.Player.Nr] = s
	}

	if byNr["1"].SecretaryCount != 1 {
		t.Errorf("player 1 SecretaryCount: got %d, want 1", byNr["1"].SecretaryCount)
	}
	if byNr["1"].CounterCount != 0 {
		t.Errorf("player 1 CounterCount: got %d, want 0", byNr["1"].CounterCount)
	}
	if byNr["2"].CounterCount != 1 {
		t.Errorf("player 2 CounterCount: got %d, want 1", byNr["2"].CounterCount)
	}
	if byNr["1"].Count != 1 {
		t.Errorf("player 1 Count: got %d, want 1", byNr["1"].Count)
	}
}

// TestGetDutyStats_NoMatches verifies empty result when there are no played matches.
func TestGetDutyStats_NoMatches(t *testing.T) {
	ctx := context.Background()

	pA := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1"}

	repo := newMatchesByPlayerRepo()
	uc := usecase.NewScoreUseCase(repo, &stubEveningRepo{}, nil)

	stats, err := uc.GetDutyStats(ctx, []domain.Player{pA}, nil)
	if err != nil {
		t.Fatalf("GetDutyStats error: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(stats))
	}
	if stats[0].Count != 0 {
		t.Errorf("expected Count=0 for player with no duties, got %d", stats[0].Count)
	}
}

// TestGetDutyStats_WithScheduleID verifies the scheduleID path deduplicates matches.
func TestGetDutyStats_WithScheduleID(t *testing.T) {
	ctx := context.Background()

	pA := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1", Name: "Alpha"}
	pB := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "2", Name: "Beta"}

	schedID := domain.ScheduleID(uuid.New())
	evID := domain.EveningID(uuid.New())
	scoreA, scoreB := 2, 0
	m := domain.Match{
		ID:          domain.MatchID(uuid.New()),
		EveningID:   evID,
		PlayerA:     pA.ID,
		PlayerB:     pB.ID,
		Played:      true,
		ScoreA:      &scoreA,
		ScoreB:      &scoreB,
		SecretaryNr: "1",
	}

	repo := newMatchesByPlayerRepo()
	repo.addForPlayer(pA.ID, m)
	repo.addForPlayer(pB.ID, m)

	ev := domain.Evening{ID: evID, Number: 1}
	uc := usecase.NewScoreUseCase(repo, &stubEveningRepo{evenings: []domain.Evening{ev}}, nil)

	stats, err := uc.GetDutyStats(ctx, []domain.Player{pA, pB}, &schedID)
	if err != nil {
		t.Fatalf("GetDutyStats error: %v", err)
	}

	byNr := make(map[string]usecase.DutyStats)
	for _, s := range stats {
		byNr[s.Player.Nr] = s
	}

	if byNr["1"].SecretaryCount != 1 {
		t.Errorf("player 1 SecretaryCount: got %d, want 1", byNr["1"].SecretaryCount)
	}
	if len(byNr["1"].SecretaryMatches) != 1 || byNr["1"].SecretaryMatches[0].EveningNr != 1 {
		t.Errorf("player 1 SecretaryMatches: got %v", byNr["1"].SecretaryMatches)
	}
}

// TestSubmit_NotFoundError verifies an error is returned for an unknown match ID.
func TestSubmit_NotFoundError(t *testing.T) {
	ctx := context.Background()

	repo := newStubMatchRepo(nil)
	uc := usecase.NewScoreUseCase(repo, nil, nil)

	err := uc.Submit(ctx, usecase.SubmitScoreInput{
		MatchID: domain.MatchID(uuid.New()),
	})
	if err == nil {
		t.Error("expected error for unknown match, got nil")
	}
}
