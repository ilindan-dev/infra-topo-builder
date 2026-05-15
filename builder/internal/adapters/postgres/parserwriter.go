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

var _ ports.ParserWriter = (*parserWriter)(nil)

// parserWriter implements the ports.ParserWriter interface for persisting parsed data
// into a PostgreSQL database. It is optimized for bulk operations (e.g., COPY, Batch)
// and ensures idempotency for repeated imports. It uses sqlc-generated queries for
// type-safe access and logs all mutating operations for traceability.
// parserWriter implements the ports.ParserWriter interface using a Postgres
// connection pool and sqlc-generated queries. It prefers bulk-friendly
// operations (batch/COPY) to maximize import throughput and keeps error
// translation to domain-level errors localized via translateError.
type parserWriter struct {
	pool    *pgxpool.Pool
	queries *gen.Queries
	logger  *slog.Logger
}

// NewParserWriter constructs a ParserWriter backed by Postgres. The returned
// implementation is optimized for high-throughput imports and will annotate
// the provided logger with adapter-specific fields.
func NewParserWriter(pool *pgxpool.Pool, logger *slog.Logger) ports.ParserWriter {
	return &parserWriter{
		pool:    pool,
		queries: gen.New(pool),
		logger:  logger.With("adapter", "postgres", "implement", "parserWriter"),
	}
}

// CreateLog creates a new parsing session record and returns the created
// domain.Log. The returned domain.Log contains the generated ID and initial
// counters. Errors are translated to canonical domain errors.
func (r *parserWriter) CreateLog(ctx context.Context, status domain.LogStatus) (*domain.Log, error) {
	log, err := r.queries.CreateLog(ctx, status)
	return new(domain.Log{
		ID:         log.ID,
		Status:     log.Status,
		NodesCount: log.NodesCount,
		PortsCount: log.PortsCount,
		CreatedAt:  log.CreatedAt.Time,
	}), translateError(err, "CreateLog", r.logger)
}

// UpdateLogStatus updates only the status field for the given log ID.
// Designed to be a lightweight write used to mark progress or final state.
func (r *parserWriter) UpdateLogStatus(ctx context.Context, id uuid.UUID, status domain.LogStatus) error {
	return translateError(r.queries.UpdateLogStatus(ctx, gen.UpdateLogStatusParams{
		ID:     id,
		Status: status,
	}), "UpdateLogStatus", r.logger)
}

// UpdateLogCounts persists aggregated counters (nodes and ports) for a log.
// Call this after a successful import to record summary metrics.
func (r *parserWriter) UpdateLogCounts(ctx context.Context, id uuid.UUID, nodesCount, portsCount int32) error {
	return translateError(r.queries.UpdateLogCounts(ctx, gen.UpdateLogCountsParams{
		NodesCount: nodesCount,
		PortsCount: portsCount,
		ID:         id,
	}), "UpdateLogCounts", r.logger)
}

// InsertNodes performs a bulk insert of node records. It converts each
// domain.Node into the sqlc-generated InsertNodesParams type. A deterministic
// ID is derived from (logID, nodeGUID) so repeated imports of the same source
// produce stable record identifiers. Any underlying SQL error is translated
// into a domain-level error.
func (r *parserWriter) InsertNodes(ctx context.Context, nodes []domain.Node) error {
	genNodes := make([]gen.InsertNodesParams, 0, len(nodes))
	for i := range nodes {
		node := &nodes[i]
		genNodes = append(genNodes, gen.InsertNodesParams{
			ID:              node.ID,
			LogID:           node.LogID,
			NodeDesc:        node.NodeDesc,
			NumPorts:        node.NumPorts,
			NodeType:        node.NodeType,
			ClassVersion:    node.ClassVersion,
			BaseVersion:     node.BaseVersion,
			SystemImageGuid: node.SystemImageGUID,
			NodeGuid:        node.NodeGUID,
			PortGuid:        node.PortGUID,
		})
	}
	_, err := r.queries.InsertNodes(ctx, genNodes)
	return translateError(err, "InsertNodes", r.logger)
}

