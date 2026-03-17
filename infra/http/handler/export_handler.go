package handler

import (
	"fmt"
	"net/http"

	excelexport "DartScheduler/infra/excel"
	"DartScheduler/infra/pdf"
	"DartScheduler/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ExportHandler struct {
	uc *usecase.ExportUseCase
}

func NewExportHandler(uc *usecase.ExportUseCase) *ExportHandler {
	return &ExportHandler{uc: uc}
}

func (h *ExportHandler) Excel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="schedule.xlsx"`)
	if err := h.uc.Export(r.Context(), excelexport.Exporter{}, w); err != nil {
		httpErrorDomain(w, err)
	}
}

func (h *ExportHandler) PDF(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="schedule.pdf"`)
	if err := h.uc.Export(r.Context(), pdf.Exporter{}, w); err != nil {
		httpErrorDomain(w, err)
	}
}

func (h *ExportHandler) EveningExcel(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	date, err := h.uc.EveningDate(r.Context(), id)
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	filename := fmt.Sprintf("wedstrijdformulier_%s.xlsx", date.Format("2006-01-02"))
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	if err := h.uc.ExportEvening(r.Context(), excelexport.EveningExporter{}, id, w); err != nil {
		httpErrorDomain(w, err)
	}
}
