package handler

import (
	"net/http"

	"DartScheduler/domain"
	"DartScheduler/usecase"

	"github.com/google/uuid"
)

func formatPlayerNames(players []domain.Player) []domain.Player {
	for i := range players {
		players[i].Name = domain.FormatDisplayName(players[i].Name)
	}
	return players
}

type StatsHandler struct {
	players   domain.PlayerRepository
	schedules domain.ScheduleRepository
	uc        *usecase.ScoreUseCase
}

func NewStatsHandler(players domain.PlayerRepository, schedules domain.ScheduleRepository, uc *usecase.ScoreUseCase) *StatsHandler {
	return &StatsHandler{players: players, schedules: schedules, uc: uc}
}

// playersForRequest loads the player list using listId param first, then falls back
// to the schedule's own PlayerListID, then to FindAll.
func (h *StatsHandler) playersForRequest(r *http.Request, schedID *domain.ScheduleID) ([]domain.Player, error) {
	// Explicit listId param from frontend (handles schedule-less fallback via SeasonService)
	if lidStr := r.URL.Query().Get("listId"); lidStr != "" {
		lid, err := uuid.Parse(lidStr)
		if err != nil {
			return nil, err
		}
		listID := domain.PlayerListID(lid)
		return h.players.FindByList(r.Context(), listID)
	}
	// Fall back to the schedule's own PlayerListID
	if schedID != nil {
		sched, err := h.schedules.FindByID(r.Context(), *schedID)
		if err == nil && sched.PlayerListID != nil {
			return h.players.FindByList(r.Context(), *sched.PlayerListID)
		}
	}
	return h.players.FindAll(r.Context())
}

func (h *StatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	var schedID *domain.ScheduleID
	if sidStr := r.URL.Query().Get("scheduleId"); sidStr != "" {
		uid, err := uuid.Parse(sidStr)
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		sid := domain.ScheduleID(uid)
		schedID = &sid
	}
	players, err := h.playersForRequest(r, schedID)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	stats, err := h.uc.GetStats(r.Context(), formatPlayerNames(players), schedID)
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	writeJSON(w, stats)
}

func (h *StatsHandler) GetDuties(w http.ResponseWriter, r *http.Request) {
	var schedID *domain.ScheduleID
	if sidStr := r.URL.Query().Get("scheduleId"); sidStr != "" {
		uid, err := uuid.Parse(sidStr)
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		sid := domain.ScheduleID(uid)
		schedID = &sid
	}
	players, err := h.playersForRequest(r, schedID)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	stats, err := h.uc.GetDutyStats(r.Context(), formatPlayerNames(players), schedID)
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	writeJSON(w, stats)
}
