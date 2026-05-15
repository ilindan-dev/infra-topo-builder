package parser

import (
	"bufio"
	"context"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/domain"
	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/utils"
)

// ParseSharpAnInfo parses the .sharp_an_info file produced by ibdiagnet.
// The format contains SW_GUID=<guid> headers followed by key=value properties.
// Properties are grouped under the most recent SW_GUID header and converted
// into domain.NodesInfo records. Records are collected and flushed in batches
// according to the parser's batchSize to avoid high memory usage.
func (p *Parser) ParseSharpAnInfo(ctx context.Context, logID uuid.UUID, reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	batch := make([]domain.NodesInfo, 0, p.batchSize)

	var currentInfo *domain.NodesInfo

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "-") {
			continue
		}

		if strings.HasPrefix(line, "SW_GUID=") {
			if err := p.handleNewSwitchContext(ctx, logID, line, &currentInfo, &batch); err != nil {
				return err
			}
			continue
		}

		p.applySharpInfoProperty(currentInfo, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if currentInfo != nil {
		batch = append(batch, *currentInfo)
	}

	return p.flushNodesInfo(ctx, &batch)
}

// handleNewSwitchContext begins a new NodesInfo context when a SW_GUID= header
// is encountered. The function appends any existing currentInfo to the provided
// batch and triggers a flush when the batch reaches the parser's configured
// batchSize. The incoming line is expected to be of the form "SW_GUID=...";
// the GUID is normalized to start with "0x" (if necessary) and is used with
// the logID to produce a deterministic NodeID via
// utils.GenerateDeterministicUUID. The currentInfo pointer is set to a new
// domain.NodesInfo initialized with that NodeID. Returns an error if the
// underlying flush operation fails.
func (p *Parser) handleNewSwitchContext(ctx context.Context, logID uuid.UUID, line string,
	currentInfo **domain.NodesInfo, batch *[]domain.NodesInfo,
) error {
	if *currentInfo != nil {
		*batch = append(*batch, **currentInfo)
		if len(*batch) >= p.batchSize {
			if err := p.flushNodesInfo(ctx, batch); err != nil {
				return err
			}
		}
	}

	rawGUID := strings.TrimPrefix(line, "SW_GUID=")
	nodeGUID := cleanGUID(rawGUID)

	*currentInfo = &domain.NodesInfo{
		NodeID: utils.GenerateDeterministicUUID(logID, nodeGUID),
	}
	return nil
}

// applySharpInfoProperty applies a single key=value property line to the
// provided NodesInfo object. Unknown keys are ignored. The function is
// tolerant to malformed lines and is a no-op when info is nil.
func (p *Parser) applySharpInfoProperty(info *domain.NodesInfo, line string) {
	if info == nil {
		return
	}

	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return
	}

	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])

	switch key {
	case "endianness":
		info.Endianness = utils.ParsePtrInt32(val)
	case "enable_endianness_per_job":
		info.EnableEndiannessPerJob = utils.ParsePtrInt32(val)
	case "reproducibility_disable":
		info.ReproducibilityDisable = utils.ParsePtrInt32(val)
	}
}
