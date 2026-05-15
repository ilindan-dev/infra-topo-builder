package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/adapters/rest"
	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/domain"
)

type mockBuilder struct {
	mockBuild func(ctx context.Context, filepath string) (uuid.UUID, error)
}

func (m *mockBuilder) BuildFromArchive(ctx context.Context, filepath string) (uuid.UUID, error) {
	return m.mockBuild(ctx, filepath)
}

type mockReader struct {
	mockGetTopology  func(ctx context.Context, logID uuid.UUID) (*domain.Topology, error)
	mockGetNodeByID  func(ctx context.Context, nodeID uuid.UUID) (domain.Node, error)
	mockGetPorts     func(ctx context.Context, nodeID uuid.UUID) ([]domain.Port, error)
	mockGetLogByID   func(ctx context.Context, logID uuid.UUID) (*domain.Log, error)
	mockGetLogStatus func(ctx context.Context, id uuid.UUID) (domain.LogStatus, error)
}

func (m *mockReader) GetTopology(ctx context.Context, logID uuid.UUID) (*domain.Topology, error) {
	return m.mockGetTopology(ctx, logID)
}

func (m *mockReader) GetNodeByID(ctx context.Context, nodeID uuid.UUID) (domain.Node, error) {
	if m.mockGetNodeByID == nil {
		return domain.Node{}, errors.New("not implemented in test")
	}
	return m.mockGetNodeByID(ctx, nodeID)
}

func (m *mockReader) GetPortsByNodeID(ctx context.Context, nodeID uuid.UUID) ([]domain.Port, error) {
	if m.mockGetPorts == nil {
		return nil, errors.New("not implemented in test")
	}
	return m.mockGetPorts(ctx, nodeID)
}

func (m *mockReader) GetLogByID(ctx context.Context, logID uuid.UUID) (*domain.Log, error) {
	if m.mockGetLogByID == nil {
		return nil, errors.New("not implemented in test")
	}
	return m.mockGetLogByID(ctx, logID)
}

func (m *mockReader) GetLogStatus(ctx context.Context, logID uuid.UUID) (domain.LogStatus, error) {
	if m.mockGetLogStatus == nil {
		return "", errors.New("not implemented in test")
	}
	return m.mockGetLogStatus(ctx, logID)
}

func setupTestMux(b *mockBuilder, r *mockReader) *http.ServeMux {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := rest.NewAPIHandler(b, r, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/parse/", handler.ParseLog)
	mux.HandleFunc("GET /api/v1/topology/{log_id}", handler.GetTopology)
	mux.HandleFunc("GET /api/v1/node/{node_id}", handler.GetNode)
	mux.HandleFunc("GET /api/v1/port/{node_id}", handler.GetPorts)
	mux.HandleFunc("GET /api/v1/log/{log_id}", handler.GetLogInfo)
	return mux
}

// TestParseLog verifies the ParseLog endpoint processes valid JSON payloads, rejects malformed
// or empty requests, and translates builder errors into appropriate HTTP statuses.
func TestParseLog(t *testing.T) {
	expectedUUID := uuid.New()

	tests := []struct {
		name         string
		body         string
		mockFunc     func(ctx context.Context, filepath string) (uuid.UUID, error)
		expectStatus int
	}{
		{
			name: "Success",
			body: `{"filepath": "data/log.zip"}`,
			mockFunc: func(_ context.Context, _ string) (uuid.UUID, error) {
				return expectedUUID, nil
			},
			expectStatus: http.StatusAccepted,
		},
		{
			name:         "Empty Body",
			body:         ``,
			mockFunc:     nil,
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "Missing Filepath",
			body:         `{"wrong_field": "test"}`,
			mockFunc:     nil,
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "Builder Error",
			body: `{"filepath": "data/broken.zip"}`,
			mockFunc: func(_ context.Context, _ string) (uuid.UUID, error) {
				return uuid.Nil, errors.New("internal mock error")
			},
			expectStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &mockBuilder{mockBuild: tt.mockFunc}
			mux := setupTestMux(b, nil)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/parse/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("expected status %d, got %d", tt.expectStatus, rr.Code)
			}

			if tt.expectStatus == http.StatusAccepted {
				var resp rest.ParseResponse
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.LogID != expectedUUID {
					t.Errorf("expected UUID %v, got %v", expectedUUID, resp.LogID)
				}
			}
		})
	}
}

// TestGetTopology verifies the GetTopology endpoint validates the UUID parameter,
// returns 400 on bad UUID format, 404 when the topology is not found, and 200 on success.
func TestGetTopology(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name         string
		logIDPath    string
		mockFunc     func(ctx context.Context, logID uuid.UUID) (*domain.Topology, error)
		expectStatus int
	}{
		{
			name:      "Success",
			logIDPath: validID.String(),
			mockFunc: func(_ context.Context, _ uuid.UUID) (*domain.Topology, error) {
				return &domain.Topology{}, nil
			},
			expectStatus: http.StatusOK,
		},
		{
			name:         "Invalid UUID Format",
			logIDPath:    "not-a-uuid",
			mockFunc:     nil,
			expectStatus: http.StatusBadRequest,
		},
		{
			name:      "Not Found",
			logIDPath: validID.String(),
			mockFunc: func(_ context.Context, _ uuid.UUID) (*domain.Topology, error) {
				return nil, domain.ErrNotFound
			},
			expectStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &mockReader{mockGetTopology: tt.mockFunc}
			mux := setupTestMux(nil, r)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/topology/"+tt.logIDPath, http.NoBody)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectStatus, rr.Code, rr.Body.String())
			}
		})
	}
}

// TestGetNode verifies the GetNode endpoint returns node data on success and maps
// domain errors (e.g., invalid data) to HTTP 400 responses.
func TestGetNode(t *testing.T) {
	validID := uuid.New()

	tests := []struct {
		name         string
		nodeIDPath   string
		mockFunc     func(ctx context.Context, nodeID uuid.UUID) (domain.Node, error)
		expectStatus int
	}{
		{
			name:       "Success",
			nodeIDPath: validID.String(),
			mockFunc: func(_ context.Context, nodeID uuid.UUID) (domain.Node, error) {
				return domain.Node{ID: nodeID}, nil
			},
			expectStatus: http.StatusOK,
		},
		{
			name:       "Invalid Data Domain Error",
			nodeIDPath: validID.String(),
			mockFunc: func(_ context.Context, _ uuid.UUID) (domain.Node, error) {
				return domain.Node{}, domain.ErrInvalidData
			},
			expectStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &mockReader{mockGetNodeByID: tt.mockFunc}
			mux := setupTestMux(nil, r)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/node/"+tt.nodeIDPath, http.NoBody)
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			if rr.Code != tt.expectStatus {
				t.Errorf("expected status %d, got %d", tt.expectStatus, rr.Code)
			}
		})
	}
}
