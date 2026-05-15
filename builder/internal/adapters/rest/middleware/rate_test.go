package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRateMiddleware checks the Rate middleware allows one request and then throttles additional requests
// returning HTTP 429 Too Many Requests for the second request.
func TestRateMiddleware(t *testing.T) {
	handler := Rate(1)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", http.NoBody)

	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req)
	if status := rr1.Code; status != http.StatusOK {
		t.Errorf("First request: got status %v, want %v", status, http.StatusOK)
	}

	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	if status := rr2.Code; status != http.StatusTooManyRequests {
		t.Errorf("Second request (throttled): got status %v, want %v", status, http.StatusTooManyRequests)
	}
}
