package middleware

import (
	"log/slog"
	"net/http"
)

// Recovery returns middleware that recovers from panics occurring in
// downstream handlers. When a panic happens it logs the recovered value using
// slog and responds with HTTP 500. The middleware intentionally does not
// re-panic so that the server can continue serving other requests.
func Recovery(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error("panic recovered", slog.Any("error", err))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
