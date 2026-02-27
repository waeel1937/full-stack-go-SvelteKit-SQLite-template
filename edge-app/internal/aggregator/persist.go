package aggregator

import (
	"log"

	"edge-app/internal/core"
	"edge-app/internal/storage"
)

type Persister struct {
	In    <-chan core.AggregateEvent
	Store *storage.Store
}

func (p *Persister) Run() {
	for a := range p.In {
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
