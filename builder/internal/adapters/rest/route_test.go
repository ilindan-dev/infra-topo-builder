package rest_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/adapters/rest"
	"github.com/ilindan-dev/infra-topo-builder/builder/internal/config"
)

// TestSetupRoutes ensures SetupRoutes registers expected endpoints and that middleware
// (recovery, rate limiting, metrics) produce expected HTTP statuses for representative
// requests to each route.
func TestSetupRoutes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := &config.Config{RPS: 100}
	apiHandler := rest.NewAPIHandler(nil, nil, logger)
	handler := rest.SetupRoutes(apiHandler, cfg, logger)

	tests := []struct {
		name, method, path string
		expectedStatus     int
	}{
		{
			name:           "Ping Healthcheck",
			method:         http.MethodGet,
			path:           "/ping",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Metrics Endpoint",
			method:         http.MethodGet,
			path:           "/metrics",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Parse Route (Bad Request due to empty body)",
			method:         http.MethodPost,
			path:           "/api/v1/parse/",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Topology Route (Panic caught by Recovery)",
			method:         http.MethodGet,
			path:           "/api/v1/topology/123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Node Route (Panic caught by Recovery)",
			method:         http.MethodGet,
			path:           "/api/v1/node/123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Port Route (Panic caught by Recovery)",
			method:         http.MethodGet,
			path:           "/api/v1/port/123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Log Route (Panic caught by Recovery)",
			method:         http.MethodGet,
			path:           "/api/v1/log/123e4567-e89b-12d3-a456-426614174000",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Unknown Route",
			method:         http.MethodGet,
			path:           "/api/v1/unknown-endpoint",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
