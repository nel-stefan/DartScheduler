package handler

import (
	"encoding/json"
	"net/http"

	"DartScheduler/domain"
	"DartScheduler/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ScoreHandler struct {
	uc *usecase.ScoreUseCase
}

func NewScoreHandler(uc *usecase.ScoreUseCase) *ScoreHandler {
	return &ScoreHandler{uc: uc}
}

type submitScoreRequest struct {
	ScoreA int `json:"scoreA"`
	ScoreB int `json:"scoreB"`
}

func (h *ScoreHandler) Submit(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}

	var req submitScoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.uc.Submit(r.Context(), usecase.SubmitScoreInput{
		MatchID: domain.MatchID(id),
		ScoreA:  req.ScoreA,
		ScoreB:  req.ScoreB,
	}); err != nil {
		httpErrorDomain(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
