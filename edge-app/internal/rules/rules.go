package rules

import (
	"context"
	"log"
	"sync"
	"time"

	"edge-app/internal/core"
	"edge-app/internal/storage"
)

type Rule struct {
	ID        string  `json:"id"`
	Key       string  `json:"key"`
	Condition string  `json:"condition"`
	Threshold float64 `json:"threshold"`
	Message   string  `json:"message"`
	Enabled   bool    `json:"enabled"`
}

type Engine struct {
	Bus   *core.Bus
	Store *storage.Store
	mu    sync.RWMutex
	rules []Rule
}

var defaultRules = []Rule{
	{ID: "temp-high", Key: "temperature", Condition: "avg_gt", Threshold: 80, Message: "Temperature above 80°C", Enabled: true},
	{ID: "temp-critical", Key: "temperature", Condition: "max_gt", Threshold: 100, Message: "Temperature critical (>100°C)", Enabled: true},
	{ID: "pressure-low", Key: "pressure", Condition: "min_lt", Threshold: 10, Message: "Pressure below minimum (10 bar)", Enabled: true},
}

func NewEngine(bus *core.Bus, store *storage.Store) (*Engine, error) {
	e := &Engine{Bus: bus, Store: store}

	rows, err := store.LoadRules()
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		for _, r := range defaultRules {
			if err := store.UpsertRule(r.ID, r.Key, r.Condition, r.Threshold, r.Message, r.Enabled); err != nil {
				return nil, err
			}
		}
		e.rules = append(e.rules, defaultRules...)
	} else {
		for _, r := range rows {
			e.rules = append(e.rules, Rule{
				ID:        r.ID,
				Key:       r.Key,
				Condition: r.Condition,
				Threshold: r.Threshold,
				Message:   r.Message,
				Enabled:   r.Enabled,
			})
		}
	}

	return e, nil
}

func (e *Engine) GetRules() []Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]Rule, len(e.rules))
	copy(out, e.rules)
	return out
}

func (e *Engine) AddRule(r Rule) error {
	if err := e.Store.UpsertRule(r.ID, r.Key, r.Condition, r.Threshold, r.Message, r.Enabled); err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, existing := range e.rules {
		if existing.ID == r.ID {
			e.rules[i] = r
			return nil
		}
	}
	e.rules = append(e.rules, r)
	return nil
}

func (e *Engine) Run(ctx context.Context) {
	in := e.Bus.SubscribeAggregates(512)
	log.Println("rule engine started")
	for {
		select {
		case <-ctx.Done():
			return
		case a, ok := <-in:
			if !ok {
				return
			}
			e.evaluate(a)
		}
	}
}

func (e *Engine) evaluate(a core.AggregateEvent) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, r := range e.rules {
		if !r.Enabled || r.Key != a.Key {
			continue
		}
		var triggered bool
		var val float64
		switch r.Condition {
		case "avg_gt":
			triggered, val = a.Avg > r.Threshold, a.Avg
		case "max_gt":
			triggered, val = a.Max > r.Threshold, a.Max
		case "min_lt":
			triggered, val = a.Min < r.Threshold, a.Min
		}
		if triggered {
			log.Printf("RULE TRIGGERED: %s %s val=%.2f", r.ID, r.Key, val)
			e.Bus.PublishAlert(core.RuleAlert{
				Key:     a.Key,
				Message: r.Message,
				Value:   val,
				Time:    time.Now().UnixMilli(),
			})
		}
	}
}
