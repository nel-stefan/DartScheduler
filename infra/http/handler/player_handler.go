package handler

import (
	"encoding/json"
	"net/http"

	"DartScheduler/infra/excel"
	"DartScheduler/usecase"
)

type PlayerHandler struct {
	uc *usecase.PlayerUseCase
}

func NewPlayerHandler(uc *usecase.PlayerUseCase) *PlayerHandler {
	return &PlayerHandler{uc: uc}
}

func (h *PlayerHandler) Import(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	defer file.Close()

	players, err := excel.ImportPlayers(file)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.uc.ImportPlayers(r.Context(), players, nil); err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"imported": len(players)})
}

func (h *PlayerHandler) List(w http.ResponseWriter, r *http.Request) {
	players, err := h.uc.ListPlayers(r.Context())
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	writeJSON(w, players)
}
