package postgres

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/adapters/postgres/gen"
	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/domain"
	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/ports"
)

var _ ports.TopologyReader = (*topologyReader)(nil)

// topologyReader implements ports.TopologyReader and provides read-only access
// to the stored topology and related entities. It maps sqlc-generated rows into
// domain models and intentionally avoids any modification of database state.
// The implementation is optimized for convenience and clarity; heavy read-path
// optimizations (caching, pagination) can be added in a separate query adapter.
type topologyReader struct {
	pool    *pgxpool.Pool
	queries *gen.Queries
	logger  *slog.Logger
}

// NewTopologyReader constructs a TopologyReader backed by Postgres. The provided
// logger will be annotated with adapter metadata for observability. Use this
// adapter in API handlers or export tools that need to read assembled topologies.
func NewTopologyReader(pool *pgxpool.Pool, logger *slog.Logger) ports.TopologyReader {
	return &topologyReader{
		pool:    pool,
		queries: gen.New(pool),
		logger:  logger.With("adapter", "postgres", "implement", "topologyReader"),
	}
}

// GetTopology returns the assembled topology (nodes with optional node info and switch info)
// for the provided parsing session (logID). It maps sqlc-generated rows into the
// domain models and omits enrichment structs when no data was present in the DB.
func (r *topologyReader) GetTopology(ctx context.Context, logID uuid.UUID) (*domain.Topology, error) {
	genTopologies, err := r.queries.GetFullTopology(ctx, logID)
	if err != nil {
		return nil, translateError(err, "GetFullTopology", r.logger)
	}

	nodes := make([]domain.Node, 0, len(genTopologies))
	for i := range genTopologies {
		row := &genTopologies[i]

		n := domain.Node{
			ID:              row.ID,
			LogID:           row.LogID,
			NodeDesc:        row.NodeDesc,
			NumPorts:        row.NumPorts,
			NodeType:        row.NodeType,
			ClassVersion:    row.ClassVersion,
			BaseVersion:     row.BaseVersion,
			SystemImageGUID: row.SystemImageGuid,
			NodeGUID:        row.NodeGuid,
			PortGUID:        row.PortGuid,
			Info:            mapNodeInfo(row),
			SwitchInfo:      mapSwitchInfo(row),
		}

		nodes = append(nodes, n)
	}

	return &domain.Topology{
		LogID: logID,
		Nodes: nodes,
	}, nil
}

// GetNodeByID returns a single domain.Node mapped from the underlying nodes table.
func (r *topologyReader) GetNodeByID(ctx context.Context, id uuid.UUID) (domain.Node, error) {
	row, err := r.queries.GetNodeByID(ctx, id)
	if err != nil {
		return domain.Node{}, translateError(err, "GetNodeByID", r.logger)
	}
	return domain.Node{
		ID:              row.ID,
		LogID:           row.LogID,
		NodeDesc:        row.NodeDesc,
		NumPorts:        row.NumPorts,
		NodeType:        row.NodeType,
		ClassVersion:    row.ClassVersion,
		BaseVersion:     row.BaseVersion,
		SystemImageGUID: row.SystemImageGuid,
		NodeGUID:        row.NodeGuid,
		PortGUID:        row.PortGuid,
	}, nil
}

// GetPortsByNodeID returns all domain.Port records for the given node ID.
func (r *topologyReader) GetPortsByNodeID(ctx context.Context, nodeID uuid.UUID) ([]domain.Port, error) {
	genPorts, err := r.queries.GetPortsByNodeID(ctx, nodeID)
	if err != nil {
		return nil, translateError(err, "GetPortsByNodeID", r.logger)
	}
	portList := make([]domain.Port, 0, len(genPorts))
	for i := range genPorts {
		p := &genPorts[i]
		portList = append(portList, domain.Port{
			ID:                                  p.ID,
			NodeID:                              p.NodeID,
			NodeGUID:                            p.NodeGuid,
			PortGUID:                            p.PortGuid,
			PortNum:                             p.PortNum,
			MKey:                                p.MKey,
			GidPrfx:                             p.GidPrfx,
			MsmLid:                              p.MsmLid,
			Lid:                                 p.Lid,
			CapMsk:                              p.CapMsk,
			MKeyLeasePeriod:                     p.MKeyLeasePeriod,
			DiagCode:                            p.DiagCode,
			LinkWidthActv:                       p.LinkWidthActv,
			LinkWidthSup:                        p.LinkWidthSup,
			LinkWidthEn:                         p.LinkWidthEn,
			LocalPortNum:                        p.LocalPortNum,
			LinkSpeedEn:                         p.LinkSpeedEn,
			LinkSpeedActv:                       p.LinkSpeedActv,
			Lmc:                                 p.Lmc,
			MKeyProtBits:                        p.MKeyProtBits,
			LinkDownDefState:                    p.LinkDownDefState,
			PortPhyState:                        p.PortPhyState,
			PortState:                           p.PortState,
			LinkSpeedSup:                        p.LinkSpeedSup,
			VlArbHighCap:                        p.VlArbHighCap,
			VlHighLimit:                         p.VlHighLimit,
			InitType:                            p.InitType,
			VlCap:                               p.VlCap,
			Msmsl:                               p.Msmsl,
			Nmtu:                                p.Nmtu,
			FilterRawOutb:                       p.FilterRawOutb,
			FilterRawInb:                        p.FilterRawInb,
			PartEnfOutb:                         p.PartEnfOutb,
			PartEnfInb:                          p.PartEnfInb,
			OpVls:                               p.OpVls,
			HoqLife:                             p.HoqLife,
			VlStallCnt:                          p.VlStallCnt,
			MtuCap:                              p.MtuCap,
			InitTypeReply:                       p.InitTypeReply,
			VlArbLowCap:                         p.VlArbLowCap,
			PKeyViolations:                      p.PKeyViolations,
			MKeyViolations:                      p.MKeyViolations,
			SubnTmo:                             p.SubnTmo,
			MulticastPKeyTrapSuppressionEnabled: p.MulticastPKeyTrapSuppressionEnabled,
			ClientReregister:                    p.ClientReregister,
			GUIDCap:                             p.GuidCap,
			QKeyViolations:                      p.QKeyViolations,
			MaxCreditHint:                       p.MaxCreditHint,
			OverrunErrs:                         p.OverrunErrs,
			LocalPhyError:                       p.LocalPhyError,
			RespTimeValue:                       p.RespTimeValue,
			LinkRoundTripLatency:                p.LinkRoundTripLatency,
			OooSlMask:                           p.OooSlMask,
			CapMsk2:                             p.CapMsk2,
			FecActv:                             p.FecActv,
			RetransActv:                         p.RetransActv,
		})
	}
	return portList, nil
}

