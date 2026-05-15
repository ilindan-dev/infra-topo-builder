package domain

import "github.com/google/uuid"

// Switch represents the specific telemetry, capabilities, and configuration
// of a network switch element extracted from the START_SWITCHES section.
type Switch struct {
	// ID is the unique internal identifier for the switch record.
	ID uuid.UUID
	// NodeID references the parent Node to which this switch belongs.
	// It is deterministically generated to link with the nodes table.
	NodeID uuid.UUID
	// NodeGUID is the globally unique identifier of the parent node.
	NodeGUID string

	// Telemetry and hardware capabilities.
	// Stored as pointers to handle missing data or "N/A" values gracefully,
	// mapping them to NULL in the database instead of false zeros.
	LinearFDBCap         *int32
	RandomFDBCap         *int32
	MCastFDBCap          *int32
	LinearFDBTop         *int32
	DefPort              *int32
	DefMCastPriPort      *int32
	DefMCastNotPriPort   *int32
	LifeTimeValue        *int32
	PortStateChange      *int32
	OptimizedSLVLMapping *int32
	LidsPerPort          *int32
	PartEnfCap           *int32
	InbEnfCap            *int32
	OutbEnfCap           *int32
	FilterRawInbCap      *int32
	FilterRawOutbCap     *int32
	ENP0                 *int32
	MCastFDBTop          *int32
}
