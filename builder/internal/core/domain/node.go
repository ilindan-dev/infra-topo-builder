package domain

import (
	"github.com/google/uuid"
)

// NodeKind represents type of logical network element.
type NodeKind string

const (
	// NodeInvalidKind indicates that input kind is unexpected.
	NodeInvalidKind NodeKind = "not_implemented_kind"
	// NodeHost indicates that the node is a host.
	NodeHost NodeKind = "host"
	// NodeSwitch indicates that the node is a switch.
	NodeSwitch NodeKind = "switch"
)

// Node represents a physical or logical network element (Host or Switch)
// extracted from the START_NODES section of the diagnostic log.
type Node struct {
	// ID is the unique internal identifier generated deterministically.
	ID uuid.UUID
	// LogID references the parsing session that created this node.
	LogID uuid.UUID
	// NodeDesc is the human-readable description of the node (e.g., "SWITCH_1").
	NodeDesc string
	// NumPorts indicates the number of ports declared on this node.
	NumPorts int32
	// NodeType represents the kind of network equipment ("host" or "switch").
	// It maps to the node_kind ENUM in the database.
	NodeType NodeKind
	// ClassVersion is the supported management class version.
	ClassVersion int32
	// BaseVersion is the base version of the management architecture.
	BaseVersion int32
	// SystemImageGUID is the globally unique identifier of the system image.
	SystemImageGUID string
	// NodeGUID is the globally unique identifier of the node itself.
	NodeGUID string
	// PortGUID is the GUID of the management port associated with this node.
	PortGUID string

	// --- Enrichment Data---

	// Info contains optional hardware metadata (from .sharp_an_info).
	Info *NodesInfo
	// SwitchInfo contains switch capabilities (only populated if NodeType == NodeSwitch).
	SwitchInfo *Switch
}
