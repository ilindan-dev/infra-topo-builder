package parser

import (
	"bufio"
	"context"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/domain"
)

// parserState represents which section of the CSV file the scanner is currently in.
// Sections include nodes, ports, switches and system info blocks.
type parserState int

const (
	stateIdle parserState = iota
	stateNodes
	statePorts
	stateSwitches
	stateSystemInfo
)

// ParseDBCSV parses a database-style CSV export from ibdiagnet. It streams the
// input, recognizes START_/END_ section markers, accumulates rows into in-memory
// batches, and flushes them to the repository using batch-friendly operations.
// Context cancellation is honored to stop long-running parsing jobs.
func (p *Parser) ParseDBCSV(ctx context.Context, logID uuid.UUID, reader io.Reader) error {
	log := p.logger.With("method", "ParseDBCSV")
	scanner := bufio.NewScanner(reader)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	state := stateIdle
	nodesBatch := make([]domain.Node, 0, p.batchSize)
	portsBatch := make([]domain.Port, 0, p.batchSize)
	switchesBatch := make([]domain.Switch, 0, p.batchSize)
	systemInfoBatch := make([]domain.NodesInfo, 0, p.batchSize)

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			log.Warn("context canceled during file parsing")
			return err
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" || isHeaderLine(line) {
			continue
		}

		isTransition, err := p.processTransition(ctx, line, &state, &nodesBatch,
			&portsBatch, &switchesBatch, &systemInfoBatch)
		if err != nil {
			log.Error("failed processing transition", "error", err)
			return err
		}
		if isTransition {
			continue
		}

		if err := p.processData(ctx, logID, line, state, &nodesBatch, &portsBatch,
			&switchesBatch, &systemInfoBatch); err != nil {
			log.Error("failed processing data line", "error", err)
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error("error scanning file", "error", err)
		return err
	}

	return p.flushAll(ctx, &nodesBatch, &portsBatch, &switchesBatch, &systemInfoBatch)
}

// isHeaderLine detects common CSV header lines so they can be skipped when streaming.
// Currently recognizes headers beginning with "NodeDesc," or "NodeGuid,".
func isHeaderLine(line string) bool {
	lower := strings.ToLower(line)
	return strings.HasPrefix(lower, "nodedesc,") ||
		strings.HasPrefix(lower, "nodeguid,") ||
		strings.HasPrefix(lower, "portguid,")
}

// processTransition inspects a scanner line to determine whether it is a
// section marker (START_* / END_*). When an END_* marker is seen the
// corresponding batch is flushed. Returns (true, nil) when a transition was
// recognized so callers skip normal data processing for that line.
func (p *Parser) processTransition(ctx context.Context, line string,
	state *parserState, nodes *[]domain.Node, ports *[]domain.Port,
	switches *[]domain.Switch, sysInfo *[]domain.NodesInfo,
) (bool, error) {
	switch {
	case strings.HasPrefix(line, "START_NODES"):
		*state = stateNodes
	case strings.HasPrefix(line, "END_NODES"):
		*state = stateIdle
		return true, p.flushNodes(ctx, nodes)

	case strings.HasPrefix(line, "START_PORTS"):
		*state = statePorts
	case strings.HasPrefix(line, "END_PORTS"):
		*state = stateIdle
		return true, p.flushPorts(ctx, ports)

	case strings.HasPrefix(line, "START_SWITCHES"):
		*state = stateSwitches
	case strings.HasPrefix(line, "END_SWITCHES"):
		*state = stateIdle
		return true, p.flushSwitches(ctx, switches)

	case strings.HasPrefix(line, "START_SYSTEM_GENERAL_INFORMATION"):
		*state = stateSystemInfo
	case strings.HasPrefix(line, "END_SYSTEM_GENERAL_INFORMATION"):
		*state = stateIdle
		return true, p.flushNodesInfo(ctx, sysInfo)

	default:
		return false, nil
	}
	return true, nil
}

// processData dispatches a non-transition CSV line to the appropriate
// append helper based on the current parser state. Each append helper may
// trigger a flush when its batch reaches the configured size.
func (p *Parser) processData(ctx context.Context, logID uuid.UUID, line string,
	state parserState, nodes *[]domain.Node, ports *[]domain.Port,
	switches *[]domain.Switch, sysInfo *[]domain.NodesInfo,
) error {
	switch state {
	case stateNodes:
		return p.appendNode(ctx, logID, line, nodes)
	case statePorts:
		return p.appendPort(ctx, logID, line, ports)
	case stateSwitches:
		return p.appendSwitch(ctx, logID, line, switches)
	case stateSystemInfo:
		return p.appendSysInfo(ctx, logID, line, sysInfo)
	default:
		return nil
	}
}

// appendNode parses a node CSV line and appends it to the provided batch.
// If the batch reaches the configured batchSize, it will be flushed to the repository.
func (p *Parser) appendNode(ctx context.Context, logID uuid.UUID, line string, batch *[]domain.Node) error {
	node, err := parseNodeLine(logID, line)
	if err != nil {
		return err
	}
	*batch = append(*batch, node)
	if len(*batch) >= p.batchSize {
		return p.flushNodes(ctx, batch)
	}
	return nil
}

// appendPort parses a port CSV line and appends it to the provided batch.
// Flushes the batch when the configured size is reached.
func (p *Parser) appendPort(ctx context.Context, logID uuid.UUID, line string, batch *[]domain.Port) error {
	port, err := parsePortLine(logID, line)
	if err != nil {
		return err
	}
	*batch = append(*batch, port)
	if len(*batch) >= p.batchSize {
		return p.flushPorts(ctx, batch)
	}
	return nil
}

// appendSwitch parses a switch CSV line and appends it to the provided batch.
// Flushes the batch when the configured batchSize is reached.
func (p *Parser) appendSwitch(ctx context.Context, logID uuid.UUID, line string, batch *[]domain.Switch) error {
	sw, err := parseSwitchLine(logID, line)
	if err != nil {
		return err
	}
	*batch = append(*batch, sw)
	if len(*batch) >= p.batchSize {
		return p.flushSwitches(ctx, batch)
	}
	return nil
}

// appendSysInfo parses a system info line (from .sharp_an_info parsing) and
// appends the resulting NodesInfo into the batch. Flushes when necessary.
func (p *Parser) appendSysInfo(ctx context.Context, logID uuid.UUID, line string, batch *[]domain.NodesInfo) error {
	info, err := parseSystemInfoLine(logID, line)
	if err != nil {
		return err
	}
	*batch = append(*batch, info)
	if len(*batch) >= p.batchSize {
		return p.flushNodesInfo(ctx, batch)
	}
	return nil
}

// flushAll flushes any remaining batches for nodes, ports, switches and
// nodes_info. It is intended to be called once parsing is complete to ensure
// all buffered records are persisted.
func (p *Parser) flushAll(ctx context.Context, nodes *[]domain.Node, ports *[]domain.Port,
	switches *[]domain.Switch, sysInfo *[]domain.NodesInfo,
) error {
	if err := p.flushNodes(ctx, nodes); err != nil {
		return err
	}
	if err := p.flushPorts(ctx, ports); err != nil {
		return err
	}
	if err := p.flushSwitches(ctx, switches); err != nil {
		return err
	}
	return p.flushNodesInfo(ctx, sysInfo)
}
