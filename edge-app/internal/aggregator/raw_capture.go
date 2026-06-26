package aggregator

import (
	"context"

	"edge-app/internal/core"
	"edge-app/internal/storage/ringbuffer"
)

type RawCapture struct {
	Bus    *core.Bus
	Buffer *ringbuffer.Buffer
}

func (r *RawCapture) Run(ctx context.Context) {
	in := r.Bus.SubscribeMetrics(2048)
	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-in:
			if !ok {
				return
			}
			r.Buffer.Add(m)
		}
	}
}
