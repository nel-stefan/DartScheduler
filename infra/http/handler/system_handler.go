package handler

import (
	"net/http"

	"DartScheduler/infra/logbuf"
)

type SystemHandler struct {
	buf *logbuf.Buffer
}

func NewSystemHandler(buf *logbuf.Buffer) *SystemHandler {
	return &SystemHandler{buf: buf}
}

func (h *SystemHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"logs": h.buf.Lines()})
}
