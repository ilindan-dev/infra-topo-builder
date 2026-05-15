package rest

import (
	"log/slog"
	"net/http"

	"github.com/VictoriaMetrics/metrics"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/adapters/rest/middleware"
	"github.com/ilindan-dev/infra-topo-builder/builder/internal/config"
)

// SetupRoutes registers HTTP endpoints and composes the global middleware
// chain applied to all routes. The function returns an http.Handler that can
// be passed to an http.Server. Middleware order is intentional:
//
//  1. Recovery - outermost; recovers panics from any downstream layer
//  2. WithLogger - records requests and responses for observability
//  3. WithMetrics - collects latency/size/count/error metrics
//  4. Rate - process-global request throttling
//
// The registered endpoints are documented in package doc.go. The mux uses
// simple patterns; if path parameter parsing is required by the router, it
// should be provided by caller or replaced with a router that supports
// parameter extraction.
func SetupRoutes(h *APIHandler, cfg *config.Config, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/parse/", h.ParseLog)
	mux.HandleFunc("GET /api/v1/topology/{log_id}", h.GetTopology)
	mux.HandleFunc("GET /api/v1/node/{node_id}", h.GetNode)
	mux.HandleFunc("GET /api/v1/port/{node_id}", h.GetPorts)
	mux.HandleFunc("GET /api/v1/log/{log_id}", h.GetLogInfo)
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, _ *http.Request) {
		metrics.WritePrometheus(w, true)
	})
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	handler := middleware.Recovery(logger)(mux)
	handler = middleware.WithLogger(logger)(handler)
	handler = middleware.WithMetrics(handler)
	handler = middleware.Rate(cfg.RPS)(handler)

	return handler
}
