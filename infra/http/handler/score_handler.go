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
	Leg1Winner     string `json:"leg1Winner"`
	Leg1Turns      int    `json:"leg1Turns"`
	Leg2Winner     string `json:"leg2Winner"`
	Leg2Turns      int    `json:"leg2Turns"`
	Leg3Winner     string `json:"leg3Winner"`
	Leg3Turns      int    `json:"leg3Turns"`
	ReportedBy     string `json:"reportedBy"`
	RescheduleDate string `json:"rescheduleDate"`
	SecretaryNr    string `json:"secretaryNr"`
	CounterNr      string `json:"counterNr"`
	PlayerA180s          int    `json:"playerA180s"`
	PlayerB180s          int    `json:"playerB180s"`
	PlayerAHighestFinish int    `json:"playerAHighestFinish"`
	PlayerBHighestFinish int    `json:"playerBHighestFinish"`
	PlayedDate           string `json:"playedDate"`
}

type reportAbsentRequest struct {
	PlayerID   string `json:"playerId"`
	ReportedBy string `json:"reportedBy"`
}

func (h *ScoreHandler) ReportAbsent(w http.ResponseWriter, r *http.Request) {
	eveningIDStr := chi.URLParam(r, "id")
	eveningID, err := uuid.Parse(eveningIDStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	var req reportAbsentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	playerID, err := uuid.Parse(req.PlayerID)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	if err := h.uc.ReportAbsent(r.Context(), domain.EveningID(eveningID), domain.PlayerID(playerID), req.ReportedBy); err != nil {
		httpErrorDomain(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
		MatchID:        domain.MatchID(id),
		Leg1Winner:     req.Leg1Winner,
		Leg1Turns:      req.Leg1Turns,
		Leg2Winner:     req.Leg2Winner,
		Leg2Turns:      req.Leg2Turns,
		Leg3Winner:     req.Leg3Winner,
		Leg3Turns:      req.Leg3Turns,
		ReportedBy:     req.ReportedBy,
		RescheduleDate: req.RescheduleDate,
		SecretaryNr:    req.SecretaryNr,
		CounterNr:      req.CounterNr,
		PlayerA180s:          req.PlayerA180s,
		PlayerB180s:          req.PlayerB180s,
		PlayerAHighestFinish: req.PlayerAHighestFinish,
		PlayerBHighestFinish: req.PlayerBHighestFinish,
		PlayedDate:           req.PlayedDate,
	}); err != nil {
		httpErrorDomain(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
