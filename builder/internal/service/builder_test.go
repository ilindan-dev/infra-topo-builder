package service

import (
	"archive/zip"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/domain"
)

// --- MOCKS ---

type mockParserWriter struct {
	mockCreateLog       func(ctx context.Context, status domain.LogStatus) (*domain.Log, error)
	mockUpdateLogStatus func(ctx context.Context, logID uuid.UUID, status domain.LogStatus) error
	mockUpdateLogCounts func(ctx context.Context, id uuid.UUID, nodesCount, portsCount int32) error
	mockInsertNodes     func(ctx context.Context, nodes []domain.Node) error
	mockInsertPorts     func(ctx context.Context, ports []domain.Port) error
	mockUpsertNodesInfo func(ctx context.Context, info []domain.NodesInfo) error
	mockUpsertSwitches  func(ctx context.Context, switches []domain.Switch) error
	mockDeleteLogData   func(ctx context.Context, logID uuid.UUID) error
}

func (m *mockParserWriter) CreateLog(ctx context.Context, status domain.LogStatus) (*domain.Log, error) {
	if m.mockCreateLog != nil {
		return m.mockCreateLog(ctx, status)
	}
	return &domain.Log{ID: uuid.New(), Status: status}, nil
}

func (m *mockParserWriter) UpdateLogStatus(ctx context.Context, logID uuid.UUID, status domain.LogStatus) error {
	if m.mockUpdateLogStatus != nil {
		return m.mockUpdateLogStatus(ctx, logID, status)
	}
	return nil
}

func (m *mockParserWriter) UpdateLogCounts(ctx context.Context, id uuid.UUID, nodesCount, portsCount int32) error {
	if m.mockUpdateLogCounts != nil {
		return m.mockUpdateLogCounts(ctx, id, nodesCount, portsCount)
	}
	return nil
}

func (m *mockParserWriter) InsertNodes(ctx context.Context, nodes []domain.Node) error {
	if m.mockInsertNodes != nil {
		return m.mockInsertNodes(ctx, nodes)
	}
	return nil
}

func (m *mockParserWriter) InsertPorts(ctx context.Context, ports []domain.Port) error {
	if m.mockInsertPorts != nil {
		return m.mockInsertPorts(ctx, ports)
	}
	return nil
}

func (m *mockParserWriter) UpsertNodesInfo(ctx context.Context, info []domain.NodesInfo) error {
	if m.mockUpsertNodesInfo != nil {
		return m.mockUpsertNodesInfo(ctx, info)
	}
	return nil
}

func (m *mockParserWriter) UpsertSwitches(ctx context.Context, switches []domain.Switch) error {
	if m.mockUpsertSwitches != nil {
		return m.mockUpsertSwitches(ctx, switches)
	}
	return nil
}

func (m *mockParserWriter) DeleteLogData(ctx context.Context, logID uuid.UUID) error {
	if m.mockDeleteLogData != nil {
		return m.mockDeleteLogData(ctx, logID)
	}
	return nil
}

type mockTopologyParser struct {
	mockParseDBCSV       func(ctx context.Context, logID uuid.UUID, r io.Reader) error
	mockParseSharpAnInfo func(ctx context.Context, logID uuid.UUID, r io.Reader) error
}

func (m *mockTopologyParser) ParseDBCSV(ctx context.Context, logID uuid.UUID, r io.Reader) error {
	if m.mockParseDBCSV != nil {
		return m.mockParseDBCSV(ctx, logID, r)
	}
	return nil
}

func (m *mockTopologyParser) ParseSharpAnInfo(ctx context.Context, logID uuid.UUID, r io.Reader) error {
	if m.mockParseSharpAnInfo != nil {
		return m.mockParseSharpAnInfo(ctx, logID, r)
	}
	return nil
}

// --- HELPER ---
func createTestZip(t *testing.T, files map[string]string) string {
	t.Helper()
	f, err := os.CreateTemp("", "test_archive_*.zip")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatalf("failed to create zip entry: %v", err)
		}
		_, _ = fw.Write([]byte(content))
	}
	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	return f.Name()
}

