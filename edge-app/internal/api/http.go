package api

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"edge-app/internal/core"
	"edge-app/internal/metrics"
	"edge-app/internal/rules"
)

// Aggregate is the JSON shape returned by /api/v1/aggregates and the SSE stream.
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
	Bus         *core.Bus
	Status      *StatusServer
	Raw         *RawServer
	RuleEngine  *rules.Engine
	KeycloakURL string
	Realm       string
	CORSOrigin  string // defaults to "*" if empty; set to exact frontend origin in production
	pubKeys     []*rsa.PublicKey
	keysMu      sync.RWMutex
}

// ── CORS ────────────────────────────────────────────────────────────────────

func (s *Server) corsOrigin() string {
	if s.CORSOrigin != "" {
		return s.CORSOrigin
	}
	return "*"
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", s.corsOrigin())
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ── JWT / Keycloak ──────────────────────────────────────────────────────────

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
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
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

func (s *Server) startKeyRefresh(ctx context.Context) {
	go func() {
		s.fetchKeys()
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.fetchKeys()
			}
		}
	}()
}

// extractBearer accepts the token from the Authorization header or the
// ?token= query param (needed for EventSource which cannot set headers).
func (s *Server) extractBearer(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return r.URL.Query().Get("token")
}

// parseToken verifies the JWT signature with RSA-SHA256 against the cached
// Keycloak public keys, then validates expiry and issuer.
func (s *Server) parseToken(tokenStr string) (roles []string, ok bool) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, false
	}

	// Cryptographic signature check
	sigInput := parts[0] + "." + parts[1]
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, false
	}

	s.keysMu.RLock()
	keys := s.pubKeys
	s.keysMu.RUnlock()

	if len(keys) == 0 {
		return nil, false
	}

	h := sha256.Sum256([]byte(sigInput))
	verified := false
	for _, key := range keys {
		if rsa.VerifyPKCS1v15(key, crypto.SHA256, h[:], sig) == nil {
			verified = true
			break
		}
	}
	if !verified {
		return nil, false
	}

	// Claims validation
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, false
	}
	var claims struct {
		Exp         int64  `json:"exp"`
		Iss         string `json:"iss"`
		RealmAccess struct {
			Roles []string `json:"roles"`
		} `json:"realm_access"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, false
	}
	if time.Now().Unix() > claims.Exp {
		return nil, false
	}
	if claims.Iss != s.KeycloakURL+"/realms/"+s.Realm {
		return nil, false
	}
	return claims.RealmAccess.Roles, true
}

// ── Auth middleware ──────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(body))
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next(w, r)
			return
		}
		token := s.extractBearer(r)
		if token == "" {
			writeJSON(w, http.StatusUnauthorized, `{"error":"unauthorized"}`)
			return
		}
		rolesSlice, valid := s.parseToken(token)
		if !valid {
			writeJSON(w, http.StatusUnauthorized, `{"error":"invalid token"}`)
			return
		}
		r.Header.Set("X-Roles", strings.Join(rolesSlice, ","))
		next(w, r)
	}
}

// hasRole checks whether X-Roles (set by requireAuth) contains role.
// Must only be called inside a requireAuth-wrapped handler.
func hasRole(r *http.Request, role string) bool {
	for _, ro := range strings.Split(r.Header.Get("X-Roles"), ",") {
		if ro == role {
			return true
		}
	}
	return false
}

// ── Handlers ─────────────────────────────────────────────────────────────────

func (s *Server) aggregates(w http.ResponseWriter, r *http.Request) {
	windowMs, _ := strconv.ParseInt(r.URL.Query().Get("window_ms"), 10, 64)
	if windowMs == 0 {
		windowMs = 1000
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := s.DB.QueryContext(r.Context(),
		`SELECT time, window, metric, avg, min, max, count
		 FROM aggregates WHERE window = ? ORDER BY time DESC LIMIT ?`,
		windowMs, limit,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	out := make([]Aggregate, 0, limit)
	for rows.Next() {
		var a Aggregate
		if err := rows.Scan(&a.Time, &a.Window, &a.Metric, &a.Avg, &a.Min, &a.Max, &a.Count); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		out = append(out, a)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (s *Server) handleRulesGet(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.RuleEngine.GetRules())
}

func (s *Server) handleRulesPost(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var rule rules.Rule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeJSON(w, http.StatusBadRequest, `{"error":"invalid body"}`)
		return
	}
	if rule.ID == "" || rule.Key == "" {
		writeJSON(w, http.StatusBadRequest, `{"error":"id and key are required"}`)
		return
	}
	if err := s.RuleEngine.AddRule(rule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rule)
}

func (s *Server) handleRules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleRulesGet(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// ── SSE stream ───────────────────────────────────────────────────────────────

// streamEvents pushes metric (latest-per-key, every 500 ms), aggregate, and
// alert events over a Server-Sent Events connection.
func (s *Server) streamEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Disable the per-connection write deadline for long-lived SSE connections.
	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Time{})

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering

	metricCh := s.Bus.SubscribeMetrics(512)
	aggCh := s.Bus.SubscribeAggregates(128)
	alertCh := s.Bus.SubscribeAlerts(64)

	// Throttle raw metrics: collect latest value per key, flush every 500 ms.
	latestMetrics := make(map[string]core.MetricEvent)
	metricFlush := time.NewTicker(500 * time.Millisecond)
	defer metricFlush.Stop()

	// Keep the connection alive through proxies.
	heartbeat := time.NewTicker(25 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case m := <-metricCh:
			latestMetrics[m.Key] = m

		case <-metricFlush.C:
			for _, m := range latestMetrics {
				b, _ := json.Marshal(m)
				fmt.Fprintf(w, "event: metric\ndata: %s\n\n", b)
			}
			if len(latestMetrics) > 0 {
				flusher.Flush()
				latestMetrics = make(map[string]core.MetricEvent)
			}

		case a := <-aggCh:
			agg := Aggregate{
				Time:   a.Time.Unix(),
				Window: int64(a.Window / time.Millisecond),
				Metric: a.Key,
				Avg:    a.Avg,
				Min:    a.Min,
				Max:    a.Max,
				Count:  a.Count,
			}
			b, _ := json.Marshal(agg)
			fmt.Fprintf(w, "event: aggregate\ndata: %s\n\n", b)
			flusher.Flush()

		case al := <-alertCh:
			b, _ := json.Marshal(al)
			fmt.Fprintf(w, "event: alert\ndata: %s\n\n", b)
			flusher.Flush()

		case <-heartbeat.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

// ── Server lifecycle ─────────────────────────────────────────────────────────

// Build starts the Keycloak key-refresh goroutine and returns a configured
// *http.Server ready for ListenAndServe. The caller is responsible for
// graceful shutdown.
func (s *Server) Build(ctx context.Context, addr string) *http.Server {
	s.startKeyRefresh(ctx)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.Handle("/metrics", metrics.Handler())
	mux.HandleFunc("/openapi.json", s.handleOpenAPI)
	mux.HandleFunc("/docs", s.handleDocs)

	mux.HandleFunc("/api/v1/aggregates", s.requireAuth(s.aggregates))
	mux.HandleFunc("/api/v1/status", s.requireAuth(s.Status.Handler))
	mux.HandleFunc("/api/v1/raw", s.requireAuth(s.Raw.Handler))
	mux.HandleFunc("/api/v1/stream", s.requireAuth(s.streamEvents))

	// Rules: GET requires auth, POST requires admin role (X-Roles set by requireAuth above)
	mux.HandleFunc("/api/v1/rules", s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if !hasRole(r, "admin") {
				writeJSON(w, http.StatusForbidden, `{"error":"forbidden","required":"admin"}`)
				return
			}
			s.handleRulesPost(w, r)
			return
		}
		s.handleRules(w, r)
	}))

	return &http.Server{
		Addr:              addr,
		Handler:           s.corsMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
