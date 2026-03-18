package handler

import (
	"log"
	"net/http"
	"os/exec"

	"DartScheduler/infra/logbuf"
)

type SystemHandler struct {
	buf          *logbuf.Buffer
	deployScript string
}

func NewSystemHandler(buf *logbuf.Buffer, deployScript string) *SystemHandler {
	return &SystemHandler{buf: buf, deployScript: deployScript}
}

func (h *SystemHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"logs": h.buf.Lines()})
}

// TriggerDeploy runs the configured deploy script in the background and returns 202 immediately.
func (h *SystemHandler) TriggerDeploy(w http.ResponseWriter, r *http.Request) {
	if h.deployScript == "" {
		http.Error(w, "deploy script not configured", http.StatusNotImplemented)
		return
	}
	log.Printf("[TriggerDeploy] starting %s", h.deployScript)
	go func() {
		out, err := exec.Command("bash", h.deployScript).CombinedOutput()
		if err != nil {
			log.Printf("[TriggerDeploy] error: %v\n%s", err, out)
		} else {
			log.Printf("[TriggerDeploy] done\n%s", out)
		}
	}()
	w.WriteHeader(http.StatusAccepted)
}