// InsertPorts performs a bulk insert of port telemetry and configuration
// records. Fields are mapped from the domain.Port into the sqlc-generated
// InsertPortsParams. This function is optimized for bulk import and will
// propagate the first error encountered as a domain-level error.
func (r *parserWriter) InsertPorts(ctx context.Context, portItems []domain.Port) error {
	genPorts := make([]gen.InsertPortsParams, 0, len(portItems))
	for i := range portItems {
		port := &portItems[i]
		genPorts = append(genPorts, gen.InsertPortsParams{
			NodeID:                              port.NodeID,
			NodeGuid:                            port.NodeGUID,
			PortGuid:                            port.PortGUID,
			PortNum:                             port.PortNum,
			MKey:                                port.MKey,
			GidPrfx:                             port.GidPrfx,
			MsmLid:                              port.MsmLid,
			Lid:                                 port.Lid,
			CapMsk:                              port.CapMsk,
			MKeyLeasePeriod:                     port.MKeyLeasePeriod,
			DiagCode:                            port.DiagCode,
			LinkWidthActv:                       port.LinkWidthActv,
			LinkWidthSup:                        port.LinkWidthSup,
			LinkWidthEn:                         port.LinkWidthEn,
			LocalPortNum:                        port.LocalPortNum,
			LinkSpeedEn:                         port.LinkSpeedEn,
			LinkSpeedActv:                       port.LinkSpeedActv,
			Lmc:                                 port.Lmc,
			MKeyProtBits:                        port.MKeyProtBits,
			LinkDownDefState:                    port.LinkDownDefState,
			PortPhyState:                        port.PortPhyState,
			PortState:                           port.PortState,
			LinkSpeedSup:                        port.LinkSpeedSup,
			VlArbHighCap:                        port.VlArbHighCap,
			VlHighLimit:                         port.VlHighLimit,
			InitType:                            port.InitType,
			VlCap:                               port.VlCap,
			Msmsl:                               port.Msmsl,
			Nmtu:                                port.Nmtu,
			FilterRawOutb:                       port.FilterRawOutb,
			FilterRawInb:                        port.FilterRawInb,
			PartEnfOutb:                         port.PartEnfOutb,
			PartEnfInb:                          port.PartEnfInb,
			OpVls:                               port.OpVls,
			HoqLife:                             port.HoqLife,
			VlStallCnt:                          port.VlStallCnt,
			MtuCap:                              port.MtuCap,
			InitTypeReply:                       port.InitTypeReply,
			VlArbLowCap:                         port.VlArbLowCap,
			PKeyViolations:                      port.PKeyViolations,
			MKeyViolations:                      port.MKeyViolations,
			SubnTmo:                             port.SubnTmo,
			MulticastPKeyTrapSuppressionEnabled: port.MulticastPKeyTrapSuppressionEnabled,
			ClientReregister:                    port.ClientReregister,
			GuidCap:                             port.GUIDCap,
			QKeyViolations:                      port.QKeyViolations,
			MaxCreditHint:                       port.MaxCreditHint,
			OverrunErrs:                         port.OverrunErrs,
			LocalPhyError:                       port.LocalPhyError,
			RespTimeValue:                       port.RespTimeValue,
			LinkRoundTripLatency:                port.LinkRoundTripLatency,
			OooSlMask:                           port.OooSlMask,
			CapMsk2:                             port.CapMsk2,
			FecActv:                             port.FecActv,
			RetransActv:                         port.RetransActv,
		})
	}
	_, err := r.queries.InsertPorts(ctx, genPorts)
	return translateError(err, "InsertPorts", r.logger)
}

