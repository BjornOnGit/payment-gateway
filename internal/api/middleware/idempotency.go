package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type IdempotencyStore struct {
	Redis *redis.Client
	TTL   time.Duration
}

// responseCapture captures status and body for middleware to inspect.
type responseCapture struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (r *responseCapture) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseCapture) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}

// NewIdempotencyMiddleware returns middleware that ensures idempotent POSTs.
func (s *IdempotencyStore) NewIdempotencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only apply to POST endpoints (customize as needed)
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get("Idempotency-Key")
		if strings.TrimSpace(key) == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		storeKey := "idempotency:" + key

		// if exists, return stored response quickly
		val, err := s.Redis.Get(ctx, storeKey).Result()
		if err == nil {
			// stored value is JSON of earlier response e.g. {"status":201,"body":{...}}
			var saved struct {
				Status int             `json:"status"`
				Body   json.RawMessage `json:"body"`
			}
			if err := json.Unmarshal([]byte(val), &saved); err == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(saved.Status)
				_, _ = w.Write(saved.Body)
				return
			}
			// if unmarshal fails, fallthrough and process normally
		}

		// Not found: capture the response, then persist if success
		rc := &responseCapture{ResponseWriter: w, status: 200}
		next.ServeHTTP(rc, r)

		// Only cache successful creates (201) - adjust as needed
		if rc.status == http.StatusCreated {
			// store status + body
			payload := struct {
				Status int             `json:"status"`
				Body   json.RawMessage `json:"body"`
			}{
				Status: rc.status,
				Body:   json.RawMessage(rc.body),
			}
			b, _ := json.Marshal(payload)
			_ = s.Redis.Set(ctx, storeKey, b, s.TTL).Err()
		}
	})
}
