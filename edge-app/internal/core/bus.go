package core

import (
	"log"
	"sync"
)

type Bus struct {
	mu         sync.RWMutex
	metricSubs []chan MetricEvent
	aggSubs    []chan AggregateEvent
	Alerts     chan RuleAlert
}

type RuleAlert struct {
	Key     string
	Message string
	Value   float64
}

func NewBus() *Bus {
	return &Bus{
		Alerts: make(chan RuleAlert, 256),
	}
}

func (b *Bus) SubscribeMetrics(bufSize int) <-chan MetricEvent {
	ch := make(chan MetricEvent, bufSize)
	b.mu.Lock()
	b.metricSubs = append(b.metricSubs, ch)
	b.mu.Unlock()
	return ch
}

func (b *Bus) PublishMetric(e MetricEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.metricSubs {
		select {
		case ch <- e:
		default:
			log.Printf("WARN: metric subscriber lagging key=%s", e.Key)
		}
	}
}

func (b *Bus) SubscribeAggregates(bufSize int) <-chan AggregateEvent {
	ch := make(chan AggregateEvent, bufSize)
	b.mu.Lock()
	b.aggSubs = append(b.aggSubs, ch)
	b.mu.Unlock()
	return ch
}

func (b *Bus) PublishAggregate(e AggregateEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.aggSubs {
		select {
		case ch <- e:
		default:
			log.Printf("WARN: aggregate subscriber lagging key=%s", e.Key)
		}
	}
}
