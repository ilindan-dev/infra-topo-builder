// Package middleware provides HTTP middleware used by REST adapters.
//
// This package offers small composable middleware components intended for use
// with net/http handler chains. Each middleware constructor returns a
// function of the form func(http.Handler) http.Handler so callers can build
// pipelines using standard patterns.
//
// Provided middleware
//   - Recovery: recovers panics from downstream handlers and returns HTTP 500.
//   - WithLogger: logs requests including method, path, status and duration.
//   - WithMetrics: records operational metrics (in-flight, latency, size,
//     counts, errors) with configurable labeling caveats.
//   - Rate: a simple token-bucket rate limiter (process-global limiter).
//
// Operational guidance
//   - Compose middleware consistently: Recovery should be outermost to catch
//     panics from other middleware and handlers; WithMetrics/WithLogger should
//     wrap business handlers to observe their behavior.
//   - Rate implements a single process-global limiter. For per-client limits
//     (IP, API key) implement a keyed limiter backed by a small in-memory cache.
//   - WithMetrics attaches labels that include request paths. To avoid
//     high-cardinality metrics, normalize paths (e.g., replace numeric IDs with
//     placeholders) before using them as labels or adapt the middleware to do so.
//
// Extensibility
//   - Add middleware here when it is transport-specific and small. Larger or
//     framework-specific middleware should live closer to the adapter that uses
//     it (for example an HTTP server package).
package middleware
