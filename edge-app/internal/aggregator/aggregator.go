package aggregator

import (
	"time"

	"edge-app/internal/core"
	"edge-app/internal/logging"
	"edge-app/internal/metrics"
)

type Aggregator struct {
	Window time.Duration
	Bus    *core.Bus
}

func (a *Aggregator) Run() {
	logging.Logger.Println("aggregator started")
	in := a.Bus.SubscribeMetrics(2048)
	ticker := time.NewTicker(a.Window)
	defer ticker.Stop()

	type acc struct {
		sum   float64
		min   float64
		max   float64
		count int
	}

	state := map[string]*acc{}

	for {
		select {
		case m := <-in:
			metrics.MetricsIngested.Inc()
			v, ok := state[m.Key]
			if !ok {
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
			state = map[string]*acc{}
		}
	}
}
