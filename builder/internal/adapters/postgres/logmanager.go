package postgres

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/adapters/postgres/gen"
	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/ports"
)

var _ ports.LogManager = (*logManager)(nil)

// logManager implements ports.LogManager and encapsulates lifecycle operations
// for parsing session logs in Postgres. Typical responsibilities include:
//   - removing a log record and any dependent data (cleanup),
//   - providing a stable API used by admin endpoints or background workers.
//
// Implementation details:
//   - Delegates SQL work to the sqlc-generated gen.Queries for type-safe access.
//   - Translates low-level DB/driver errors into canonical domain errors.
//   - Designed to be small and testable so lifecycle policies can be exercised
//     independently from business import logic.
type logManager struct {
	pool    *pgxpool.Pool
	queries *gen.Queries
	logger  *slog.Logger
}

// NewLogManager constructs a Postgres-backed LogManager. The returned
// instance is suitable for administrative and background tasks that manage
// log lifecycle (for example, deletion or archive cleanup). The logger is
// annotated with adapter metadata for observability.
func NewLogManager(pool *pgxpool.Pool, logger *slog.Logger) ports.LogManager {
	return &logManager{
		pool:    pool,
		queries: gen.New(pool),
		logger:  logger.With("adapter", "postgres", "implement", "logManager"),
	}
}

// DeleteLog removes the log record identified by id. The function delegates
// the delete operation to the generated SQL helper and translates low-level
// errors into domain-level errors. Callers should be aware of cascade
// semantics (implementation dependent) and perform any required cleanup.
func (r *logManager) DeleteLog(ctx context.Context, id uuid.UUID) error {
	err := r.queries.DeleteLog(ctx, id)
	return translateError(err, "DeleteLog", r.logger)
}
