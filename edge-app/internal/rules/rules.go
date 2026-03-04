package rules

import (
	"log"
	"sync"

	"edge-app/internal/core"
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
	mu    sync.RWMutex
	rules []Rule
}

func NewEngine(bus *core.Bus) *Engine {
	return &Engine{
		Bus: bus,
		rules: []Rule{
			{ID: "temp-high", Key: "temperature", Condition: "avg_gt", Threshold: 80, Message: "Temperature above 80", Enabled: true},
			{ID: "temp-critical", Key: "temperature", Condition: "max_gt", Threshold: 100, Message: "Temperature critical", Enabled: true},
			{ID: "pressure-low", Key: "pressure", Condition: "min_lt", Threshold: 10, Message: "Pressure below minimum", Enabled: true},
		},
	}
}

func (e *Engine) GetRules() []Rule {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]Rule, len(e.rules))
	copy(out, e.rules)
	return out
}

func (e *Engine) AddRule(r Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, existing := range e.rules {
		if existing.ID == r.ID {
			e.rules[i] = r
			return
		}
	}
	e.rules = append(e.rules, r)
}

func (e *Engine) Run() {
	in := e.Bus.SubscribeAggregates(512)
	log.Println("rule engine started")
	for a := range in {
		e.mu.RLock()
		for _, r := range e.rules {
			if !r.Enabled || r.Key != a.Key {
				continue
			}
			triggered := false
			var val float64
			switch r.Condition {
			case "avg_gt":
				triggered = a.Avg > r.Threshold
				val = a.Avg
			case "max_gt":
				triggered = a.Max > r.Threshold
				val = a.Max
			case "min_lt":
				triggered = a.Min < r.Threshold
				val = a.Min
			}
			if triggered {
				log.Printf("RULE TRIGGERED: %s %s val=%.2f", r.ID, r.Key, val)
				select {
				case e.Bus.Alerts <- core.RuleAlert{Key: a.Key, Message: r.Message, Value: val}:
				default:
				}
			}
		}
		e.mu.RUnlock()
	}
}
