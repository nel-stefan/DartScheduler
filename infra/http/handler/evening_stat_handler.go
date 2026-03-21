package handler

import (
	"encoding/json"
	"net/http"

	"DartScheduler/domain"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type EveningStatHandler struct {
	repo domain.EveningPlayerStatRepository
}

func NewEveningStatHandler(repo domain.EveningPlayerStatRepository) *EveningStatHandler {
	return &EveningStatHandler{repo: repo}
}

func (h *EveningStatHandler) GetByEvening(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	stats, err := h.repo.FindByEvening(r.Context(), domain.EveningID(id))
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	type row struct {
		PlayerID      string `json:"playerId"`
		OneEighties   int    `json:"oneEighties"`
		HighestFinish int    `json:"highestFinish"`
	}
	out := make([]row, 0, len(stats))
	for _, s := range stats {
		out = append(out, row{s.PlayerID.String(), s.OneEighties, s.HighestFinish})
	}
	writeJSON(w, out)
}

func (h *EveningStatHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	eveningID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	playerID, err := uuid.Parse(chi.URLParam(r, "playerId"))
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	var body struct {
		OneEighties   int `json:"oneEighties"`
		HighestFinish int `json:"highestFinish"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	stat := domain.EveningPlayerStat{
		EveningID:     domain.EveningID(eveningID),
		PlayerID:      domain.PlayerID(playerID),
		OneEighties:   body.OneEighties,
		HighestFinish: body.HighestFinish,
	}
	if err := h.repo.Upsert(r.Context(), stat); err != nil {
		httpErrorDomain(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