// GetLogByID returns a domain.Log aggregate for the provided log ID.
func (r *topologyReader) GetLogByID(ctx context.Context, id uuid.UUID) (*domain.Log, error) {
	lg, err := r.queries.GetLogByID(ctx, id)
	if err != nil {
		return nil, translateError(err, "GetLogByID", r.logger)
	}
	return &domain.Log{
		ID:         lg.ID,
		Status:     lg.Status,
		NodesCount: lg.NodesCount,
		PortsCount: lg.PortsCount,
		CreatedAt:  lg.CreatedAt.Time,
	}, nil
}

// GetLogStatus returns only the current status for lightweight checks.
func (r *topologyReader) GetLogStatus(ctx context.Context, id uuid.UUID) (domain.LogStatus, error) {
	status, err := r.queries.GetLogStatus(ctx, id)
	if err != nil {
		return "", translateError(err, "GetLogStatus", r.logger)
	}
	return status, nil
}

// mapNodeInfo returns a populated domain.NodesInfo when the SQL row contains
// any non-nil enrichment columns; otherwise it returns nil to indicate absence
// of metadata. This keeps the domain Node.Info field nil when no .sharp_an_info
// data exists for the node.
func mapNodeInfo(row *gen.GetFullTopologyRow) *domain.NodesInfo {
	if row.SerialNumber == nil && row.PartNumber == nil && row.Revision == nil && row.ProductName == nil &&
		row.Endianness == nil && row.EnableEndiannessPerJob == nil && row.ReproducibilityDisable == nil {
		return nil
	}

	return &domain.NodesInfo{
		ID:                     uuid.Nil,
		NodeID:                 row.ID,
		SerialNumber:           row.SerialNumber,
		PartNumber:             row.PartNumber,
		Revision:               row.Revision,
		ProductName:            row.ProductName,
		Endianness:             row.Endianness,
		EnableEndiannessPerJob: row.EnableEndiannessPerJob,
		ReproducibilityDisable: row.ReproducibilityDisable,
	}
}

// mapSwitchInfo returns a populated domain.Switch when any switch capability
// columns are present in the SQL row; otherwise it returns nil. This mirrors
// the behavior for NodesInfo and keeps the domain SwitchInfo pointer nil when
// the database lacks enrichment for the node.
func mapSwitchInfo(row *gen.GetFullTopologyRow) *domain.Switch {
	if allNil(
		row.LinearFdbCap, row.RandomFdbCap, row.McastFdbCap, row.LinearFdbTop,
		row.DefPort, row.DefMcastPriPort, row.DefMcastNotPriPort, row.LifeTimeValue,
		row.PortStateChange, row.OptimizedSlvlMapping, row.LidsPerPort,
		row.PartEnfCap, row.InbEnfCap, row.OutbEnfCap, row.FilterRawInbCap,
		row.FilterRawOutbCap, row.Enp0, row.McastFdbTop,
	) {
		return nil
	}

	return &domain.Switch{
		ID:                   uuid.Nil,
		NodeID:               row.ID,
		NodeGUID:             row.NodeGuid,
		LinearFDBCap:         row.LinearFdbCap,
		RandomFDBCap:         row.RandomFdbCap,
		MCastFDBCap:          row.McastFdbCap,
		LinearFDBTop:         row.LinearFdbTop,
		DefPort:              row.DefPort,
		DefMCastPriPort:      row.DefMcastPriPort,
		DefMCastNotPriPort:   row.DefMcastNotPriPort,
		LifeTimeValue:        row.LifeTimeValue,
		PortStateChange:      row.PortStateChange,
		OptimizedSLVLMapping: row.OptimizedSlvlMapping,
		LidsPerPort:          row.LidsPerPort,
		PartEnfCap:           row.PartEnfCap,
		InbEnfCap:            row.InbEnfCap,
		OutbEnfCap:           row.OutbEnfCap,
		FilterRawInbCap:      row.FilterRawInbCap,
		FilterRawOutbCap:     row.FilterRawOutbCap,
		ENP0:                 row.Enp0,
		MCastFDBTop:          row.McastFdbTop,
	}
}

// allNil returns true when all provided pointer arguments are nil. It's a small
// helper used by mapNodeInfo/mapSwitchInfo to decide whether enrichment data
// exists in the SQL row.
func allNil[T any](ptrs ...*T) bool {
	for _, ptr := range ptrs {
		if ptr != nil {
			return false
		}
	}
	return true
}
