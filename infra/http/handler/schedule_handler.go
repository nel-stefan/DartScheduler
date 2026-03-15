package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"DartScheduler/domain"
	"DartScheduler/infra/excel"
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
	Season          string `json:"season"`
	NumEvenings     int    `json:"numEvenings"`
	StartDate       string `json:"startDate"` // "2026-04-01"
	IntervalDays    int    `json:"intervalDays"`
	InhaalNrs       []int  `json:"inhaalNrs"`
	VrijeNrs        []int  `json:"vrijeNrs"`
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
		Season:          req.Season,
		NumEvenings:     req.NumEvenings,
		StartDate:       startDate,
		IntervalDays:    req.IntervalDays,
		InhaalNrs:       req.InhaalNrs,
		VrijeNrs:        req.VrijeNrs,
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

func (h *ScheduleHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.uc.ListSchedules(r.Context())
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	writeJSON(w, list)
}

func (h *ScheduleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	sched, err := h.uc.GetByID(r.Context(), domain.ScheduleID(id))
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	writeJSON(w, sched)
}

type addInhaalAvondRequest struct {
	Date string `json:"date"` // "2026-03-15"
}

func (h *ScheduleHandler) AddInhaalAvond(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	schedID, err := uuid.Parse(idStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	var req addInhaalAvondRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	sched, err := h.uc.AddInhaalAvond(r.Context(), domain.ScheduleID(schedID), date)
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	writeJSON(w, sched)
}

func (h *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	if err := h.uc.DeleteSchedule(r.Context(), domain.ScheduleID(id)); err != nil {
		httpErrorDomain(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ScheduleHandler) DeleteEvening(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "eveningId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	if err := h.uc.DeleteEvening(r.Context(), domain.EveningID(id)); err != nil {
		httpErrorDomain(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ScheduleHandler) ImportSeason(w http.ResponseWriter, r *http.Request) {
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

	competitionName := r.FormValue("competitionName")
	season := r.FormValue("season")
	if competitionName == "" {
		competitionName = "Geïmporteerd seizoen"
	}
	if season == "" {
		season = competitionName
	}

	imported, err := excel.ImportSeason(file)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}

	// Convert excel.SeasonMatchRow → usecase.SeasonMatchRow
	rows := make([]usecase.SeasonMatchRow, len(imported.Matches))
	for i, er := range imported.Matches {
		rows[i] = usecase.SeasonMatchRow{
			EveningNr:  er.EveningNr,
			Date:       er.Date,
			NrA:        er.NrA,
			NameA:      er.NameA,
			NrB:        er.NrB,
			NameB:      er.NameB,
			Leg1Winner: er.Leg1Winner,
			Leg1Turns:  er.Leg1Turns,
			Leg2Winner: er.Leg2Winner,
			Leg2Turns:  er.Leg2Turns,
			Leg3Winner: er.Leg3Winner,
			Leg3Turns:  er.Leg3Turns,
			ScoreA:     er.ScoreA,
			ScoreB:     er.ScoreB,
			Secretary:  er.Secretary,
			Counter:    er.Counter,
		}
	}
	inhaalEvenings := make([]usecase.InhaalEvening, len(imported.InhaalEvenings))
	for i, ie := range imported.InhaalEvenings {
		inhaalEvenings[i] = usecase.InhaalEvening{EveningNr: ie.EveningNr, Date: ie.Date}
	}

	sched, err := h.uc.ImportSeason(r.Context(), competitionName, season, rows, inhaalEvenings)
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	writeJSON(w, sched)
}
