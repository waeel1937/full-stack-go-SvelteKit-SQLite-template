package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type Aggregate struct {
	Time   int64   `json:"time"`
	Window int64   `json:"window_ms"`
	Metric string  `json:"metric"`
	Avg    float64 `json:"avg"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	Count  int     `json:"count"`
}

type Server struct {
	DB *sql.DB
}

func (s *Server) aggregates(w http.ResponseWriter, r *http.Request) {
	windowMs, _ := strconv.ParseInt(r.URL.Query().Get("window_ms"), 10, 64)
	if windowMs == 0 {
		windowMs = 1000
	}

	rows, err := s.DB.Query(`
SELECT time, window, metric, avg, min, max, count
FROM aggregates
WHERE window = ?
ORDER BY time DESC
LIMIT 100
`, windowMs)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	out := []Aggregate{}
	for rows.Next() {
		var a Aggregate
		if err := rows.Scan(
			&a.Time,
			&a.Window,
			&a.Metric,
			&a.Avg,
			&a.Min,
			&a.Max,
			&a.Count,
		); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		out = append(out, a)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/aggregates", s.aggregates)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return srv.ListenAndServe()
}
