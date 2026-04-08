package handler

import (
	"encoding/json"
	"log"
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
	CatchUpNrs      []int  `json:"inhaalNrs"`
	SkipNrs         []int  `json:"vrijeNrs"`
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

	GlobalProgress.Start(1_200_000) // initial estimate; updated to actual total on first progress callback
	defer GlobalProgress.Done()

	sched, err := h.uc.Generate(r.Context(), usecase.GenerateScheduleInput{
		CompetitionName: req.CompetitionName,
		Season:          req.Season,
		NumEvenings:     req.NumEvenings,
		StartDate:       startDate,
		IntervalDays:    req.IntervalDays,
		CatchUpNrs:      req.CatchUpNrs,
		SkipNrs:         req.SkipNrs,
		ProgressFn: func(step, total int) {
			GlobalProgress.SetTotal(total)
			GlobalProgress.Update(step)
		},
	})
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	log.Printf("[INFO] schema gegenereerd seizoen=%q avonden=%d id=%s", sched.CompetitionName, len(sched.Evenings), sched.ID)
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

type addCatchUpEveningRequest struct {
	Date string `json:"date"` // "2026-03-15"
}

func (h *ScheduleHandler) AddCatchUpEvening(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	schedID, err := uuid.Parse(idStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	var req addCatchUpEveningRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	sched, err := h.uc.AddCatchUpEvening(r.Context(), domain.ScheduleID(schedID), date)
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	writeJSON(w, sched)
}

func (h *ScheduleHandler) RegenerateSchedule(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}

	GlobalProgress.Start(1_200_000) // initial estimate; updated to actual total on first progress callback
	defer GlobalProgress.Done()

	sched, err := h.uc.Regenerate(r.Context(), domain.ScheduleID(id), func(step, total int) {
		GlobalProgress.SetTotal(total)
		GlobalProgress.Update(step)
	})
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	log.Printf("[INFO] schema herberekend id=%s", sched.ID)
	writeJSON(w, sched)
}

func (h *ScheduleHandler) RenameSchedule(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	var body struct {
		CompetitionName string `json:"competitionName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	if err := h.uc.RenameSchedule(r.Context(), domain.ScheduleID(id), body.CompetitionName); err != nil {
		httpErrorDomain(w, err)
		return
	}
	log.Printf("[INFO] seizoen hernoemd id=%s naam=%q", id, body.CompetitionName)
	w.WriteHeader(http.StatusNoContent)
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
	log.Printf("[INFO] schema verwijderd id=%s", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ScheduleHandler) SetActive(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	if err := h.uc.SetActive(r.Context(), domain.ScheduleID(id)); err != nil {
		httpErrorDomain(w, err)
		return
	}
	log.Printf("[INFO] actief seizoen ingesteld id=%s", id)
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

func (h *ScheduleHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	info, err := h.uc.GetInfo(r.Context(), domain.ScheduleID(id))
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	writeJSON(w, info)
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
		competitionName = "Imported season"
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
	catchUpEvenings := make([]usecase.CatchUpEvening, len(imported.CatchUpEvenings))
	for i, ie := range imported.CatchUpEvenings {
		catchUpEvenings[i] = usecase.CatchUpEvening{EveningNr: ie.EveningNr, Date: ie.Date}
	}

	sched, err := h.uc.ImportSeason(r.Context(), competitionName, season, rows, catchUpEvenings)
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	log.Printf("[INFO] seizoen geïmporteerd naam=%q avonden=%d", competitionName, len(sched.Evenings))
	writeJSON(w, sched)
}
