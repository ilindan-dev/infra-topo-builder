package ports

import (
	"context"
	"io"

	"github.com/google/uuid"
)

// TopologyParser defines the port for extracting topology data from raw file streams.
// The implementation (adapter) is responsible for reading the specific file formats
// (e.g., db_csv, sharp_an_info) and persisting the chunks to the database via
// the ParserWriter port.
type TopologyParser interface {
	// ParseDBCSV reads the CSV stream, extracts nodes and ports using an FSM,
	// and flushes them to the database in batches.
	ParseDBCSV(ctx context.Context, logID uuid.UUID, reader io.Reader) error

	// ParseSharpAnInfo reads the INI-like stream, extracts switch and node
	// hardware configurations, and upserts them to the database.
	ParseSharpAnInfo(ctx context.Context, logID uuid.UUID, reader io.Reader) error
}
