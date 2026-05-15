package parser

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/domain"
	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/ports"
)

type mockParserWriter struct {
	createLogFn       func(ctx context.Context, status domain.LogStatus) (*domain.Log, error)
	updateLogStatusFn func(ctx context.Context, id uuid.UUID, status domain.LogStatus) error
	updateLogCountsFn func(ctx context.Context, id uuid.UUID, nodesCount, portsCount int32) error
	insertNodesFn     func(ctx context.Context, batch []domain.Node) error
	insertPortsFn     func(ctx context.Context, batch []domain.Port) error
	upsertSwitchesFn  func(ctx context.Context, batch []domain.Switch) error
	upsertNodesInfoFn func(ctx context.Context, batch []domain.NodesInfo) error
	deleteLogDataFn   func(ctx context.Context, logID uuid.UUID) error
}

func (m *mockParserWriter) CreateLog(ctx context.Context, status domain.LogStatus) (*domain.Log, error) {
	if m.createLogFn != nil {
		return m.createLogFn(ctx, status)
	}
	return &domain.Log{}, nil
}

func (m *mockParserWriter) UpdateLogStatus(ctx context.Context, id uuid.UUID, status domain.LogStatus) error {
	if m.updateLogStatusFn != nil {
		return m.updateLogStatusFn(ctx, id, status)
	}
	return nil
}

func (m *mockParserWriter) UpdateLogCounts(ctx context.Context, id uuid.UUID, nodesCount, portsCount int32) error {
	if m.updateLogCountsFn != nil {
		return m.updateLogCountsFn(ctx, id, nodesCount, portsCount)
	}
	return nil
}

func (m *mockParserWriter) InsertNodes(ctx context.Context, batch []domain.Node) error {
	if m.insertNodesFn != nil {
		return m.insertNodesFn(ctx, batch)
	}
	return nil
}

func (m *mockParserWriter) InsertPorts(ctx context.Context, batch []domain.Port) error {
	if m.insertPortsFn != nil {
		return m.insertPortsFn(ctx, batch)
	}
	return nil
}

func (m *mockParserWriter) UpsertSwitches(ctx context.Context, batch []domain.Switch) error {
	if m.upsertSwitchesFn != nil {
		return m.upsertSwitchesFn(ctx, batch)
	}
	return nil
}

func (m *mockParserWriter) UpsertNodesInfo(ctx context.Context, batch []domain.NodesInfo) error {
	if m.upsertNodesInfoFn != nil {
		return m.upsertNodesInfoFn(ctx, batch)
	}
	return nil
}

func (m *mockParserWriter) DeleteLogData(ctx context.Context, logID uuid.UUID) error {
	if m.deleteLogDataFn != nil {
		return m.deleteLogDataFn(ctx, logID)
	}
	return nil
}

var _ ports.ParserWriter = (*mockParserWriter)(nil)

func newTestParser(repo ports.ParserWriter) *Parser {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return &Parser{
		repo:      repo,
		batchSize: 100,
		logger:    logger,
	}
}

func TestFlushNodes_EmptyBatch(t *testing.T) {
	mock := &mockParserWriter{
		insertNodesFn: func(_ context.Context, _ []domain.Node) error {
			t.Error("InsertNodes should not be called for empty batch")
			return nil
		},
	}
	p := newTestParser(mock)
	var batch []domain.Node

	err := p.flushNodes(context.Background(), &batch)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if len(batch) != 0 {
		t.Errorf("expected batch length 0, got %d", len(batch))
	}
}

func TestFlushNodes_Success(t *testing.T) {
	logID := uuid.New()
	nodeID := uuid.New()
	var insertedNodes []domain.Node

	mock := &mockParserWriter{
		insertNodesFn: func(_ context.Context, batch []domain.Node) error {
			insertedNodes = append(insertedNodes, batch...)
			return nil
		},
	}
	p := newTestParser(mock)
	batch := []domain.Node{
		{ID: nodeID, LogID: logID, NodeDesc: "Switch-1", NodeType: domain.NodeSwitch},
	}

	err := p.flushNodes(context.Background(), &batch)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if len(batch) != 0 {
		t.Errorf("expected batch to be cleared, got length %d", len(batch))
	}
	if len(insertedNodes) != 1 {
		t.Errorf("expected 1 node to be inserted, got %d", len(insertedNodes))
	}
	if insertedNodes[0].ID != nodeID {
		t.Errorf("expected node ID %v, got %v", nodeID, insertedNodes[0].ID)
	}
}

func TestFlushNodes_RepositoryError(t *testing.T) {
	mock := &mockParserWriter{
		insertNodesFn: func(_ context.Context, _ []domain.Node) error {
			return errors.New("database error")
		},
	}
	p := newTestParser(mock)
	nodeID := uuid.New()
	logID := uuid.New()
	batch := []domain.Node{
		{ID: nodeID, LogID: logID, NodeDesc: "Switch-1", NodeType: domain.NodeSwitch},
	}

	err := p.flushNodes(context.Background(), &batch)

	if err == nil {
		t.Error("expected error from repository, got nil")
	}
	if len(batch) != 1 {
		t.Errorf("expected batch to remain unchanged after error, got length %d", len(batch))
	}
}

func TestFlushPorts_EmptyBatch(t *testing.T) {
	mock := &mockParserWriter{
		insertPortsFn: func(_ context.Context, _ []domain.Port) error {
			t.Error("InsertPorts should not be called for empty batch")
			return nil
		},
	}
	p := newTestParser(mock)
	var batch []domain.Port

	err := p.flushPorts(context.Background(), &batch)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if len(batch) != 0 {
		t.Errorf("expected batch length 0, got %d", len(batch))
	}
}

