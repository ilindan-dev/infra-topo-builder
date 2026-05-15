package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VictoriaMetrics/metrics"
)

// TestResponseWriter verifies that the custom responseWriter records HTTP status codes correctly
// for implicit writes, explicit WriteHeader calls, and ignores subsequent WriteHeader calls.
func TestResponseWriter(t *testing.T) {
	t.Run("implicit 200 OK on Write", func(t *testing.T) {
		rr := httptest.NewRecorder()
		rw := newResponseWriter(rr)

		_, err := rw.Write([]byte("hello"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rw.statusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rw.statusCode)
		}
	})

	t.Run("explicit WriteHeader", func(t *testing.T) {
		rr := httptest.NewRecorder()
		rw := newResponseWriter(rr)

		rw.WriteHeader(http.StatusBadRequest)
		_, err := rw.Write([]byte("bad request"))
		if err != nil {
			t.Errorf("Write error := %v", err)
		}

		if rw.statusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rw.statusCode)
		}
	})

	t.Run("multiple WriteHeader calls ignored", func(t *testing.T) {
		rr := httptest.NewRecorder()
		rw := newResponseWriter(rr)

		rw.WriteHeader(http.StatusCreated)
		rw.WriteHeader(http.StatusInternalServerError)

		if rw.statusCode != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, rw.statusCode)
		}
	})
}

// TestWithMetrics ensures the WithMetrics middleware records request duration and status metrics
// and exposes them via the VictoriaMetrics Prometheus output.
func TestWithMetrics(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	handler := WithMetrics(next)

	req := httptest.NewRequest(http.MethodGet, "/api/custom-path", http.NoBody)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTeapot {
		t.Errorf("expected final response code %d, got %d", http.StatusTeapot, rr.Code)
	}

	var buf bytes.Buffer
	metrics.WritePrometheus(&buf, false)
	output := buf.String()

	expectedMetric := `http_request_duration_seconds_count{status="418",url="/api/custom-path"}`

	if !strings.Contains(output, expectedMetric) {
		t.Errorf("expected metrics to contain %q, but got output:\n%s", expectedMetric, output)
	}
}
