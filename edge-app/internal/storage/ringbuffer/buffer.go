package ringbuffer

import (
	"sync"

	"edge-app/internal/core"
)

type Buffer struct {
	mu     sync.Mutex
	size   int
	data   []core.MetricEvent
	cursor int
	full   bool
}

func New(size int) *Buffer {
	return &Buffer{
		size: size,
		data: make([]core.MetricEvent, size),
	}
}

func (b *Buffer) Add(e core.MetricEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.data[b.cursor] = e
	b.cursor = (b.cursor + 1) % b.size
	if b.cursor == 0 {
		b.full = true
	}
}

func (b *Buffer) Snapshot() []core.MetricEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	var out []core.MetricEvent
	if b.full {
		out = append(out, b.data[b.cursor:]...)
		out = append(out, b.data[:b.cursor]...)
	} else {
		out = append(out, b.data[:b.cursor]...)
	}
	return out
}