// --- TESTS ---

func TestBuildFromArchive_Validation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &mockParserWriter{}
	parser := &mockTopologyParser{}
	svc := NewBuilder(repo, parser, logger)

	t.Run("File Does Not Exist", func(t *testing.T) {
		_, err := svc.BuildFromArchive(context.Background(), "/path/to/nowhere.zip")
		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
		if !errors.Is(err, domain.ErrInvalidData) {
			t.Errorf("expected ErrInvalidData, got: %v", err)
		}
	})

	t.Run("Create Log Fails", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "empty_*.zip")
		defer os.Remove(tmpFile.Name())

		repo.mockCreateLog = func(_ context.Context, _ domain.LogStatus) (*domain.Log, error) {
			return nil, errors.New("db down")
		}

		_, err := svc.BuildFromArchive(context.Background(), tmpFile.Name())
		if err == nil {
			t.Error("expected error when CreateLog fails")
		}
	})
}

func TestProcessArchiveAsync_SuccessWorkflow(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	zipPath := createTestZip(t, map[string]string{
		"nodes.db_csv":        "node_id,name",
		"ports.sharp_an_info": "port_id,state",
	})
	defer os.Remove(zipPath)

	var statuses []domain.LogStatus
	repo := &mockParserWriter{
		mockUpdateLogStatus: func(_ context.Context, _ uuid.UUID, status domain.LogStatus) error {
			statuses = append(statuses, status)
			return nil
		},
	}
	parser := &mockTopologyParser{}

	svc := NewBuilder(repo, parser, logger).(*builderService)

	svc.processArchiveAsync(context.Background(), uuid.New(), zipPath)

	if len(statuses) != 2 || statuses[0] != domain.StatusInProgress || statuses[1] != domain.StatusSuccess {
		t.Errorf("unexpected status transitions: %v", statuses)
	}
}

func TestProcessArchiveAsync_CompensationOnParseError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	zipPath := createTestZip(t, map[string]string{
		"nodes.db_csv":        "bad data",
		"ports.sharp_an_info": "bad data",
	})
	defer os.Remove(zipPath)

	deletedCalled := false
	var finalStatus domain.LogStatus

	repo := &mockParserWriter{
		mockUpdateLogStatus: func(_ context.Context, _ uuid.UUID, status domain.LogStatus) error {
			finalStatus = status
			return nil
		},
		mockDeleteLogData: func(_ context.Context, _ uuid.UUID) error {
			deletedCalled = true
			return nil
		},
	}
	parser := &mockTopologyParser{
		mockParseDBCSV: func(_ context.Context, _ uuid.UUID, _ io.Reader) error {
			return errors.New("simulated parsing error")
		},
	}

	svc := NewBuilder(repo, parser, logger).(*builderService)
	svc.processArchiveAsync(context.Background(), uuid.New(), zipPath)

	if !deletedCalled {
		t.Error("expected DeleteLogData to be called during compensation, but it wasn't")
	}
	if finalStatus != domain.StatusError {
		t.Errorf("expected final status to be error, got %v", finalStatus)
	}
}

func TestExtractAndParse_MissingFiles(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := NewBuilder(&mockParserWriter{}, &mockTopologyParser{}, logger).(*builderService)

	t.Run("Missing db_csv", func(t *testing.T) {
		zipPath := createTestZip(t, map[string]string{
			"ports.sharp_an_info": "data",
		})
		defer os.Remove(zipPath)

		err := svc.extractAndParse(context.Background(), uuid.New(), zipPath)
		if !errors.Is(err, domain.ErrInvalidData) {
			t.Errorf("expected ErrInvalidData, got %v", err)
		}
	})

	t.Run("Missing sharp_an_info", func(t *testing.T) {
		zipPath := createTestZip(t, map[string]string{
			"nodes.db_csv": "data",
		})
		defer os.Remove(zipPath)

		err := svc.extractAndParse(context.Background(), uuid.New(), zipPath)
		if !errors.Is(err, domain.ErrInvalidData) {
			t.Errorf("expected ErrInvalidData, got %v", err)
		}
	})
}
