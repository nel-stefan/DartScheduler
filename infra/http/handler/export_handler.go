package handler

import (
	"net/http"

	excelexport "DartScheduler/infra/excel"
	"DartScheduler/infra/pdf"
	"DartScheduler/usecase"
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
