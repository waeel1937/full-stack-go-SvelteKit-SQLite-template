package aggregator

import (
	"context"
	"log"
	"time"

	"edge-app/internal/core"
	"edge-app/internal/metrics"
)

type Aggregator struct {
	Window time.Duration
	Bus    *core.Bus
}

func (a *Aggregator) Run(ctx context.Context) {
	log.Println("aggregator started")
	in := a.Bus.SubscribeMetrics(2048)
	ticker := time.NewTicker(a.Window)
	defer ticker.Stop()

	type acc struct {
		sum   float64
		min   float64
		max   float64
		count int
	}

	state := make(map[string]*acc)

	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-in:
			if !ok {
				return
			}
			metrics.MetricsIngested.Inc()
			v, exists := state[m.Key]
			if !exists {
				state[m.Key] = &acc{sum: m.Value, min: m.Value, max: m.Value, count: 1}
			} else {
				v.sum += m.Value
				if m.Value < v.min {
					v.min = m.Value
				}
				if m.Value > v.max {
					v.max = m.Value
				}
				v.count++
			}
		case t := <-ticker.C:
			for k, v := range state {
				a.Bus.PublishAggregate(core.AggregateEvent{
					Time:   t,
					Window: a.Window,
					Key:    k,
					Avg:    v.sum / float64(v.count),
					Min:    v.min,
					Max:    v.max,
					Count:  v.count,
				})
				metrics.AggregatesEmitted.Inc()
			}
			// Clear in-place to avoid GC pressure from reallocating the map every tick
			for k := range state {
				delete(state, k)
			}
		}
	}
}
