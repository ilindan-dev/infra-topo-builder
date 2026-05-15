// Package rest implements the HTTP REST adapter for the topology builder.
//
// The package exposes HTTP handlers that translate JSON/HTTP requests into
// core domain operations (builders and readers) via well-defined ports. It
// follows the adapter pattern: the REST layer handles transport concerns
// (routing, JSON encoding/decoding, HTTP status codes, middleware) and
// delegates business logic to the core port interfaces.
//
// Endpoints (high level):
// - POST   /api/v1/parse/              - start parsing an archive, returns log_id
// - GET    /api/v1/topology/{log_id}   - retrieve topology for a log
// - GET    /api/v1/node/{node_id}      - retrieve a single node by ID
// - GET    /api/v1/port/{node_id}      - retrieve ports for a node
// - GET    /api/v1/log/{log_id}        - retrieve import/log metadata
// - GET    /ping                       - health check
//
// Middleware
// Middleware applied to the routes (see route.SetupRoutes): Recovery, WithLogger,
// WithMetrics, Rate. Recovery should be outermost to catch panics from any
// downstream layer; metrics and logging wrap the business handlers to ensure
// observability. Rate provides process-global throttling; implement keyed
// limiters if per-client limits are required.
//
// Error handling
// This adapter maps domain-level errors (defined in core/domain) to HTTP
// status codes: ErrNotFound -> 404, ErrInvalidData -> 400, other domain errors
// -> 500. Handlers should preserve context.Context cancellation semantics and
// avoid blocking operations directly in request goroutines (long-running work
// is delegated to background workers via the service layer).
package rest
