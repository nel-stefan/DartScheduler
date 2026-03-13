package handler

import (
	"net/http"

	"DartScheduler/domain"
	"DartScheduler/usecase"
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
	stats, err := h.uc.GetStats(r.Context(), players)
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	writeJSON(w, stats)
}
