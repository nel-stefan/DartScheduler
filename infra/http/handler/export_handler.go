package handler

import (
	"fmt"
	"net/http"

	excelexport "DartScheduler/infra/excel"
	htmlexport "DartScheduler/infra/html"
	"DartScheduler/infra/pdf"
	"DartScheduler/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ExportHandler struct {
	uc           *usecase.ExportUseCase
	excelEvExp   excelexport.EveningExporter
	pdfEvExp     pdf.EveningExporter
	htmlEvPrinter htmlexport.EveningPrinter
}

func NewExportHandler(uc *usecase.ExportUseCase, clubName, logoPath string) *ExportHandler {
	return &ExportHandler{
		uc:            uc,
		excelEvExp:    excelexport.EveningExporter{ClubName: clubName},
		pdfEvExp:      pdf.EveningExporter{ClubName: clubName, LogoPath: logoPath},
		htmlEvPrinter: htmlexport.EveningPrinter{ClubName: clubName, LogoPath: logoPath},
	}
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

func (h *ExportHandler) EveningPrint(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httpError(w, err, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.uc.ExportEvening(r.Context(), h.htmlEvPrinter, id, w); err != nil {
		httpErrorDomain(w, err)
	}
}

func (h *ExportHandler) EveningPDF(w http.ResponseWriter, r *http.Request) {
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
	filename := fmt.Sprintf("wedstrijdformulier_%s.pdf", date.Format("2006-01-02"))
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	if err := h.uc.ExportEvening(r.Context(), h.pdfEvExp, id, w); err != nil {
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
	if err := h.uc.ExportEvening(r.Context(), h.excelEvExp, id, w); err != nil {
		httpErrorDomain(w, err)
	}
}
