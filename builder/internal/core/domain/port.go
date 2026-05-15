package domain

import "github.com/google/uuid"

// Port represents the telemetry and configuration state of a single InfiniBand port.
// It contains metrics extracted from the START_PORTS section of the diagnostic log.
type Port struct {
	// ID is the unique internal identifier for the port record.
	ID uuid.UUID
	// NodeID references the parent Node to which this port belongs.
	NodeID uuid.UUID
	// NodeGUID is the global identifier of the parent node.
	NodeGUID string
	// PortGUID is the global identifier of this specific port.
	PortGUID string
	// PortNum is the physical index number of the port on the node.
	PortNum int32
	// MKey is the management key associated with the port.
	MKey string
	// GidPrfx is the Global Identifier prefix for the subnet.
	GidPrfx string

	// Telemetry and configuration metrics (nullable fields handle "N/A" or missing data).
	MsmLid                              *int32
	Lid                                 *int32
	CapMsk                              *int64
	MKeyLeasePeriod                     *int32
	DiagCode                            *int32
	LinkWidthActv                       *int32
	LinkWidthSup                        *int32
	LinkWidthEn                         *int32
	LocalPortNum                        *int32
	LinkSpeedEn                         *int32
	LinkSpeedActv                       *int32
	Lmc                                 *int32
	MKeyProtBits                        *int32
	LinkDownDefState                    *int32
	PortPhyState                        *int32
	PortState                           *int32
	LinkSpeedSup                        *int32
	VlArbHighCap                        *int32
	VlHighLimit                         *int32
	InitType                            *int32
	VlCap                               *int32
	Msmsl                               *int32
	Nmtu                                *int32
	FilterRawOutb                       *int32
	FilterRawInb                        *int32
	PartEnfOutb                         *int32
	PartEnfInb                          *int32
	OpVls                               *int32
	HoqLife                             *int32
	VlStallCnt                          *int32
	MtuCap                              *int32
	InitTypeReply                       *int32
	VlArbLowCap                         *int32
	PKeyViolations                      *int32
	MKeyViolations                      *int32
	SubnTmo                             *int32
	MulticastPKeyTrapSuppressionEnabled *int32
	ClientReregister                    *int32
	GUIDCap                             *int32
	QKeyViolations                      *int32
	MaxCreditHint                       *int32
	OverrunErrs                         *int32
	LocalPhyError                       *int32

	// String-based metrics that may contain explicit "N/A" values.
	RespTimeValue        *string
	LinkRoundTripLatency *string
	OooSlMask            *string
	CapMsk2              *string
	FecActv              *string
	RetransActv          *string
}
