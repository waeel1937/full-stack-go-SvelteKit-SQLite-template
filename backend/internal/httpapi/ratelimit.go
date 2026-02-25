package httpapi

import (
	"net/http"
	"sync"
	"time"
)

type limiter struct {
	mu     sync.Mutex
	tokens int
	last   time.Time
}

func RateLimit(rps int) func(http.Handler) http.Handler {
	l := &limiter{tokens: rps, last: time.Now()}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l.mu.Lock()
			now := time.Now()
			elapsed := now.Sub(l.last).Seconds()
			l.tokens += int(elapsed * float64(rps))
			if l.tokens > rps {
				l.tokens = rps
			}
			l.last = now

			if l.tokens <= 0 {
				l.mu.Unlock()
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			l.tokens--
			l.mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
