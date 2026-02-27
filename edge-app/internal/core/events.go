package core

import "time"

type MetricEvent struct {
	Time   time.Time
	Source string
	Key    string
	Value  float64
	OK     bool
}

type AggregateEvent struct {
	Time   time.Time
	Window time.Duration
	Key    string
	Avg    float64
	Min    float64
	Max    float64
	Count  int
}
