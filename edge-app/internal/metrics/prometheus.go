package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	MetricsIngested = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "edge_metrics_ingested_total",
			Help: "Total number of raw metrics ingested",
		},
	)

	AggregatesEmitted = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "edge_aggregates_emitted_total",
			Help: "Total number of aggregates emitted",
		},
	)
)

func Init() {
	prometheus.MustRegister(MetricsIngested)
	prometheus.MustRegister(AggregatesEmitted)
}

func Handler() http.Handler {
	return promhttp.Handler()
}
