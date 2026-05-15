package middleware

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

// responseWriter wraps http.ResponseWriter to capture the status code and
// the total number of bytes written by downstream handlers. Middlewares use
// this to observe response properties without changing handler code.
type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
	written     int // total bytes written
}

// WriteHeader records the response status and forwards the call. Subsequent
// calls are ignored to match http.ResponseWriter semantics.
func (rw *responseWriter) WriteHeader(statusCode int) {
	if rw.wroteHeader {
		return
	}
	rw.statusCode = statusCode
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write ensures a header is present, writes the payload and accumulates the
// written byte count.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.written += n
	return n, err
}

// newResponseWriter returns a responseWriter pre-set with HTTP 200 so that
// handlers that never call WriteHeader are still observable.
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// inflightRequests is an atomic counter that tracks the number of currently
// processing requests. Using sync/atomic avoids locks and is efficient for
// a simple global counter.
var inflightRequests atomic.Int64

func init() {
	// Register a gauge that reports the current in-flight request count. The
	// gauge callback is queried by the metrics scraper; registering here keeps
	// metric setup centralized and idempotent.
	metrics.GetOrCreateGauge("http_inflight_requests", func() float64 {
		return float64(inflightRequests.Load())
	})
}

// WithMetrics returns middleware that instruments incoming HTTP requests with
// common operational metrics. It records:
//
// - http_inflight_requests (gauge): number of active in-flight requests
// - http_request_duration_seconds{status, url} (histogram): request latency
// - http_response_size_bytes{status, url} (histogram): response size in bytes
// - http_requests_total{method, status, url} (counter): total requests
// - http_errors_total{status, url} (counter): server error (5xx) occurrences
//
// Important notes:
//   - Metrics are labeled by URL path and status. To prevent high-cardinality
//     metric explosions, consider normalizing paths (e.g., removing numeric IDs)
//     before using them as labels.
//   - The inflight gauge is incremented at request start and safely decremented
//     with a defer so it remains correct if handlers panic or return early.
//   - This middleware uses VictoriaMetrics' string-keyed metric API; avoid
//     generating unbounded unique metric labels at runtime.
func WithMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		inflightRequests.Add(1)
		defer inflightRequests.Add(-1)

		rw := newResponseWriter(w)

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()

		// Record duration and response size with status+path labels.
		durMetric := fmt.Sprintf(`http_request_duration_seconds{status="%d",url=%q}`, rw.statusCode, r.URL.Path)
		metrics.GetOrCreateHistogram(durMetric).Update(duration)

		szMetric := fmt.Sprintf(`http_response_size_bytes{status="%d",url=%q}`, rw.statusCode, r.URL.Path)
		metrics.GetOrCreateHistogram(szMetric).Update(float64(rw.written))

		// Increment request count (method, status, path).
		reqMetric := fmt.Sprintf(`http_requests_total{method=%q,status="%d",url=%q}`,
			r.Method, rw.statusCode, r.URL.Path)
		metrics.GetOrCreateCounter(reqMetric).Inc()

		// Track server errors (5xx) separately.
		if rw.statusCode >= 500 {
			errMetric := fmt.Sprintf(`http_errors_total{status="%d",url=%q}`, rw.statusCode, r.URL.Path)
			metrics.GetOrCreateCounter(errMetric).Inc()
		}
	})
}
