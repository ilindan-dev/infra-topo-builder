// Package ports defines the persistence ports used by the core application.
// The interface surface is organized with CQRS in mind: Commands (mutating
// operations) are separated from Queries (read-only retrievals). Concrete
// adapters may choose to implement a unified Repository or split Command and
// Query sides for scalability, caching, or replication purposes.
package ports

import (
	"context"

	"github.com/google/uuid"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/domain"
)

// TopologyReader exposes read-only Query methods used by API layers and other consumers.
// Queries must not modify state and should be safe to call frequently.
type TopologyReader interface {
	// GetTopology returns the assembled topology (nodes and their ports) for a
	// specific parsing session (logID). This is a convenience read used by APIs
	// and export tooling.
	GetTopology(ctx context.Context, logID uuid.UUID) (*domain.Topology, error)

	// GetNodeByID returns a single node; callers can use this to enrich or verify
	// parsed node data. Returns domain.ErrNotFound when the node is missing.
	GetNodeByID(ctx context.Context, id uuid.UUID) (domain.Node, error)

	// GetPortsByNodeID returns all ports that belong to the given node.
	GetPortsByNodeID(ctx context.Context, nodeID uuid.UUID) ([]domain.Port, error)

	// GetLogByID returns the full domain.Log aggregate or an ErrNotFound when missing.
	GetLogByID(ctx context.Context, id uuid.UUID) (*domain.Log, error)

	// GetLogStatus returns only the current status for lightweight checks.
	GetLogStatus(ctx context.Context, id uuid.UUID) (domain.LogStatus, error)
}

// ParserWriter exposes Command methods used by the parser/FSM to persist parsed data.
// Commands are mutating operations (inserts/updates) and should be implemented
// to be efficient for bulk imports and clear about idempotency guarantees.
type ParserWriter interface {
	// CreateLog creates a new parsing session record with the given initial status
	// and returns the created domain.Log (including its generated ID).
	CreateLog(ctx context.Context, status domain.LogStatus) (*domain.Log, error)

	// UpdateLogStatus updates only the status field of an existing log record.
	UpdateLogStatus(ctx context.Context, id uuid.UUID, status domain.LogStatus) error

	// UpdateLogCounts updates aggregated counters (nodes and ports) after parsing is done.
	UpdateLogCounts(ctx context.Context, id uuid.UUID, nodesCount, portsCount int32) error

	// InsertNodes and InsertPorts are intended for high-throughput bulk inserts
	// used during import; implementations may use COPY/driver-specific bulk APIs.
	InsertNodes(ctx context.Context, nodes []domain.Node) error
	InsertPorts(ctx context.Context, ports []domain.Port) error

	// UpsertNodesInfo and UpsertSwitches perform idempotent upserts for enrichment
	// records (nodes info, switch capabilities). They should be safe to call
	// multiple times during incremental imports.
	UpsertNodesInfo(ctx context.Context, info []domain.NodesInfo) error
	UpsertSwitches(ctx context.Context, switches []domain.Switch) error

	// DeleteLogData deletes all parsed data (nodes, ports, switches) for the log,
	// leaving the log entry itself untouched for error history.
	DeleteLogData(ctx context.Context, logID uuid.UUID) error
}

// LogManager groups lifecycle operations (e.g., deletes) that may be used by
// background jobs or administrative/cleanup endpoints.
type LogManager interface {
	// DeleteLog removes a log record and (optionally) its related data. Implementations
	// should document whether deletes cascade or require separate cleanup calls.
	DeleteLog(ctx context.Context, id uuid.UUID) error
}
