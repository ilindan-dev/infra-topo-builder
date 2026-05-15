// Package ports defines the persistence and read/write boundaries (ports)
// used by the core domain. The package takes a CQRS-inspired approach and
// separates mutating commands (used by the parser and other writers) from
// read-only queries (used by consumers that present or analyse topology).
//
// Design overview
// - Commands (mutating ports)
//   - ParserWriter: bulk-oriented methods invoked by the streaming parser to
//     persist Nodes, Ports, NodesInfo and Switch records. Implementations are
//     expected to be efficient (support COPY/batched statements) and to behave
//     idempotently where possible. Bulk methods should accept slices and may
//     return domain-level errors describing per-item failures or the first
//     encountered error.
//   - LogManager: create, update and delete import/log records and manage the
//     lifecycle of parsing jobs.
//
// - Queries (read-only ports)
//   - TopologyReader: read-side methods that return domain.Topology and
//     related entities (GetTopology, GetNodeByID, GetPortsByNodeID,
//     GetLogByID, GetLogStatus). Readers should be safe for concurrent use and
//     return domain.ErrNotFound when an entity is absent.
//
// Error and context contract
// All port methods take context.Context and should translate low-level driver
// errors into domain-level errors (ErrNotFound, ErrAlreadyExists,
// ErrConflict, ErrInvalidData, ErrInternalServer). Adapters are responsible for
// mapping driver-specific constraints (unique index names, pg error codes) to
// these canonical errors and for thorough logging of driver details.
//
// Batching and deterministic IDs
// Parser-facing commands commonly accept large batches. Implementations should
// avoid unbounded memory growth and prefer streaming/bulk DB primitives. When
// callers provide deterministic IDs (for example name-based UUIDs derived by
// the parser), adapters should preserve and use those IDs rather than
// generating new ones to maintain idempotency across repeated imports.
//
// Implementation notes
// Concrete adapter implementations live in internal/adapters/* (for example
// internal/adapters/postgres). Those adapters should document any additional
// constraints such as SQLC batch execution semantics, transactional scopes,
// and migration requirements.
package ports
