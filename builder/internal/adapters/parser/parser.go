package parser

import (
	"log/slog"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/ports"
)

const defaultBatchSize = 10000

// Parser converts diagnostic files into domain models and writes them using
// the provided ports.ParserWriter. It processes files in a streaming manner
// and flushes records in batches to reduce DB pressure.
type Parser struct {
	repo      ports.ParserWriter
	batchSize int
	logger    *slog.Logger
}

// NewParser constructs a Parser. If batchSize <= 0 the defaultBatchSize is used.
// The returned value implements ports.TopologyParser and is safe for use in
// concurrent request handlers (the Parser itself is stateless beyond the repo).
func NewParser(repo ports.ParserWriter, batchSize int, logger *slog.Logger) ports.TopologyParser {
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	return &Parser{
		repo:      repo,
		batchSize: batchSize,
		logger:    logger.With("adapter", "parser"),
	}
}
