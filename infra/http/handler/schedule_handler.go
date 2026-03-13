package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"DartScheduler/domain"
	"DartScheduler/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ScheduleHandler struct {
	uc *usecase.ScheduleUseCase
}

func NewScheduleHandler(uc *usecase.ScheduleUseCase) *ScheduleHandler {
	return &ScheduleHandler{uc: uc}
}

type generateRequest struct {
	CompetitionName string `json:"competitionName"`
	NumEvenings     int    `json:"numEvenings"`
	StartDate       string `json:"startDate"`    // "2026-04-01"
	IntervalDays    int    `json:"intervalDays"`
}

func (h *ScheduleHandler) Generate(w http.ResponseWriter, r *http.Request) {
	var req generateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	if req.IntervalDays <= 0 {
		req.IntervalDays = 7
	}

	sched, err := h.uc.Generate(r.Context(), usecase.GenerateScheduleInput{
		CompetitionName: req.CompetitionName,
		NumEvenings:     req.NumEvenings,
		StartDate:       startDate,
		IntervalDays:    req.IntervalDays,
	})
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	writeJSON(w, sched)
}

func (h *ScheduleHandler) Get(w http.ResponseWriter, r *http.Request) {
	sched, err := h.uc.GetLatest(r.Context())
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	writeJSON(w, sched)
}

func (h *ScheduleHandler) GetEvening(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	sched, err := h.uc.GetLatest(r.Context())
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	for _, ev := range sched.Evenings {
		if ev.ID == domain.EveningID(id) {
			writeJSON(w, ev)
			return
		}
	}
	http.Error(w, "evening not found", http.StatusNotFound)
}