func TestFlushPorts_Success(t *testing.T) {
	nodeID := uuid.New()
	portID := uuid.New()
	var insertedPorts []domain.Port

	mock := &mockParserWriter{
		insertPortsFn: func(_ context.Context, batch []domain.Port) error {
			insertedPorts = append(insertedPorts, batch...)
			return nil
		},
	}
	p := newTestParser(mock)
	batch := []domain.Port{
		{ID: portID, NodeID: nodeID, PortNum: 1, PortGUID: "0x123456"},
	}

	err := p.flushPorts(context.Background(), &batch)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if len(batch) != 0 {
		t.Errorf("expected batch to be cleared, got length %d", len(batch))
	}
	if len(insertedPorts) != 1 {
		t.Errorf("expected 1 port to be inserted, got %d", len(insertedPorts))
	}
}

func TestFlushPorts_RepositoryError(t *testing.T) {
	mock := &mockParserWriter{
		insertPortsFn: func(_ context.Context, _ []domain.Port) error {
			return errors.New("database error")
		},
	}
	p := newTestParser(mock)
	nodeID := uuid.New()
	portID := uuid.New()
	batch := []domain.Port{
		{ID: portID, NodeID: nodeID, PortNum: 1, PortGUID: "0x123456"},
	}

	err := p.flushPorts(context.Background(), &batch)

	if err == nil {
		t.Error("expected error from repository, got nil")
	}
	if len(batch) != 1 {
		t.Errorf("expected batch to remain unchanged after error, got length %d", len(batch))
	}
}

func TestFlushSwitches_EmptyBatch(t *testing.T) {
	mock := &mockParserWriter{
		upsertSwitchesFn: func(_ context.Context, _ []domain.Switch) error {
			t.Error("UpsertSwitches should not be called for empty batch")
			return nil
		},
	}
	p := newTestParser(mock)
	var batch []domain.Switch

	err := p.flushSwitches(context.Background(), &batch)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if len(batch) != 0 {
		t.Errorf("expected batch length 0, got %d", len(batch))
	}
}

func TestFlushSwitches_Success(t *testing.T) {
	nodeID := uuid.New()
	nodeGUID := "0x123456"
	var upsertedSwitches []domain.Switch

	mock := &mockParserWriter{
		upsertSwitchesFn: func(_ context.Context, batch []domain.Switch) error {
			upsertedSwitches = append(upsertedSwitches, batch...)
			return nil
		},
	}
	p := newTestParser(mock)
	batch := []domain.Switch{
		{NodeID: nodeID, NodeGUID: nodeGUID, LinearFDBCap: new(int32(100))},
	}

	err := p.flushSwitches(context.Background(), &batch)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if len(batch) != 0 {
		t.Errorf("expected batch to be cleared, got length %d", len(batch))
	}
	if len(upsertedSwitches) != 1 {
		t.Errorf("expected 1 switch to be upserted, got %d", len(upsertedSwitches))
	}
}

func TestFlushSwitches_RepositoryError(t *testing.T) {
	mock := &mockParserWriter{
		upsertSwitchesFn: func(_ context.Context, _ []domain.Switch) error {
			return errors.New("database error")
		},
	}
	p := newTestParser(mock)
	nodeID := uuid.New()
	nodeGUID := "0x123456"
	batch := []domain.Switch{
		{NodeID: nodeID, NodeGUID: nodeGUID, LinearFDBCap: new(int32(100))},
	}

	err := p.flushSwitches(context.Background(), &batch)

	if err == nil {
		t.Error("expected error from repository, got nil")
	}
	if len(batch) != 1 {
		t.Errorf("expected batch to remain unchanged after error, got length %d", len(batch))
	}
}

func TestFlushNodesInfo_EmptyBatch(t *testing.T) {
	mock := &mockParserWriter{
		upsertNodesInfoFn: func(_ context.Context, _ []domain.NodesInfo) error {
			t.Error("UpsertNodesInfo should not be called for empty batch")
			return nil
		},
	}
	p := newTestParser(mock)
	var batch []domain.NodesInfo

	err := p.flushNodesInfo(context.Background(), &batch)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if len(batch) != 0 {
		t.Errorf("expected batch length 0, got %d", len(batch))
	}
}

func TestFlushNodesInfo_Success(t *testing.T) {
	nodeID := uuid.New()
	var upsertedNodesInfo []domain.NodesInfo

	mock := &mockParserWriter{
		upsertNodesInfoFn: func(_ context.Context, batch []domain.NodesInfo) error {
			upsertedNodesInfo = append(upsertedNodesInfo, batch...)
			return nil
		},
	}
	p := newTestParser(mock)
	batch := []domain.NodesInfo{
		{NodeID: nodeID, ProductName: new("Mellanox")},
	}

	err := p.flushNodesInfo(context.Background(), &batch)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if len(batch) != 0 {
		t.Errorf("expected batch to be cleared, got length %d", len(batch))
	}
	if len(upsertedNodesInfo) != 1 {
		t.Errorf("expected 1 nodes info to be upserted, got %d", len(upsertedNodesInfo))
	}
}

func TestFlushNodesInfo_RepositoryError(t *testing.T) {
	mock := &mockParserWriter{
		upsertNodesInfoFn: func(_ context.Context, _ []domain.NodesInfo) error {
			return errors.New("database error")
		},
	}
	p := newTestParser(mock)
	nodeID := uuid.New()
	batch := []domain.NodesInfo{
		{NodeID: nodeID, ProductName: new("Mellanox")},
	}

	err := p.flushNodesInfo(context.Background(), &batch)

	if err == nil {
		t.Error("expected error from repository, got nil")
	}
	if len(batch) != 1 {
		t.Errorf("expected batch to remain unchanged after error, got length %d", len(batch))
	}
}
