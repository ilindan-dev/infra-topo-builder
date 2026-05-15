package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// WithLogger returns middleware that logs HTTP requests after they have been
// processed. It records method, path, response status, duration and remote
// address. The logger passed in should be pre-configured with desired
// handlers/levels; this middleware only emits Info-level events for
// successful request handling.
func WithLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			log.Info("http request handled",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.statusCode),
				slog.Duration("duration", time.Since(start)),
				slog.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}
