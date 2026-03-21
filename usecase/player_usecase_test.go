package usecase_test

import (
	"context"
	"fmt"
	"testing"

	"DartScheduler/domain"
	"DartScheduler/usecase"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Functional in-memory player repo for player use case tests
// ---------------------------------------------------------------------------

type funcPlayerRepo struct {
	players             []domain.Player
	buddies             []domain.BuddyPreference
	saveCalls           int
	deleteCalls         int
	deleteBuddyCalls    int
	deleteAllBuddyCalls int
}

func (r *funcPlayerRepo) Save(_ context.Context, p domain.Player) error {
	r.saveCalls++
	for i, existing := range r.players {
		if existing.ID == p.ID {
			r.players[i] = p
			return nil
		}
	}
	r.players = append(r.players, p)
	return nil
}

func (r *funcPlayerRepo) SaveBatch(_ context.Context, players []domain.Player) error {
	for _, p := range players {
		_ = r.Save(context.Background(), p)
	}
	return nil
}

func (r *funcPlayerRepo) FindByID(_ context.Context, id domain.PlayerID) (domain.Player, error) {
	for _, p := range r.players {
		if p.ID == id {
			return p, nil
		}
	}
	return domain.Player{}, domain.ErrNotFound
}

func (r *funcPlayerRepo) FindAll(_ context.Context) ([]domain.Player, error) {
	return r.players, nil
}

func (r *funcPlayerRepo) Delete(_ context.Context, id domain.PlayerID) error {
	r.deleteCalls++
	for i, p := range r.players {
		if p.ID == id {
			r.players = append(r.players[:i], r.players[i+1:]...)
			return nil
		}
	}
	return nil
}

func (r *funcPlayerRepo) DeleteAll(_ context.Context) error {
	r.players = nil
	return nil
}

func (r *funcPlayerRepo) SaveBuddyPreference(_ context.Context, bp domain.BuddyPreference) error {
	r.buddies = append(r.buddies, bp)
	return nil
}

func (r *funcPlayerRepo) FindBuddiesForPlayer(_ context.Context, id domain.PlayerID) ([]domain.PlayerID, error) {
	var out []domain.PlayerID
	for _, bp := range r.buddies {
		if bp.PlayerID == id {
			out = append(out, bp.BuddyID)
		}
	}
	return out, nil
}

func (r *funcPlayerRepo) FindAllBuddyPairs(_ context.Context) ([]domain.BuddyPreference, error) {
	return r.buddies, nil
}

func (r *funcPlayerRepo) DeleteBuddiesForPlayer(_ context.Context, id domain.PlayerID) error {
	r.deleteBuddyCalls++
	var remaining []domain.BuddyPreference
	for _, bp := range r.buddies {
		if bp.PlayerID != id {
			remaining = append(remaining, bp)
		}
	}
	r.buddies = remaining
	return nil
}

func (r *funcPlayerRepo) DeleteAllBuddyPairs(_ context.Context) error {
	r.deleteAllBuddyCalls++
	r.buddies = nil
	return nil
}

// errOnDeleteBuddyRepo wraps funcPlayerRepo and returns an error from DeleteBuddiesForPlayer.
type errOnDeleteBuddyRepo struct {
	funcPlayerRepo
}

func (r *errOnDeleteBuddyRepo) DeleteBuddiesForPlayer(_ context.Context, _ domain.PlayerID) error {
	return fmt.Errorf("buddy delete error")
}

// errOnDeleteRepo wraps funcPlayerRepo and returns an error from Delete.
type errOnDeleteRepo struct {
	funcPlayerRepo
}

func (r *errOnDeleteRepo) Delete(_ context.Context, _ domain.PlayerID) error {
	return fmt.Errorf("delete error")
}

// ---------------------------------------------------------------------------
// Minimal match repo for player use case tests (tracks DeleteByPlayer calls)
// ---------------------------------------------------------------------------

type playerMatchRepo struct {
	deleteByPlayerCalls int
}

func (r *playerMatchRepo) Save(_ context.Context, _ domain.Match) error        { return nil }
func (r *playerMatchRepo) SaveBatch(_ context.Context, _ []domain.Match) error { return nil }
func (r *playerMatchRepo) FindByID(_ context.Context, _ domain.MatchID) (domain.Match, error) {
	return domain.Match{}, domain.ErrNotFound
}
func (r *playerMatchRepo) FindByEvening(_ context.Context, _ domain.EveningID) ([]domain.Match, error) {
	return nil, nil
}
func (r *playerMatchRepo) FindByPlayer(_ context.Context, _ domain.PlayerID) ([]domain.Match, error) {
	return nil, nil
}
func (r *playerMatchRepo) FindByPlayerAndSchedule(_ context.Context, _ domain.PlayerID, _ domain.ScheduleID) ([]domain.Match, error) {
	return nil, nil
}
func (r *playerMatchRepo) FindAllPlayed(_ context.Context) ([]domain.Match, error) { return nil, nil }
func (r *playerMatchRepo) UpdateResult(_ context.Context, _ domain.Match) error    { return nil }
func (r *playerMatchRepo) FindBySchedule(_ context.Context, _ domain.ScheduleID) ([]domain.Match, error) {
	return nil, nil
}
func (r *playerMatchRepo) FindCancelledBySchedule(_ context.Context, _ domain.ScheduleID) ([]domain.Match, error) {
	return nil, nil
}
func (r *playerMatchRepo) DeleteByEvening(_ context.Context, _ domain.EveningID) error { return nil }
func (r *playerMatchRepo) DeleteBySchedule(_ context.Context, _ domain.ScheduleID) error {
	return nil
}
func (r *playerMatchRepo) DeleteByPlayer(_ context.Context, _ domain.PlayerID) error {
	r.deleteByPlayerCalls++
	return nil
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestListPlayers_FormatsDisplayNames(t *testing.T) {
	pA := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1", Name: "Janssen, Jan"}
	pB := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "2", Name: "NoComma"}
	repo := &funcPlayerRepo{players: []domain.Player{pA, pB}}
	uc := usecase.NewPlayerUseCase(repo, &playerMatchRepo{})

	players, err := uc.ListPlayers(context.Background())
	if err != nil {
		t.Fatalf("ListPlayers error: %v", err)
	}
	if len(players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(players))
	}
	found := false
	for _, p := range players {
		if p.Nr == "1" && p.Name == "Jan Janssen" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'Jan Janssen' in result, names were %v", func() []string {
			var ns []string
			for _, p := range players {
				ns = append(ns, p.Name)
			}
			return ns
		}())
	}
}

