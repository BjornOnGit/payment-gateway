package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	Redis *redis.Client
	// allowed requests per window
	Limit int
	// window duration
	Window time.Duration
	// header containing client identifier
	ClientIDHeader string
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	lim := 60
	if v := os.Getenv("RATE_LIMIT_PER_MIN"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			lim = i
		}
	}
	return &RateLimiter{
		Redis:          rdb,
		Limit:          lim,
		Window:         time.Minute,
		ClientIDHeader: "X-Client-Id",
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientID := r.Header.Get(rl.ClientIDHeader)
		if clientID == "" {
			// fallback to IP
			clientID = r.RemoteAddr
		}
		ctx := r.Context()
		key := fmt.Sprintf("rate:%s:%d", clientID, time.Now().Unix()/int64(rl.Window.Seconds()))

		// INCR and set TTL if first
		val, err := rl.Redis.Incr(ctx, key).Result()
		if err != nil {
			// on Redis error, prefer fail-open (allow) but log server side
			// you might want to fail-closed in a stricter environment
			next.ServeHTTP(w, r)
			return
		}
		if val == 1 {
			// first hit in window -> set ttl to align with window expiry
			_ = rl.Redis.Expire(ctx, key, rl.Window).Err()
		}

		if int(val) > rl.Limit {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate limit exceeded"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}
