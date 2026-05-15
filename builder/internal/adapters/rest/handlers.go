package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/domain"
	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/ports"
)

// APIHandler handles HTTP requests for the REST API and delegates business
// operations to the core ports. It intentionally remains thin: all domain
// logic is executed by the provided TopologyBuilder and TopologyReader
// interfaces. The handler centralises JSON encoding/decoding, HTTP status
// mapping and simple validation.
//
// The handler's methods are safe to call concurrently from multiple
// goroutines because they only access immutable references to the injected
// ports and logger.
type APIHandler struct {
	builder ports.TopologyBuilder
	reader  ports.TopologyReader
	logger  *slog.Logger
}

// NewAPIHandler constructs an APIHandler wired with the builder and reader
// port implementations and a structured logger. The logger will be annotated
// with "layer=rest" to help with log filtering.
func NewAPIHandler(b ports.TopologyBuilder, r ports.TopologyReader, l *slog.Logger) *APIHandler {
	return &APIHandler{
		builder: b,
		reader:  r,
		logger:  l.With("layer", "rest"),
	}
}

// ParseRequest represents the JSON body expected by the ParseLog endpoint.
// Filepath should point to a zip archive accessible by the server process.
// The handler validates that the filepath is non-empty before scheduling
// background processing.
type ParseRequest struct {
	Filepath string `json:"filepath"`
}

// ParseResponse is returned by ParseLog when a parse job is accepted. The
// LogID field identifies the created parsing session and can be used by
// callers to poll status via the read-side APIs.
type ParseResponse struct {
	LogID uuid.UUID `json:"log_id"`
}

// ParseLog handles POST /api/v1/parse/
// The endpoint accepts a JSON body matching ParseRequest and schedules a
// background import job via the builder port. On success it returns HTTP
// 202 Accepted and a ParseResponse containing the created log ID. Validation
// errors return HTTP 400; unexpected failures return HTTP 500.
func (h *APIHandler) ParseLog(w http.ResponseWriter, r *http.Request) {
	var req ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, r, http.StatusBadRequest, "invalid json body")
		return
	}

	if req.Filepath == "" {
		h.sendError(w, r, http.StatusBadRequest, "filepath is required")
		return
	}

	logID, err := h.builder.BuildFromArchive(r.Context(), req.Filepath)
	if err != nil {
		h.sendError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendJSON(w, http.StatusAccepted, ParseResponse{LogID: logID})
}

// GetTopology handles GET /api/v1/topology/{log_id}
// Returns the complete domain.Topology for the provided log ID. If the
// log_id is malformed a 400 is returned. If no topology exists for the
// log_id a 404 is returned. On success returns HTTP 200 with the topology
// encoded as JSON.
func (h *APIHandler) GetTopology(w http.ResponseWriter, r *http.Request) {
	logID, err := uuid.Parse(r.PathValue("log_id"))
	if err != nil {
		h.sendError(w, r, http.StatusBadRequest, "invalid log_id format")
		return
	}

	topology, err := h.reader.GetTopology(r.Context(), logID)
	if err != nil {
		h.handleDomainError(w, r, err)
		return
	}

	h.sendJSON(w, http.StatusOK, topology)
}

// GetNode handles GET /api/v1/node/{node_id}
// Returns a single domain.Node identified by node_id. Returns 400 for
// invalid UUID format or 404 when the node cannot be found.
func (h *APIHandler) GetNode(w http.ResponseWriter, r *http.Request) {
	nodeID, err := uuid.Parse(r.PathValue("node_id"))
	if err != nil {
		h.sendError(w, r, http.StatusBadRequest, "invalid node_id format")
		return
	}

	node, err := h.reader.GetNodeByID(r.Context(), nodeID)
	if err != nil {
		h.handleDomainError(w, r, err)
		return
	}

	h.sendJSON(w, http.StatusOK, node)
}

// GetPorts handles GET /api/v1/port/{node_id}
// Returns all ports associated with the specified node. Returns 400 for an
// invalid node_id format and 404 if the node does not exist.
func (h *APIHandler) GetPorts(w http.ResponseWriter, r *http.Request) {
	nodeID, err := uuid.Parse(r.PathValue("node_id"))
	if err != nil {
		h.sendError(w, r, http.StatusBadRequest, "invalid node_id format")
		return
	}

	portList, err := h.reader.GetPortsByNodeID(r.Context(), nodeID)
	if err != nil {
		h.handleDomainError(w, r, err)
		return
	}

	h.sendJSON(w, http.StatusOK, portList)
}

// GetLogInfo handles GET /api/v1/log/{log_id}
// Returns metadata about a parsing session, including status and counters.
// Returns 400 for malformed IDs and 404 if the log is not found.
func (h *APIHandler) GetLogInfo(w http.ResponseWriter, r *http.Request) {
	logID, err := uuid.Parse(r.PathValue("log_id"))
	if err != nil {
		h.sendError(w, r, http.StatusBadRequest, "invalid log_id format")
		return
	}

	logInfo, err := h.reader.GetLogByID(r.Context(), logID)
	if err != nil {
		h.handleDomainError(w, r, err)
		return
	}

	h.sendJSON(w, http.StatusOK, logInfo)
}

// sendJSON encodes the provided data as JSON and writes it with the
// provided HTTP status code. The function logs encoding errors but does not
// send additional error responses since it is usually called after headers
// have been written.
func (h *APIHandler) sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode json response", "error", err)
	}
}

// sendError sends a JSON-formatted error response and logs a warning with
// the request path and status. Used for both client and server-side errors.
func (h *APIHandler) sendError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	h.logger.Warn("request failed", "status", status, "path", r.URL.Path, "error", msg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// handleDomainError translates domain-layer errors into appropriate HTTP
// responses. Known domain errors are mapped to 4xx responses; all unknown
// errors are logged and returned as HTTP 500 to avoid leaking internal
// details to clients.
func (h *APIHandler) handleDomainError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		h.sendError(w, r, http.StatusNotFound, "resource not found")
	case errors.Is(err, domain.ErrInvalidData):
		h.sendError(w, r, http.StatusBadRequest, err.Error())
	default:
		h.logger.Error("internal error", "error", err)
		h.sendError(w, r, http.StatusInternalServerError, "internal server error")
	}
}
