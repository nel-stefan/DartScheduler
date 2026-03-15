package handler

import (
	"net/http"

	"DartScheduler/domain"
	"DartScheduler/usecase"

	"github.com/google/uuid"
)

type StatsHandler struct {
	players domain.PlayerRepository
	uc      *usecase.ScoreUseCase
}

func NewStatsHandler(players domain.PlayerRepository, uc *usecase.ScoreUseCase) *StatsHandler {
	return &StatsHandler{players: players, uc: uc}
}

func (h *StatsHandler) Get(w http.ResponseWriter, r *http.Request) {
	players, err := h.players.FindAll(r.Context())
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
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
	stats, err := h.uc.GetStats(r.Context(), players, schedID)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	writeJSON(w, stats)
}

func (h *StatsHandler) GetDuties(w http.ResponseWriter, r *http.Request) {
	players, err := h.players.FindAll(r.Context())
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
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
	stats, err := h.uc.GetDutyStats(r.Context(), players, schedID)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	writeJSON(w, stats)
}
