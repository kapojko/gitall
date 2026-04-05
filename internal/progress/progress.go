package progress

import (
	"fmt"
	"io"
	"sync"
)

type Progress struct {
	mu      sync.Mutex
	writer  io.Writer
	total   int
	current int
	prefix  string
	enabled bool
}

func New(w io.Writer) *Progress {
	return &Progress{
		writer:  w,
		enabled: true,
	}
}

func (p *Progress) SetPrefix(prefix string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.prefix = prefix
}

func (p *Progress) Enable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = true
}

func (p *Progress) Disable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = false
}

func (p *Progress) SetCurrent(current, total int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current = current
	p.total = total
	p.printLocked()
}

func (p *Progress) Finish() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.enabled && p.writer != nil {
		fmt.Fprintln(p.writer)
	}
}

func (p *Progress) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.enabled && p.writer != nil {
		fmt.Fprint(p.writer, "\r\033[K")
	}
}

func (p *Progress) printLocked() {
	if !p.enabled || p.writer == nil {
		return
	}

	prefix := p.prefix
	if prefix == "" {
		prefix = "Walking directories"
	}

	barWidth := 30
	filled := 0
	if p.total > 0 {
		filled = (barWidth * p.current) / p.total
		if filled > barWidth {
			filled = barWidth
		}
	}

	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "="
		} else if i == filled {
			bar += ">"
		} else {
			bar += "."
		}
	}
	bar += "]"

	fmt.Fprintf(p.writer, "\r%s %s %d/%d", prefix, bar, p.current, p.total)
}
