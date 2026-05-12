package postgres

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// migrationFiles contains SQL migration scripts embedded in a binary file.
//
//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrate updates the database schema to the latest version using the go-migrate turn.
// Returns an error if the migration could not be applied.
func Migrate(pool *pgxpool.Pool, logger *slog.Logger) error {
	log := logger.With("adapter", "postgres", "function", "Migrate")
	log.Debug("running migration")

	files, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("failed to load migration files: %w", err)
	}

	db := stdlib.OpenDB(*pool.Config().ConnConfig)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Error("Failed to close DB connection", "error", err)
		}
	}(db)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Error("Failed to create postgres driver", "error", err)
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", files, "postgres", driver)
	if err != nil {
		log.Error("Failed to create migration instance", "error", err)
		return fmt.Errorf("failed to create migration: %w", err)
	}

	err = m.Up()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Error("migration failed", "error", err)
			return fmt.Errorf("migration failed: %w", err)
		}
		log.Debug("migration did not change anything")
	}

	log.Debug("migration finished")
	return nil
}
