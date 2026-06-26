package core

import (
	"log"
	"sync"
)

type RuleAlert struct {
	Key     string  `json:"key"`
	Message string  `json:"message"`
	Value   float64 `json:"value"`
	Time    int64   `json:"time"`
}

type Bus struct {
	mu         sync.RWMutex
	metricSubs []chan MetricEvent
	aggSubs    []chan AggregateEvent
	alertSubs  []chan RuleAlert
}

func NewBus() *Bus { return &Bus{} }

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

func (b *Bus) SubscribeAlerts(bufSize int) <-chan RuleAlert {
	ch := make(chan RuleAlert, bufSize)
	b.mu.Lock()
	b.alertSubs = append(b.alertSubs, ch)
	b.mu.Unlock()
	return ch
}

func (b *Bus) PublishAlert(a RuleAlert) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.alertSubs {
		select {
		case ch <- a:
		default:
			log.Printf("WARN: alert subscriber lagging key=%s", a.Key)
		}
	}
}
