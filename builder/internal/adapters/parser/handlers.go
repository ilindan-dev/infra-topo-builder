package parser

import (
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/domain"
	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/utils"
)

func cleanGUID(guid string) string {
	g := strings.TrimSpace(guid)
	g = strings.ToLower(g)
	if g != "" && !strings.HasPrefix(g, "0x") {
		g = "0x" + g
	}
	return g
}

// parseNodeLine converts a CSV node line into a domain.Node. It validates
// the expected column count and returns an error for malformed lines. The
// domain.Node.ID is a deterministic UUID derived from (logID, nodeGUID).
func parseNodeLine(logID uuid.UUID, line string) (domain.Node, error) {
	parts := strings.Split(line, ",")
	if len(parts) < 8 {
		return domain.Node{}, errors.New("invalid node line: insufficient columns")
	}

	nodeDesc := strings.Trim(parts[0], `"`)
	nodeTypeInt := utils.ParseInt32(parts[2])

	nodeKind, err := utils.ParseNodeKind(nodeTypeInt)
	if err != nil {
		return domain.Node{}, err
	}

	nodeGUID := cleanGUID(parts[6])
	portGUID := cleanGUID(parts[7])

	return domain.Node{
		ID:              utils.GenerateDeterministicUUID(logID, nodeGUID),
		LogID:           logID,
		NodeDesc:        nodeDesc,
		NumPorts:        utils.ParseInt32(parts[1]),
		NodeType:        nodeKind,
		ClassVersion:    utils.ParseInt32(parts[3]),
		BaseVersion:     utils.ParseInt32(parts[4]),
		SystemImageGUID: cleanGUID(parts[5]),
		NodeGUID:        nodeGUID,
		PortGUID:        portGUID,
	}, nil
}

// parseSwitchLine converts a CSV switch capability line into a domain.Switch.
// Returns an error if the expected column count is not met.
func parseSwitchLine(logID uuid.UUID, line string) (domain.Switch, error) {
	parts := strings.Split(line, ",")
	if len(parts) < 19 {
		return domain.Switch{}, errors.New("invalid switch line: insufficient columns")
	}

	nodeGUID := cleanGUID(parts[0])

	return domain.Switch{
		ID:                   uuid.New(),
		NodeID:               utils.GenerateDeterministicUUID(logID, nodeGUID),
		NodeGUID:             nodeGUID,
		LinearFDBCap:         utils.ParsePtrInt32(parts[1]),
		RandomFDBCap:         utils.ParsePtrInt32(parts[2]),
		MCastFDBCap:          utils.ParsePtrInt32(parts[3]),
		LinearFDBTop:         utils.ParsePtrInt32(parts[4]),
		DefPort:              utils.ParsePtrInt32(parts[5]),
		DefMCastPriPort:      utils.ParsePtrInt32(parts[6]),
		DefMCastNotPriPort:   utils.ParsePtrInt32(parts[7]),
		LifeTimeValue:        utils.ParsePtrInt32(parts[8]),
		PortStateChange:      utils.ParsePtrInt32(parts[9]),
		OptimizedSLVLMapping: utils.ParsePtrInt32(parts[10]),
		LidsPerPort:          utils.ParsePtrInt32(parts[11]),
		PartEnfCap:           utils.ParsePtrInt32(parts[12]),
		InbEnfCap:            utils.ParsePtrInt32(parts[13]),
		OutbEnfCap:           utils.ParsePtrInt32(parts[14]),
		FilterRawInbCap:      utils.ParsePtrInt32(parts[15]),
		FilterRawOutbCap:     utils.ParsePtrInt32(parts[16]),
		ENP0:                 utils.ParsePtrInt32(parts[17]),
		MCastFDBTop:          utils.ParsePtrInt32(parts[18]),
	}, nil
}

