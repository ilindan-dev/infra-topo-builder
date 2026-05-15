package domain

import (
	"time"

	"github.com/google/uuid"
)

// LogStatus represents the current state of a parsing session.
type LogStatus string

const (
	// StatusPending indicates that parsing has not started yet and is pending.
	StatusPending LogStatus = "pending"
	// StatusInProgress indicates that the parsing is in progress.
	StatusInProgress LogStatus = "in_progress"
	// StatusSuccess indicates that the parsing was completed successfully.
	StatusSuccess LogStatus = "success"
	// StatusError indicates that an error occurred during parsing.
	StatusError LogStatus = "error"
)

// Log represents a parsing session for an ibdiagnet topology file.
// It acts as the aggregate root for all nodes and ports parsed during a single run.
type Log struct {
	// ID is the unique identifier of the log entry.
	ID uuid.UUID
	// Status indicates the current state of the parsing process (e.g., "pending", "success", "error").
	Status LogStatus
	// NodesCount holds the aggregated number of successfully parsed nodes.
	NodesCount int32
	// PortsCount holds the aggregated number of successfully parsed ports.
	PortsCount int32
	// CreatedAt is the timestamp when the parsing session was initiated.
	CreatedAt time.Time
}
