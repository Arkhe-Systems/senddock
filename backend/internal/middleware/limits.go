package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/arkhe-systems/senddock/internal/cache"
)

const maxBodySize = 10 * 1024 * 1024

func LimitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
		next.ServeHTTP(w, r)
	})
}

type RateLimiter struct {
	redis  *cache.Redis
	limit  int64
	window time.Duration
}

func NewRateLimiter(redis *cache.Redis, limit int64, window time.Duration) *RateLimiter {
	return &RateLimiter{redis: redis, limit: limit, window: window}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rl.redis == nil {
			next.ServeHTTP(w, r)
			return
		}

		ip := clientIP(r)

		key := "rl:" + ip
		count, err := rl.redis.Increment(r.Context(), key, rl.window)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if count > rl.limit {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if comma := strings.IndexByte(forwarded, ','); comma >= 0 {
			forwarded = forwarded[:comma]
		}
		if ip := strings.TrimSpace(forwarded); ip != "" {
			return ip
		}
	}
	if real := strings.TrimSpace(r.Header.Get("X-Real-IP")); real != "" {
		return real
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
