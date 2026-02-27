package api

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

type Status struct {
	Time        int64   `json:"time"`
	Goroutines  int     `json:"goroutines"`
	MemoryMB    float64 `json:"memory_mb"`
	UptimeSec   int64   `json:"uptime_sec"`
}

type StatusServer struct {
	start time.Time
}

func NewStatusServer() *StatusServer {
	return &StatusServer{start: time.Now()}
}

func (s *StatusServer) Handler(w http.ResponseWriter, r *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	out := Status{
		Time:       time.Now().Unix(),
		Goroutines: runtime.NumGoroutine(),
		MemoryMB:   float64(m.Alloc) / 1024 / 1024,
		UptimeSec:  int64(time.Since(s.start).Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}
