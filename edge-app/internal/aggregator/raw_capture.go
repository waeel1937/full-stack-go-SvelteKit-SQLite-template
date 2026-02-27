package aggregator

import (
	"edge-app/internal/core"
	"edge-app/internal/storage/ringbuffer"
)

type RawCapture struct {
	In     <-chan core.MetricEvent
	Buffer *ringbuffer.Buffer
}

func (r *RawCapture) Run() {
	for m := range r.In {
		r.Buffer.Add(m)
	}
}
