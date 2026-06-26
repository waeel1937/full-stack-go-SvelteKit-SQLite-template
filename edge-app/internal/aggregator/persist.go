package aggregator

import (
	"context"
	"log"

	"edge-app/internal/core"
	"edge-app/internal/storage"
)

type Persister struct {
	Bus   *core.Bus
	Store *storage.Store
}

func (p *Persister) Run(ctx context.Context) {
	in := p.Bus.SubscribeAggregates(512)
	for {
		select {
		case <-ctx.Done():
			return
		case a, ok := <-in:
			if !ok {
				return
			}
			if err := p.Store.InsertAggregate(a.Time, a.Window, a.Key, a.Avg, a.Min, a.Max, a.Count); err != nil {
				log.Println("persist error:", err)
			}
		}
	}
}
