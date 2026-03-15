package handler

import (
	"encoding/json"
	"net/http"

	"DartScheduler/domain"
	"DartScheduler/infra/excel"
	"DartScheduler/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

type updatePlayerRequest struct {
	Nr          string `json:"nr"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Sponsor     string `json:"sponsor"`
	Address     string `json:"address"`
	PostalCode  string `json:"postalCode"`
	City        string `json:"city"`
	Phone       string `json:"phone"`
	Mobile      string `json:"mobile"`
	MemberSince string `json:"memberSince"`
	Class       string `json:"class"`
}

func (h *PlayerHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	var req updatePlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	p := domain.Player{
		ID:          id,
		Nr:          req.Nr,
		Name:        req.Name,
		Email:       req.Email,
		Sponsor:     req.Sponsor,
		Address:     req.Address,
		PostalCode:  req.PostalCode,
		City:        req.City,
		Phone:       req.Phone,
		Mobile:      req.Mobile,
		MemberSince: req.MemberSince,
		Class:       req.Class,
	}
	if err := h.uc.UpdatePlayer(r.Context(), p); err != nil {
		httpErrorDomain(w, err)
		return
	}
	writeJSON(w, p)
}

type setBuddiesRequest struct {
	BuddyIDs []string `json:"buddyIds"`
}

func (h *PlayerHandler) GetBuddies(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	buddyIDs, err := h.uc.GetBuddies(r.Context(), id)
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	// Return as string slice
	out := make([]string, len(buddyIDs))
	for i, bid := range buddyIDs {
		out[i] = bid.String()
	}
	writeJSON(w, out)
}

func (h *PlayerHandler) SetBuddies(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	var req setBuddiesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	buddyIDs := make([]domain.PlayerID, 0, len(req.BuddyIDs))
	for _, s := range req.BuddyIDs {
		bid, err := uuid.Parse(s)
		if err != nil {
			httpError(w, err, http.StatusBadRequest)
			return
		}
		buddyIDs = append(buddyIDs, bid)
	}
	if err := h.uc.SetBuddies(r.Context(), id, buddyIDs); err != nil {
		httpErrorDomain(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
