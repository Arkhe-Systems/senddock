package middleware

import (
	"net/http"
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

		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		}

		key := "rl:" + ip
		count, err := rl.redis.Increment(r.Context(), key, rl.window)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if count > rl.limit {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}