// UpsertNodesInfo performs idempotent upserts of node metadata (nodes_info).
// The function uses sqlc's batched helper which must be explicitly Exec'd to
// flush the queued statements. Per-item errors are logged; the first error
// is returned after translation to a domain error so callers can react.
func (r *parserWriter) UpsertNodesInfo(ctx context.Context, info []domain.NodesInfo) error {
	genInfo := make([]gen.UpsertNodesInfoParams, 0, len(info))
	for i := range info {
		nodeInfo := &info[i]
		genInfo = append(genInfo, gen.UpsertNodesInfoParams{
			NodeID:                 nodeInfo.NodeID,
			SerialNumber:           nodeInfo.SerialNumber,
			PartNumber:             nodeInfo.PartNumber,
			Revision:               nodeInfo.Revision,
			ProductName:            nodeInfo.ProductName,
			Endianness:             nodeInfo.Endianness,
			EnableEndiannessPerJob: nodeInfo.EnableEndiannessPerJob,
			ReproducibilityDisable: nodeInfo.ReproducibilityDisable,
		})
	}
	result := r.queries.UpsertNodesInfo(ctx, genInfo)

	var firstErr error
	result.Exec(func(i int, err error) {
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			r.logger.Error("upsert nodes_info batch item failed", "index", i, "error", err)
		}
	})
	if firstErr != nil {
		return translateError(firstErr, "UpsertNodesInfo", r.logger)
	}
	return nil
}

// UpsertSwitches performs idempotent upserts of switch capability records.
// Uses sqlc's batched helper and logs per-item failures. The first encountered
// error (if any) will be returned after translation to a domain error.
func (r *parserWriter) UpsertSwitches(ctx context.Context, switches []domain.Switch) error {
	genSwitches := make([]gen.UpsertSwitchesParams, 0, len(switches))
	for i := range switches {
		sw := &switches[i]
		genSwitches = append(genSwitches, gen.UpsertSwitchesParams{
			NodeID:               sw.NodeID,
			NodeGuid:             sw.NodeGUID,
			LinearFdbCap:         sw.LinearFDBCap,
			RandomFdbCap:         sw.RandomFDBCap,
			McastFdbCap:          sw.MCastFDBCap,
			LinearFdbTop:         sw.LinearFDBTop,
			DefPort:              sw.DefPort,
			DefMcastPriPort:      sw.DefMCastPriPort,
			DefMcastNotPriPort:   sw.DefMCastNotPriPort,
			LifeTimeValue:        sw.LifeTimeValue,
			PortStateChange:      sw.PortStateChange,
			OptimizedSlvlMapping: sw.OptimizedSLVLMapping,
			LidsPerPort:          sw.LidsPerPort,
			PartEnfCap:           sw.PartEnfCap,
			InbEnfCap:            sw.InbEnfCap,
			OutbEnfCap:           sw.OutbEnfCap,
			FilterRawInbCap:      sw.FilterRawInbCap,
			FilterRawOutbCap:     sw.FilterRawOutbCap,
			Enp0:                 sw.ENP0,
			McastFdbTop:          sw.MCastFDBTop,
		})
	}
	result := r.queries.UpsertSwitches(ctx, genSwitches)

	var firstErr error
	result.Exec(func(i int, err error) {
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			r.logger.Error("upsert switches batch item failed", "index", i, "error", err)
		}
	})
	if firstErr != nil {
		return translateError(firstErr, "UpsertSwitches", r.logger)
	}
	return nil
}

// DeleteLogData deletes all persistent records associated with the given
// parsing session (logID). Typical deletions include nodes, ports, nodes_info,
// switch capability rows and any other rows that reference the log. The SQL
// implementation is expected to run the deletions in a single transaction and
// be safe to call multiple times (idempotent); calling this method when no
// rows exist for the provided logID should succeed without error. Any
// low-level database error is translated into a canonical domain error via
// translateError so callers can handle domain-level semantics.
func (r *parserWriter) DeleteLogData(ctx context.Context, logID uuid.UUID) error {
	err := r.queries.DeleteLogData(ctx, logID)
	return translateError(err, "DeleteLogData", r.logger)
}
