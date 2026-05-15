package service

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/domain"
	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/ports"
)

var _ ports.TopologyBuilder = (*builderService)(nil)

// builderService is the concrete implementation of ports.TopologyBuilder used
// by the application. It orchestrates the parsing of archive artifacts and
// persistence of parsed entities via the provided ParserWriter and
// TopologyParser ports. The struct is unexported: construction should go
// through NewBuilder so callers receive the ports.TopologyBuilder interface.
type builderService struct {
	repo   ports.ParserWriter
	parser ports.TopologyParser
	logger *slog.Logger
}

// NewBuilder constructs a TopologyBuilder service instance. The returned
// value implements ports.TopologyBuilder and coordinates parsing and storage
// operations. The logger will be decorated with component metadata.
func NewBuilder(repo ports.ParserWriter, parser ports.TopologyParser, logger *slog.Logger) ports.TopologyBuilder {
	return &builderService{
		repo:   repo,
		parser: parser,
		logger: logger.With("layer", "service", "component", "builder"),
	}
}

// BuildFromArchive validates the provided archive path, creates a new import
// log record and schedules asynchronous processing of the archive. It returns
// the created log ID immediately so callers can poll status via the
// application's read-side APIs. Errors are returned for immediate
// validation failures (for example when the archive is missing) or when the
// initial log record cannot be created.
func (s *builderService) BuildFromArchive(ctx context.Context, archivePath string) (uuid.UUID, error) {
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		return uuid.Nil, fmt.Errorf("%w: archive %s does not exist", domain.ErrInvalidData, archivePath)
	}

	logRecord, err := s.repo.CreateLog(ctx, domain.StatusPending)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create log session: %w", err)
	}

	// Run parse+persist in background. Use WithoutCancel so the background
	// worker continues even if the HTTP request context is cancelled by the
	// caller.
	go s.processArchiveAsync(context.WithoutCancel(ctx), logRecord.ID, archivePath)

	return logRecord.ID, nil
}

// processArchiveAsync executes the extract+parse workflow in a background
// goroutine. It updates the log status (in_progress → success|error) and logs
// any errors encountered. The function is intentionally unexported because
// background scheduling is managed by BuildFromArchive.
func (s *builderService) processArchiveAsync(ctx context.Context, logID uuid.UUID, archivePath string) {
	log := s.logger.With("method", "processArchiveAsync", "log_id", logID)
	log.Info("started background archive processing")

	if err := s.repo.UpdateLogStatus(ctx, logID, domain.StatusInProgress); err != nil {
		log.Error("failed to update status to in_progress", "error", err)
		return
	}

	if err := s.extractAndParse(ctx, logID, archivePath); err != nil {
		log.Error("parsing failed", "error", err)

		if cleanupErr := s.repo.DeleteLogData(ctx, logID); cleanupErr != nil {
			log.Error("failed to cleanup dirty data after parsing error", "error", cleanupErr)
		}

		_ = s.repo.UpdateLogStatus(ctx, logID, domain.StatusError)
		return
	}

	if err := s.repo.UpdateLogStatus(ctx, logID, domain.StatusSuccess); err != nil {
		log.Error("failed to update status to success", "error", err)
		return
	}

	log.Info("archive processing completed successfully")
}

// extractAndParse opens the provided ZIP archive, locates the expected
// artifacts (.db_csv and .sharp_an_info) and dispatches them to the parser
// callbacks. It returns domain.ErrInvalidData when required files are
// missing and wraps parser errors to provide context. The function also logs
// file-open/close failures for observability.
func (s *builderService) extractAndParse(ctx context.Context, logID uuid.UUID, archivePath string) error {
	log := s.logger.With("method", "extractAndParse", "logID", logID)

	r, err := zip.OpenReader(archivePath)
	if err != nil {
		log.Error("failed to open archive", "error", err)
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer func() {
		err := r.Close()
		if err != nil {
			log.Warn("failed to close zip file", "error", err)
		}
	}()

	var dbCsvFile *zip.File
	var sharpInfoFile *zip.File

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".db_csv") {
			dbCsvFile = f
		} else if strings.HasSuffix(f.Name, ".sharp_an_info") {
			sharpInfoFile = f
		}
	}

	if dbCsvFile == nil {
		log.Error("failed to find .db_csv file")
		return fmt.Errorf("%w: missing .db_csv file in archive", domain.ErrInvalidData)
	}
	if sharpInfoFile == nil {
		log.Error("failed to find .sharp_an_info file")
		return fmt.Errorf("%w: missing .sharp_an_info file in archive", domain.ErrInvalidData)
	}

	if err := s.parseZipFile(ctx, logID, dbCsvFile, s.parser.ParseDBCSV); err != nil {
		return fmt.Errorf("failed to parse db_csv: %w", err)
	}

	if err := s.parseZipFile(ctx, logID, sharpInfoFile, s.parser.ParseSharpAnInfo); err != nil {
		return fmt.Errorf("failed to parse sharp_an_info: %w", err)
	}

	return nil
}

// parseZipFile opens a file from the zip archive and invokes the provided
// parser function with the file's reader. The caller is responsible for
// selecting the correct parser callback (for example ParseDBCSV or
// ParseSharpAnInfo). Any errors returned by the parser are propagated to the
// caller. The function logs open/close failures for observability.
func (s *builderService) parseZipFile(
	ctx context.Context, logID uuid.UUID, file *zip.File, parseFunc func(context.Context, uuid.UUID, io.Reader) error,
) error {
	log := s.logger.With("method", "parseZipFile", "logID", logID)
	rc, err := file.Open()
	if err != nil {
		log.Error("failed to open zip file", "error", err)
		return err
	}
	defer func() {
		err := rc.Close()
		if err != nil {
			log.Warn("failed to close zip file", "error", err)
		}
	}()

	return parseFunc(ctx, logID, rc)
}
