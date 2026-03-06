package api

import (
	"crypto/rsa"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
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
	DB          *sql.DB
	Status      *StatusServer
	Raw         *RawServer
	RuleEngine  *rules.Engine
	KeycloakURL string
	Realm       string
	pubKeys     []*rsa.PublicKey
	keysMu      sync.RWMutex
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type jwks struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kty string `json:"kty"`
	N   string `json:"n"`
	E   string `json:"e"`
	Use string `json:"use"`
}

func (s *Server) fetchKeys() {
	url := s.KeycloakURL + "/realms/" + s.Realm + "/protocol/openid-connect/certs"
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var keys jwks
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return
	}
	var pubKeys []*rsa.PublicKey
	for _, k := range keys.Keys {
		if k.Kty != "RSA" || k.Use != "sig" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		n := new(big.Int).SetBytes(nBytes)
		e := 0
		for _, b := range eBytes {
			e = e<<8 + int(b)
		}
		pubKeys = append(pubKeys, &rsa.PublicKey{N: n, E: e})
	}
	s.keysMu.Lock()
	s.pubKeys = pubKeys
	s.keysMu.Unlock()
}

func (s *Server) parseToken(tokenStr string) ([]string, bool) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, false
	}
	var claims struct {
		Exp   int64    `json:"exp"`
		Iss   string   `json:"iss"`
		Roles []string `json:"roles"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, false
	}
	if time.Now().Unix() > claims.Exp {
		return nil, false
	}
	expected1 := s.KeycloakURL + "/realms/" + s.Realm
	expected2 := "http://localhost:8180/realms/" + s.Realm
	if claims.Iss != expected1 && claims.Iss != expected2 {
		return nil, false
	}
	return claims.Roles, true
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next(w, r)
			return
		}
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		roles, valid := s.parseToken(token)
		if !valid {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}
		r.Header.Set("X-Roles", strings.Join(roles, ","))
		next(w, r)
	}
}

func (s *Server) requireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		roles := strings.Split(r.Header.Get("X-Roles"), ",")
		for _, ro := range roles {
			if ro == role {
				next(w, r)
				return
			}
		}
		http.Error(w, `{"error":"forbidden","required":"` + role + `"}`, http.StatusForbidden)
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

func (s *Server) handleRulesGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.RuleEngine.GetRules())
}

func (s *Server) handleRulesPost(w http.ResponseWriter, r *http.Request) {
	var rule rules.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	s.RuleEngine.AddRule(rule)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

func (s *Server) handleRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleRulesGet(w, r)
	case http.MethodPost:
		roles := strings.Split(r.Header.Get("X-Roles"), ",")
		isAdmin := false
		for _, ro := range roles {
			if ro == "admin" {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			http.Error(w, `{"error":"forbidden","message":"admin role required to create rules"}`, http.StatusForbidden)
			return
		}
		s.handleRulesPost(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) Run(addr string) error {
	go func() {
		for {
			s.fetchKeys()
			time.Sleep(5 * time.Minute)
		}
	}()
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/v1/aggregates", s.requireAuth(s.aggregates))
	mux.HandleFunc("/api/v1/status", s.requireAuth(s.Status.Handler))
	mux.HandleFunc("/api/v1/raw", s.requireAuth(s.Raw.Handler))
	mux.HandleFunc("/api/v1/rules", s.requireAuth(s.handleRules))
	mux.Handle("/metrics", metrics.Handler())
	srv := &http.Server{
		Addr:              addr,
		Handler:           cors(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return srv.ListenAndServe()
}
