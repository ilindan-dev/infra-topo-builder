package ports

import (
	"context"

	"github.com/google/uuid"
)

// TopologyBuilder defines the primary inbound port (Use Case) for the application.
// It acts as the orchestrator for the complex process of unpacking and parsing
// diagnostic archives.
type TopologyBuilder interface {
	// BuildFromArchive initiates the parsing of an ibdiagnet archive located
	// at the given filepath (relative to the mounted data/ directory).
	// It creates a new log record in the database, starts the asynchronous
	// parsing process, and returns the generated log ID immediately so the
	// client can track its status.
	BuildFromArchive(ctx context.Context, filepath string) (uuid.UUID, error)
}
