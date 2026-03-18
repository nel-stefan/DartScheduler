// Package logbuf provides a thread-safe ring buffer for log lines that
// implements io.Writer so it can be wired into log.SetOutput.
package logbuf

import (
	"strings"
	"sync"
)

const defaultMax = 200

// Buffer is a thread-safe ring buffer of log lines.
type Buffer struct {
	mu    sync.Mutex
	lines []string
	max   int
}

// New returns a Buffer that keeps at most max lines.
func New(max int) *Buffer {
	if max <= 0 {
		max = defaultMax
	}
	return &Buffer{max: max}
}

// Write implements io.Writer. Each call may contain one or more
// newline-terminated lines.
func (b *Buffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, line := range strings.Split(strings.TrimRight(string(p), "\n"), "\n") {
		if line == "" {
			continue
		}
		b.lines = append(b.lines, line)
		if len(b.lines) > b.max {
			b.lines = b.lines[len(b.lines)-b.max:]
		}
	}
	return len(p), nil
}

// Lines returns a copy of all buffered lines, oldest first.
func (b *Buffer) Lines() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]string, len(b.lines))
	copy(out, b.lines)
	return out
}
