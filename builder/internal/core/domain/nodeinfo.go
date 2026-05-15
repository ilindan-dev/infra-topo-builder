package domain

import "github.com/google/uuid"

// NodesInfo contains additional hardware metadata for a specific node,
// typically aggregated from the .sharp_an_info configuration file.
type NodesInfo struct {
	// ID is the unique internal identifier for the metadata record.
	ID uuid.UUID
	// NodeID references the parent Node entity.
	NodeID uuid.UUID
	// SerialNumber is the hardware serial number of the device.
	SerialNumber *string
	// PartNumber is the manufacturer's part number.
	PartNumber *string
	// Revision is the hardware revision code.
	Revision *string
	// ProductName is the commercial name of the equipment (e.g., Mellanox).
	ProductName *string
	// Endianness specifies the byte order (e.g., 0 for Little-Endian, 1 for Big-Endian).
	Endianness *int32
	// EnableEndiannessPerJob indicates if endianness is dynamically configured per job.
	EnableEndiannessPerJob *int32
	// ReproducibilityDisable indicates whether reproducibility constraints are disabled.
	ReproducibilityDisable *int32
}