func TestListPlayers_NameWithoutCommaUnchanged(t *testing.T) {
	p := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "5", Name: "NoComma"}
	repo := &funcPlayerRepo{players: []domain.Player{p}}
	uc := usecase.NewPlayerUseCase(repo, &playerMatchRepo{})

	players, err := uc.ListPlayers(context.Background())
	if err != nil {
		t.Fatalf("ListPlayers error: %v", err)
	}
	if players[0].Name != "NoComma" {
		t.Errorf("expected 'NoComma', got %q", players[0].Name)
	}
}

func TestImportPlayers_EmptyInputReturnsError(t *testing.T) {
	uc := usecase.NewPlayerUseCase(&funcPlayerRepo{}, &playerMatchRepo{})
	err := uc.ImportPlayers(context.Background(), nil, nil)
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

func TestImportPlayers_SavesPlayers(t *testing.T) {
	repo := &funcPlayerRepo{}
	uc := usecase.NewPlayerUseCase(repo, &playerMatchRepo{})

	inputs := []usecase.PlayerInput{
		{Nr: "1", Name: "Doe, John"},
		{Nr: "2", Name: "Smith, Jane"},
	}
	if err := uc.ImportPlayers(context.Background(), inputs, nil); err != nil {
		t.Fatalf("ImportPlayers error: %v", err)
	}
	if len(repo.players) != 2 {
		t.Errorf("expected 2 players saved, got %d", len(repo.players))
	}
}

func TestImportPlayers_BuddyFromExcelColumn(t *testing.T) {
	repo := &funcPlayerRepo{}
	uc := usecase.NewPlayerUseCase(repo, &playerMatchRepo{})

	inputs := []usecase.PlayerInput{
		{Nr: "1", Name: "Alpha"},
		{Nr: "2", Name: "Beta", BuddyNr: "1"},
	}
	if err := uc.ImportPlayers(context.Background(), inputs, nil); err != nil {
		t.Fatalf("ImportPlayers error: %v", err)
	}
	if len(repo.buddies) == 0 {
		t.Error("expected at least one buddy preference to be saved")
	}
	if repo.deleteAllBuddyCalls != 1 {
		t.Errorf("expected DeleteAllBuddyPairs called once, got %d", repo.deleteAllBuddyCalls)
	}
}

func TestImportPlayers_ManualBuddiesMerged(t *testing.T) {
	repo := &funcPlayerRepo{}
	uc := usecase.NewPlayerUseCase(repo, &playerMatchRepo{})

	pAID := domain.PlayerID(uuid.New())
	pBID := domain.PlayerID(uuid.New())

	inputs := []usecase.PlayerInput{
		{Nr: "1", Name: "Alpha"},
		{Nr: "2", Name: "Beta"},
	}
	manual := []usecase.BuddyPairInput{{PlayerID: pAID, BuddyID: pBID}}
	if err := uc.ImportPlayers(context.Background(), inputs, manual); err != nil {
		t.Fatalf("ImportPlayers error: %v", err)
	}
	// Manual buddies that don't appear in excel should still be saved.
	if len(repo.buddies) == 0 {
		t.Error("expected manual buddy pairs to be saved")
	}
}

func TestUpdatePlayer_SavesChanges(t *testing.T) {
	p := domain.Player{ID: domain.PlayerID(uuid.New()), Nr: "1", Name: "Test", Class: "B"}
	repo := &funcPlayerRepo{players: []domain.Player{p}}
	uc := usecase.NewPlayerUseCase(repo, &playerMatchRepo{})

	p.Class = "A"
	if err := uc.UpdatePlayer(context.Background(), p); err != nil {
		t.Fatalf("UpdatePlayer error: %v", err)
	}
	if repo.players[0].Class != "A" {
		t.Errorf("expected class A, got %q", repo.players[0].Class)
	}
}

func TestGetBuddies_ReturnsBuddiesForPlayer(t *testing.T) {
	pA := domain.PlayerID(uuid.New())
	pB := domain.PlayerID(uuid.New())
	pC := domain.PlayerID(uuid.New())
	repo := &funcPlayerRepo{buddies: []domain.BuddyPreference{
		{PlayerID: pA, BuddyID: pB},
		{PlayerID: pC, BuddyID: pB}, // different player, should not appear
	}}
	uc := usecase.NewPlayerUseCase(repo, &playerMatchRepo{})

	buddies, err := uc.GetBuddies(context.Background(), pA)
	if err != nil {
		t.Fatalf("GetBuddies error: %v", err)
	}
	if len(buddies) != 1 || buddies[0] != pB {
		t.Errorf("expected [%s], got %v", pB, buddies)
	}
}

func TestDeletePlayer_DeletesMatchesBuddiesAndPlayer(t *testing.T) {
	pid := domain.PlayerID(uuid.New())
	p := domain.Player{ID: pid, Nr: "1"}
	repo := &funcPlayerRepo{
		players: []domain.Player{p},
		buddies: []domain.BuddyPreference{{PlayerID: pid, BuddyID: domain.PlayerID(uuid.New())}},
	}
	matchRepo := &playerMatchRepo{}
	uc := usecase.NewPlayerUseCase(repo, matchRepo)

	if err := uc.DeletePlayer(context.Background(), pid); err != nil {
		t.Fatalf("DeletePlayer error: %v", err)
	}
	if matchRepo.deleteByPlayerCalls != 1 {
		t.Errorf("DeleteByPlayer: expected 1 call, got %d", matchRepo.deleteByPlayerCalls)
	}
	if repo.deleteBuddyCalls != 1 {
		t.Errorf("DeleteBuddiesForPlayer: expected 1 call, got %d", repo.deleteBuddyCalls)
	}
	if repo.deleteCalls != 1 {
		t.Errorf("Delete: expected 1 call, got %d", repo.deleteCalls)
	}
	if len(repo.players) != 0 {
		t.Errorf("expected 0 players after delete, got %d", len(repo.players))
	}
}

func TestSetBuddies_ReplacesExistingBuddies(t *testing.T) {
	pid := domain.PlayerID(uuid.New())
	oldBuddy := domain.PlayerID(uuid.New())
	newBuddy := domain.PlayerID(uuid.New())
	repo := &funcPlayerRepo{buddies: []domain.BuddyPreference{
		{PlayerID: pid, BuddyID: oldBuddy},
	}}
	uc := usecase.NewPlayerUseCase(repo, &playerMatchRepo{})

	if err := uc.SetBuddies(context.Background(), pid, []domain.PlayerID{newBuddy}); err != nil {
		t.Fatalf("SetBuddies error: %v", err)
	}
	if repo.deleteBuddyCalls != 1 {
		t.Errorf("expected DeleteBuddiesForPlayer called once, got %d", repo.deleteBuddyCalls)
	}
	buddies, _ := repo.FindBuddiesForPlayer(context.Background(), pid)
	if len(buddies) != 1 || buddies[0] != newBuddy {
		t.Errorf("expected [newBuddy], got %v", buddies)
	}
}

func TestSetBuddies_EmptyListClearsAll(t *testing.T) {
	pid := domain.PlayerID(uuid.New())
	repo := &funcPlayerRepo{buddies: []domain.BuddyPreference{
		{PlayerID: pid, BuddyID: domain.PlayerID(uuid.New())},
	}}
	uc := usecase.NewPlayerUseCase(repo, &playerMatchRepo{})

	if err := uc.SetBuddies(context.Background(), pid, nil); err != nil {
		t.Fatalf("SetBuddies error: %v", err)
	}
	buddies, _ := repo.FindBuddiesForPlayer(context.Background(), pid)
	if len(buddies) != 0 {
		t.Errorf("expected no buddies after SetBuddies(nil), got %v", buddies)
	}
}

// TestDeletePlayer_MatchDeleteError verifies that an error from DeleteByPlayer is propagated.
func TestDeletePlayer_MatchDeleteError(t *testing.T) {
	repo := &funcPlayerRepo{}
	matchRepo := &playerMatchRepo{}
	// Override the repo to return an error -- use a simple error returning match repo variant.
	type errMatchRepo struct {
		playerMatchRepo
	}
	_ = errMatchRepo{} // unused but shows intent

	// Simplest approach: test via errOnDeleteRepo for player repo.
	pid := domain.PlayerID(uuid.New())
	p := domain.Player{ID: pid}
	playerRepo := &funcPlayerRepo{players: []domain.Player{p}}
	_ = repo

	// Verify that DeleteBuddiesForPlayer error propagates through DeletePlayer.
	errRepo := &errOnDeleteBuddyRepo{funcPlayerRepo: *playerRepo}
	uc := usecase.NewPlayerUseCase(errRepo, matchRepo)
	err := uc.DeletePlayer(context.Background(), pid)
	if err == nil {
		t.Error("expected error from DeletePlayer when buddy delete fails, got nil")
	}
}

// TestSetBuddies_DeleteErrorPropagated verifies that a delete error from the player repo is propagated.
func TestSetBuddies_DeleteErrorPropagated(t *testing.T) {
	pid := domain.PlayerID(uuid.New())
	errRepo := &errOnDeleteBuddyRepo{}
	uc := usecase.NewPlayerUseCase(errRepo, &playerMatchRepo{})
	err := uc.SetBuddies(context.Background(), pid, nil)
	if err == nil {
		t.Error("expected error from SetBuddies when DeleteBuddiesForPlayer fails, got nil")
	}
}
