package usecase_test

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"DartScheduler/domain"
	"DartScheduler/usecase"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Minimal exporter mocks
// ---------------------------------------------------------------------------

type mockExporter struct {
	called bool
}

func (e *mockExporter) Export(_ context.Context, _ domain.Schedule, _ []domain.Player, _ io.Writer) error {
	e.called = true
	return nil
}

type mockEveningExporter struct {
	called bool
}

func (e *mockEveningExporter) ExportEvening(_ context.Context, _ domain.Schedule, _ domain.Evening, _ []domain.Player, _ io.Writer) error {
	e.called = true
	return nil
}

// ---------------------------------------------------------------------------
// Repo stubs for export use case
// ---------------------------------------------------------------------------

type exportScheduleRepo struct {
	sched domain.Schedule
}

func (r *exportScheduleRepo) Save(_ context.Context, _ domain.Schedule) error { return nil }
func (r *exportScheduleRepo) FindLatest(_ context.Context) (domain.Schedule, error) {
	return r.sched, nil
}
func (r *exportScheduleRepo) FindByID(_ context.Context, _ domain.ScheduleID) (domain.Schedule, error) {
	return r.sched, nil
}
func (r *exportScheduleRepo) FindAll(_ context.Context) ([]domain.Schedule, error) {
	return []domain.Schedule{r.sched}, nil
}
func (r *exportScheduleRepo) Delete(_ context.Context, _ domain.ScheduleID) error { return nil }

type exportEveningRepo struct {
	evenings []domain.Evening
}

func (r *exportEveningRepo) Save(_ context.Context, _ domain.Evening, _ domain.ScheduleID) error {
	return nil
}
func (r *exportEveningRepo) FindByID(_ context.Context, id domain.EveningID) (domain.Evening, error) {
	for _, ev := range r.evenings {
		if ev.ID == id {
			return ev, nil
		}
	}
	return domain.Evening{}, domain.ErrNotFound
}
func (r *exportEveningRepo) FindBySchedule(_ context.Context, _ domain.ScheduleID) ([]domain.Evening, error) {
	return r.evenings, nil
}
func (r *exportEveningRepo) Delete(_ context.Context, _ domain.EveningID) error { return nil }
func (r *exportEveningRepo) DeleteBySchedule(_ context.Context, _ domain.ScheduleID) error {
	return nil
}

type exportMatchRepo struct {
	byEvening map[domain.EveningID][]domain.Match
	cancelled []domain.Match
}

func newExportMatchRepo() *exportMatchRepo {
	return &exportMatchRepo{byEvening: make(map[domain.EveningID][]domain.Match)}
}

func (r *exportMatchRepo) Save(_ context.Context, _ domain.Match) error        { return nil }
func (r *exportMatchRepo) SaveBatch(_ context.Context, _ []domain.Match) error { return nil }
func (r *exportMatchRepo) FindByID(_ context.Context, _ domain.MatchID) (domain.Match, error) {
	return domain.Match{}, domain.ErrNotFound
}
func (r *exportMatchRepo) FindByEvening(_ context.Context, id domain.EveningID) ([]domain.Match, error) {
	return r.byEvening[id], nil
}
func (r *exportMatchRepo) FindByPlayer(_ context.Context, _ domain.PlayerID) ([]domain.Match, error) {
	return nil, nil
}
func (r *exportMatchRepo) FindByPlayerAndSchedule(_ context.Context, _ domain.PlayerID, _ domain.ScheduleID) ([]domain.Match, error) {
	return nil, nil
}
func (r *exportMatchRepo) FindAllPlayed(_ context.Context) ([]domain.Match, error) { return nil, nil }
func (r *exportMatchRepo) UpdateResult(_ context.Context, _ domain.Match) error    { return nil }
func (r *exportMatchRepo) FindBySchedule(_ context.Context, _ domain.ScheduleID) ([]domain.Match, error) {
	return nil, nil
}
func (r *exportMatchRepo) FindCancelledBySchedule(_ context.Context, _ domain.ScheduleID) ([]domain.Match, error) {
	return r.cancelled, nil
}
func (r *exportMatchRepo) DeleteByEvening(_ context.Context, _ domain.EveningID) error   { return nil }
func (r *exportMatchRepo) DeleteBySchedule(_ context.Context, _ domain.ScheduleID) error { return nil }
func (r *exportMatchRepo) DeleteByPlayer(_ context.Context, _ domain.PlayerID) error     { return nil }

type exportPlayerRepo struct {
	players []domain.Player
}

func (r *exportPlayerRepo) Save(_ context.Context, _ domain.Player) error        { return nil }
func (r *exportPlayerRepo) SaveBatch(_ context.Context, _ []domain.Player) error { return nil }
func (r *exportPlayerRepo) FindByID(_ context.Context, _ domain.PlayerID) (domain.Player, error) {
	return domain.Player{}, domain.ErrNotFound
}
func (r *exportPlayerRepo) FindAll(_ context.Context) ([]domain.Player, error) { return r.players, nil }
func (r *exportPlayerRepo) Delete(_ context.Context, _ domain.PlayerID) error  { return nil }
func (r *exportPlayerRepo) DeleteAll(_ context.Context) error                  { return nil }
func (r *exportPlayerRepo) SaveBuddyPreference(_ context.Context, _ domain.BuddyPreference) error {
	return nil
}
func (r *exportPlayerRepo) FindBuddiesForPlayer(_ context.Context, _ domain.PlayerID) ([]domain.PlayerID, error) {
	return nil, nil
}
func (r *exportPlayerRepo) FindAllBuddyPairs(_ context.Context) ([]domain.BuddyPreference, error) {
	return nil, nil
}
func (r *exportPlayerRepo) DeleteBuddiesForPlayer(_ context.Context, _ domain.PlayerID) error {
	return nil
}
func (r *exportPlayerRepo) DeleteAllBuddyPairs(_ context.Context) error { return nil }

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestExport_DelegatesCorrectly(t *testing.T) {
	sched := domain.Schedule{ID: uuid.New()}
	evID := domain.EveningID(uuid.New())
	ev := domain.Evening{ID: evID, Number: 1, Date: time.Now()}

	uc := usecase.NewExportUseCase(
		&exportScheduleRepo{sched: sched},
		&exportEveningRepo{evenings: []domain.Evening{ev}},
		newExportMatchRepo(),
		&exportPlayerRepo{},
	)

	exp := &mockExporter{}
	if err := uc.Export(context.Background(), exp, &bytes.Buffer{}); err != nil {
		t.Fatalf("Export error: %v", err)
	}
	if !exp.called {
		t.Error("expected exporter.Export to be called")
	}
}

