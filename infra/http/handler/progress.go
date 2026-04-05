package handler

import (
	"net/http"
	"sync"
)

// ProgressTracker tracks the current annealing progress across a single HTTP request.
// Only one schedule generation can run at a time.
type ProgressTracker struct {
	mu    sync.RWMutex
	step  int
	total int
	busy  bool
}

// GlobalProgress is the application-wide annealing progress tracker.
var GlobalProgress = &ProgressTracker{}

// Start marks the tracker as busy and resets progress to 0/total.
func (pt *ProgressTracker) Start(total int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.step = 0
	pt.total = total
	pt.busy = true
}

// Update sets the current step.
func (pt *ProgressTracker) Update(step int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.step = step
}

// SetTotal updates the total step count. Called when the scheduler determines
// the actual step count (which may differ from the initial estimate due to scaling).
func (pt *ProgressTracker) SetTotal(total int) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	if total > 0 {
		pt.total = total
	}
}

// Done marks the tracker as no longer busy.
func (pt *ProgressTracker) Done() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.step = pt.total
	pt.busy = false
}

// ProgressHandler exposes annealing progress via HTTP.
type ProgressHandler struct{}

// NewProgressHandler creates a new ProgressHandler.
func NewProgressHandler() *ProgressHandler { return &ProgressHandler{} }

type progressResponse struct {
	Step    int  `json:"step"`
	Total   int  `json:"total"`
	Percent int  `json:"percent"`
	Busy    bool `json:"busy"`
}

// GetProgress returns the current annealing progress as JSON.
func (h *ProgressHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	GlobalProgress.mu.RLock()
	step := GlobalProgress.step
	total := GlobalProgress.total
	busy := GlobalProgress.busy
	GlobalProgress.mu.RUnlock()

	percent := 0
	if total > 0 {
		percent = step * 100 / total
	}
	writeJSON(w, progressResponse{
		Step:    step,
		Total:   total,
		Percent: percent,
		Busy:    busy,
	})
}
