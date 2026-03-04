package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"edge-app/internal/metrics"
	"edge-app/internal/rules"
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
	DB         *sql.DB
	Status     *StatusServer
	Raw        *RawServer
	RuleEngine *rules.Engine
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,X-API-Key,X-Request-Id")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) aggregates(w http.ResponseWriter, r *http.Request) {
	windowMs, _ := strconv.ParseInt(r.URL.Query().Get("window_ms"), 10, 64)
	if windowMs == 0 {
		windowMs = 1000
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 || limit > 500 {
		limit = 100
	}

	rows, err := s.DB.Query(
		"SELECT time, window, metric, avg, min, max, count FROM aggregates WHERE window = ? ORDER BY time DESC LIMIT ?",
		windowMs, limit,
	)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer rows.Close()

	out := []Aggregate{}
	for rows.Next() {
		var a Aggregate
		if err := rows.Scan(&a.Time, &a.Window, &a.Metric, &a.Avg, &a.Min, &a.Max, &a.Count); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		out = append(out, a)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (s *Server) handleRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.RuleEngine.GetRules())
	case http.MethodPost:
		var rule rules.Rule
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		s.RuleEngine.AddRule(rule)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(rule)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/aggregates", s.aggregates)
	mux.HandleFunc("/api/v1/status", s.Status.Handler)
	mux.HandleFunc("/api/v1/raw", s.Raw.Handler)
	mux.HandleFunc("/api/v1/rules", s.handleRules)
	mux.Handle("/metrics", metrics.Handler())

	srv := &http.Server{
		Addr:              addr,
		Handler:           cors(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return srv.ListenAndServe()
}
