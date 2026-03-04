package aggregator

import (
	"edge-app/internal/core"
	"edge-app/internal/storage/ringbuffer"
)

type RawCapture struct {
	Bus    *core.Bus
	Buffer *ringbuffer.Buffer
}

func (r *RawCapture) Run() {
	in := r.Bus.SubscribeMetrics(2048)
	for m := range in {
		r.Buffer.Add(m)
	}
}
