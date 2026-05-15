package parser

import (
	"context"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/domain"
)

// flushNodes persists the provided node batch using the repository and
// resets the slice to zero length. No-op when batch is empty.
func (p *Parser) flushNodes(ctx context.Context, batch *[]domain.Node) error {
	if len(*batch) == 0 {
		return nil
	}
	if err := p.repo.InsertNodes(ctx, *batch); err != nil {
		return err
	}
	*batch = (*batch)[:0]
	return nil
}

// flushPorts persists the provided port batch using the repository and
// resets the slice to zero length. No-op when batch is empty.
func (p *Parser) flushPorts(ctx context.Context, batch *[]domain.Port) error {
	if len(*batch) == 0 {
		return nil
	}
	if err := p.repo.InsertPorts(ctx, *batch); err != nil {
		return err
	}
	*batch = (*batch)[:0]
	return nil
}

// flushSwitches performs idempotent upserts for switch capability records
// and clears the input slice on success.
func (p *Parser) flushSwitches(ctx context.Context, batch *[]domain.Switch) error {
	if len(*batch) == 0 {
		return nil
	}
	if err := p.repo.UpsertSwitches(ctx, *batch); err != nil {
		return err
	}
	*batch = (*batch)[:0]
	return nil
}

// flushNodesInfo upserts node metadata records (nodes_info) in an
// idempotent fashion and clears the provided slice when successful.
func (p *Parser) flushNodesInfo(ctx context.Context, batch *[]domain.NodesInfo) error {
	if len(*batch) == 0 {
		return nil
	}
	if err := p.repo.UpsertNodesInfo(ctx, *batch); err != nil {
		return err
	}
	*batch = (*batch)[:0]
	return nil
}
