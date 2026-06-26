package core

import "time"

type MetricEvent struct {
	Time   time.Time `json:"time"`
	Source string    `json:"source"`
	Key    string    `json:"key"`
	Value  float64   `json:"value"`
	OK     bool      `json:"ok"`
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
