package handler

import (
	"net/http"

	"DartScheduler/usecase"
)

// PlayerListHandler serves player list endpoints.
type PlayerListHandler struct {
	uc *usecase.PlayerUseCase
}

func NewPlayerListHandler(uc *usecase.PlayerUseCase) *PlayerListHandler {
	return &PlayerListHandler{uc: uc}
}

// List returns all named player lists.
func (h *PlayerListHandler) List(w http.ResponseWriter, r *http.Request) {
	lists, err := h.uc.ListPlayerLists(r.Context())
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	writeJSON(w, lists)
}
