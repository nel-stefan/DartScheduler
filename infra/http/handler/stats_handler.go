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

// playersForSchedule loads the player list scoped to the schedule's PlayerListID when set.
func (h *StatsHandler) playersForSchedule(r *http.Request, schedID *domain.ScheduleID) ([]domain.Player, error) {
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
	players, err := h.playersForSchedule(r, schedID)
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
	players, err := h.playersForSchedule(r, schedID)
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
