// Package service provides the application-level orchestration for building
// topology data from archives. The service layer sits between the core domain
// and adapters (persistence, parsing) and coordinates long-running import
// jobs, status updates and error handling.
//
// Responsibilities
//   - Accept user requests to start imports and create import logs.
//   - Run parsing of archive artifacts and persist results using the
//     ports.ParserWriter interface.
//   - Update import log lifecycle statuses (pending, in_progress, success,
//     error) via ports.LogManager/ParserWriter implementations.
//   - Execute heavy work in background goroutines so HTTP handlers can return
//     immediately with a log ID that callers can poll.
//
// Error handling and contracts
// Service methods return domain-level errors (ErrInvalidData, ErrNotFound,
// etc.) when appropriate and otherwise wrap lower-level errors to provide
// contextual information. All exported methods accept context.Context and
// should respect cancellation for immediate validation steps; long-running
// background workers intentionally use context.WithoutCancel where continued
// processing is desired even if the caller cancels.
package service
