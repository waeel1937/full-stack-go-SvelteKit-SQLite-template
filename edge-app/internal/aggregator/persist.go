package aggregator

import (
	"log"

	"edge-app/internal/core"
	"edge-app/internal/storage"
)

type Persister struct {
	Bus   *core.Bus
	Store *storage.Store
}

func (p *Persister) Run() {
	in := p.Bus.SubscribeAggregates(512)
	for a := range in {
		if err := p.Store.InsertAggregate(
			a.Time,
			a.Window,
			a.Key,
			a.Avg,
			a.Min,
			a.Max,
			a.Count,
		); err != nil {
			log.Println("persist error:", err)
		}
	}
}