func TestEveningDate_ReturnsCorrectDate(t *testing.T) {
	sched := domain.Schedule{ID: uuid.New()}
	evID := domain.EveningID(uuid.New())
	want := time.Date(2025, 4, 7, 0, 0, 0, 0, time.UTC)
	ev := domain.Evening{ID: evID, Number: 1, Date: want}

	uc := usecase.NewExportUseCase(
		&exportScheduleRepo{sched: sched},
		&exportEveningRepo{evenings: []domain.Evening{ev}},
		newExportMatchRepo(),
		&exportPlayerRepo{},
	)

	got, err := uc.EveningDate(context.Background(), evID)
	if err != nil {
		t.Fatalf("EveningDate error: %v", err)
	}
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestEveningDate_NotFound(t *testing.T) {
	sched := domain.Schedule{ID: uuid.New()}
	uc := usecase.NewExportUseCase(
		&exportScheduleRepo{sched: sched},
		&exportEveningRepo{evenings: nil},
		newExportMatchRepo(),
		&exportPlayerRepo{},
	)

	_, err := uc.EveningDate(context.Background(), domain.EveningID(uuid.New()))
	if err == nil {
		t.Error("expected ErrNotFound, got nil")
	}
}

func TestExportEvening_RegularEvening(t *testing.T) {
	sched := domain.Schedule{ID: uuid.New()}
	evID := domain.EveningID(uuid.New())
	ev := domain.Evening{ID: evID, Number: 2, Date: time.Now(), IsCatchUpEvening: false}

	uc := usecase.NewExportUseCase(
		&exportScheduleRepo{sched: sched},
		&exportEveningRepo{evenings: []domain.Evening{ev}},
		newExportMatchRepo(),
		&exportPlayerRepo{},
	)

	exp := &mockEveningExporter{}
	if err := uc.ExportEvening(context.Background(), exp, evID, &bytes.Buffer{}); err != nil {
		t.Fatalf("ExportEvening error: %v", err)
	}
	if !exp.called {
		t.Error("expected evening exporter to be called")
	}
}

func TestExportEvening_CatchUpEvening(t *testing.T) {
	sched := domain.Schedule{ID: uuid.New()}
	evID := domain.EveningID(uuid.New())
	ev := domain.Evening{ID: evID, Number: 3, Date: time.Now(), IsCatchUpEvening: true}

	uc := usecase.NewExportUseCase(
		&exportScheduleRepo{sched: sched},
		&exportEveningRepo{evenings: []domain.Evening{ev}},
		newExportMatchRepo(),
		&exportPlayerRepo{},
	)

	exp := &mockEveningExporter{}
	if err := uc.ExportEvening(context.Background(), exp, evID, &bytes.Buffer{}); err != nil {
		t.Fatalf("ExportEvening error: %v", err)
	}
	if !exp.called {
		t.Error("expected evening exporter to be called for catch-up evening")
	}
}

func TestExportEvening_NotFound(t *testing.T) {
	sched := domain.Schedule{ID: uuid.New()}
	uc := usecase.NewExportUseCase(
		&exportScheduleRepo{sched: sched},
		&exportEveningRepo{evenings: nil},
		newExportMatchRepo(),
		&exportPlayerRepo{},
	)

	exp := &mockEveningExporter{}
	err := uc.ExportEvening(context.Background(), exp, domain.EveningID(uuid.New()), &bytes.Buffer{})
	if err == nil {
		t.Error("expected ErrNotFound for unknown evening ID")
	}
}

func TestExportEvening_RegularWithCancelledMatches(t *testing.T) {
	sched := domain.Schedule{ID: uuid.New()}
	evID := domain.EveningID(uuid.New())
	otherEvID := domain.EveningID(uuid.New())
	ev := domain.Evening{ID: evID, Number: 2, Date: time.Now(), IsCatchUpEvening: false}

	matchRepo := newExportMatchRepo()
	// cancelled match from a different evening
	matchRepo.cancelled = []domain.Match{
		{ID: domain.MatchID(uuid.New()), EveningID: otherEvID},
	}

	uc := usecase.NewExportUseCase(
		&exportScheduleRepo{sched: sched},
		&exportEveningRepo{evenings: []domain.Evening{ev}},
		matchRepo,
		&exportPlayerRepo{},
	)

	exp := &mockEveningExporter{}
	if err := uc.ExportEvening(context.Background(), exp, evID, &bytes.Buffer{}); err != nil {
		t.Fatalf("ExportEvening error: %v", err)
	}
	if !exp.called {
		t.Error("expected evening exporter to be called")
	}
}
