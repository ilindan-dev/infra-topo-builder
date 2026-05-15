package middleware

import (
	"net/http"

	"golang.org/x/time/rate"
)

// Rate returns a middleware that enforces a global request rate limit.
// The limiter is shared across all requests handled by the process. rps
// configures the allowed requests per second and also sets the burst size to
// the same value. This is a simple process-global limiter; for per-client
// limits implement a key-based limiter (for example keyed by IP).
func Rate(rps int) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(rps), rps) // Burst = RPS
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
