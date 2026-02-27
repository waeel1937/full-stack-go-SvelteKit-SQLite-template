package core

type Bus struct {
	Metrics    chan MetricEvent
	Aggregates chan AggregateEvent
}

func NewBus() *Bus {
	return &Bus{
		Metrics:    make(chan MetricEvent, 1024),
		Aggregates: make(chan AggregateEvent, 1024),
	}
}
