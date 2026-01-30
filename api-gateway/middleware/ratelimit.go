package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client   *redis.Client
	limit    int
	duration time.Duration
}

func NewRateLimiter(client *redis.Client, limit int, duration time.Duration) *RateLimiter {
	return &RateLimiter{
		client:   client,
		limit:    limit,
		duration: duration,
	}
}

func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		ip := r.Header.Get("X-Real-IP")
		if ip == "" {
			ip = r.RemoteAddr
		}

		key := fmt.Sprintf("ratelimit:%s", ip)
		
		// Get current count
		count, err := rl.client.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			// If Redis is down, allow the request
			next.ServeHTTP(w, r)
			return
		}

		if count >= rl.limit {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded","message":"too many requests"}`))
			return
		}

		// Increment counter
		pipe := rl.client.Pipeline()
		pipe.Incr(ctx, key)
		if count == 0 {
			pipe.Expire(ctx, key, rl.duration)
		}
		_, err = pipe.Exec(ctx)
		if err != nil {
			// If Redis operation fails, allow the request
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