// parsePortLine parses a port CSV line into a domain.Port. The function
// validates the column count and maps nullable values using the utils helpers.
func parsePortLine(logID uuid.UUID, line string) (domain.Port, error) {
	parts := strings.Split(line, ",")
	if len(parts) < 54 {
		return domain.Port{}, errors.New("invalid port line: insufficient columns")
	}

	nodeGUID := cleanGUID(parts[0])
	portGUID := cleanGUID(parts[1])

	return domain.Port{
		ID:                                  uuid.New(),
		NodeID:                              utils.GenerateDeterministicUUID(logID, nodeGUID),
		NodeGUID:                            nodeGUID,
		PortGUID:                            portGUID,
		PortNum:                             utils.ParseInt32(parts[2]),
		MKey:                                strings.TrimSpace(parts[3]),
		GidPrfx:                             strings.TrimSpace(parts[4]),
		MsmLid:                              utils.ParsePtrInt32(parts[5]),
		Lid:                                 utils.ParsePtrInt32(parts[6]),
		CapMsk:                              utils.ParsePtrInt64(parts[7]),
		MKeyLeasePeriod:                     utils.ParsePtrInt32(parts[8]),
		DiagCode:                            utils.ParsePtrInt32(parts[9]),
		LinkWidthActv:                       utils.ParsePtrInt32(parts[10]),
		LinkWidthSup:                        utils.ParsePtrInt32(parts[11]),
		LinkWidthEn:                         utils.ParsePtrInt32(parts[12]),
		LocalPortNum:                        utils.ParsePtrInt32(parts[13]),
		LinkSpeedEn:                         utils.ParsePtrInt32(parts[14]),
		LinkSpeedActv:                       utils.ParsePtrInt32(parts[15]),
		Lmc:                                 utils.ParsePtrInt32(parts[16]),
		MKeyProtBits:                        utils.ParsePtrInt32(parts[17]),
		LinkDownDefState:                    utils.ParsePtrInt32(parts[18]),
		PortPhyState:                        utils.ParsePtrInt32(parts[19]),
		PortState:                           utils.ParsePtrInt32(parts[20]),
		LinkSpeedSup:                        utils.ParsePtrInt32(parts[21]),
		VlArbHighCap:                        utils.ParsePtrInt32(parts[22]),
		VlHighLimit:                         utils.ParsePtrInt32(parts[23]),
		InitType:                            utils.ParsePtrInt32(parts[24]),
		VlCap:                               utils.ParsePtrInt32(parts[25]),
		Msmsl:                               utils.ParsePtrInt32(parts[26]),
		Nmtu:                                utils.ParsePtrInt32(parts[27]),
		FilterRawOutb:                       utils.ParsePtrInt32(parts[28]),
		FilterRawInb:                        utils.ParsePtrInt32(parts[29]),
		PartEnfOutb:                         utils.ParsePtrInt32(parts[30]),
		PartEnfInb:                          utils.ParsePtrInt32(parts[31]),
		OpVls:                               utils.ParsePtrInt32(parts[32]),
		HoqLife:                             utils.ParsePtrInt32(parts[33]),
		VlStallCnt:                          utils.ParsePtrInt32(parts[34]),
		MtuCap:                              utils.ParsePtrInt32(parts[35]),
		InitTypeReply:                       utils.ParsePtrInt32(parts[36]),
		VlArbLowCap:                         utils.ParsePtrInt32(parts[37]),
		PKeyViolations:                      utils.ParsePtrInt32(parts[38]),
		MKeyViolations:                      utils.ParsePtrInt32(parts[39]),
		SubnTmo:                             utils.ParsePtrInt32(parts[40]),
		MulticastPKeyTrapSuppressionEnabled: utils.ParsePtrInt32(parts[41]),
		ClientReregister:                    utils.ParsePtrInt32(parts[42]),
		GUIDCap:                             utils.ParsePtrInt32(parts[43]),
		QKeyViolations:                      utils.ParsePtrInt32(parts[44]),
		MaxCreditHint:                       utils.ParsePtrInt32(parts[45]),
		OverrunErrs:                         utils.ParsePtrInt32(parts[46]),
		LocalPhyError:                       utils.ParsePtrInt32(parts[47]),
		RespTimeValue:                       utils.ParsePtrString(parts[48]),
		LinkRoundTripLatency:                utils.ParsePtrString(parts[49]),
		OooSlMask:                           utils.ParsePtrString(parts[50]),
		CapMsk2:                             utils.ParsePtrString(parts[51]),
		FecActv:                             utils.ParsePtrString(parts[52]),
		RetransActv:                         utils.ParsePtrString(parts[53]),
	}, nil
}

// parseSystemInfoLine converts a .sharp_an_info record line into a domain.NodesInfo.
// It expects at least [nodeGUID, serial, part, revision, product_name] columns.
func parseSystemInfoLine(logID uuid.UUID, line string) (domain.NodesInfo, error) {
	parts := strings.Split(line, ",")
	if len(parts) < 5 {
		return domain.NodesInfo{}, errors.New("invalid system info line: insufficient columns")
	}

	nodeGUID := cleanGUID(parts[0])

	return domain.NodesInfo{
		ID:           uuid.New(),
		NodeID:       utils.GenerateDeterministicUUID(logID, nodeGUID),
		SerialNumber: utils.ParsePtrString(parts[1]),
		PartNumber:   utils.ParsePtrString(parts[2]),
		Revision:     utils.ParsePtrString(parts[3]),
		ProductName:  utils.ParsePtrString(parts[4]),
	}, nil
}
