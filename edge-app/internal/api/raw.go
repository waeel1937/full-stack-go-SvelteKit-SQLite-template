package api

import (
	"encoding/json"
	"net/http"

	"edge-app/internal/core"
	"edge-app/internal/storage/ringbuffer"
)

type RawServer struct {
	Buffer *ringbuffer.Buffer
}

func (r *RawServer) Handler(w http.ResponseWriter, _ *http.Request) {
	data := r.Buffer.Snapshot()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

type RawMetric struct {
	Time   int64   `json:"time"`
	Source string  `json:"source"`
	Key    string  `json:"key"`
	Value  float64 `json:"value"`
	OK     bool    `json:"ok"`
}

func ConvertRaw(in []core.MetricEvent) []RawMetric {
	out := make([]RawMetric, 0, len(in))
	for _, m := range in {
		out = append(out, RawMetric{
			Time:   m.Time.UnixNano(),
			Source: m.Source,
			Key:    m.Key,
			Value:  m.Value,
			OK:     m.OK,
		})
	}
	return out
}
