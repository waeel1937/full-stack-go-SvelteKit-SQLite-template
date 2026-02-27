package aggregator

import (
	"time"

	"edge-app/internal/core"
)

type Aggregator struct {
	Window time.Duration
	In     <-chan core.MetricEvent
	Out    chan<- core.AggregateEvent
}

func (a *Aggregator) Run() {
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
		case m := <-a.In:
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
				a.Out <- core.AggregateEvent{
					Time:   t,
					Window: a.Window,
					Key:    k,
					Avg:    v.sum / float64(v.count),
					Min:    v.min,
					Max:    v.max,
					Count:  v.count,
				}
			}
			state = map[string]*acc{}
		}
	}
}
