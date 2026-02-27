package rules

import (
	"log"

	"edge-app/internal/core"
)

type Engine struct {
	In <-chan core.AggregateEvent
}

func (e *Engine) Run() {
	for a := range e.In {
		if a.Avg > 100 {
			log.Printf("rule triggered: %s avg=%f", a.Key, a.Avg)
		}
	}
}
