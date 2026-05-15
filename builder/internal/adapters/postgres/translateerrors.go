package postgres

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/domain"
)

// translateError maps low-level DB/driver errors to domain-level errors
// and logs details for observability. This keeps the core business logic
// independent from driver-specific error types.
//
// Mappings:
//   - pgx.ErrNoRows -> domain.ErrNotFound
//   - Postgres unique constraint violation -> domain.ErrAlreadyExists (with constraint name)
//   - Postgres foreign key violation -> domain.ErrConflict (with constraint name)
//   - String truncation -> domain.ErrInvalidData
//
// Any other error is wrapped as domain.ErrInternalServer.
func translateError(err error, method string, log *slog.Logger) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		log.Error("not found rows", "error", err, "method", method)
		return domain.ErrNotFound
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			log.Error("entity already exists", "error", err, "method", method, "constraint", pgErr.ConstraintName)
			return fmt.Errorf("%w: %s", domain.ErrAlreadyExists, pgErr.ConstraintName)
		case pgerrcode.ForeignKeyViolation:
			log.Error("conflict (foreign key)", "error", err, "method", method, "constraint", pgErr.ConstraintName)
			return fmt.Errorf("%w: %s", domain.ErrConflict, pgErr.ConstraintName)
		case pgerrcode.StringDataRightTruncationDataException:
			log.Error("invalid data (string truncation)", "error", err, "method", method)
			return fmt.Errorf("%w: value too long", domain.ErrInvalidData)
		}
	}

	log.Error("internal error", "error", err, "method", method)
	return fmt.Errorf("%w: %w", domain.ErrInternalServer, err)
}
