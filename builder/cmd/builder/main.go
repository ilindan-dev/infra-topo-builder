// Package main is the entry point for the InfraTopoBuilder microservice.
// It initializes configuration, connects to the database, runs migrations,
// and starts the REST API server.
package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/adapters/postgres"
	"github.com/ilindan-dev/infra-topo-builder/builder/internal/config"
)

func main() {
	cfg := config.MustLoad()
	logger := mustMakeLogger(cfg.LogLevel)
	if err := run(cfg, logger); err != nil {
		logger.Error("failed to run service", "error", err)
		os.Exit(1)
	}
}

// run initializes all application dependencies (database pool, loggers),
// applies database migrations, and starts the HTTP server with graceful shutdown.
func run(cfg *config.Config, logger *slog.Logger) error {
	cfgPgx, err := pgxpool.ParseConfig(cfg.DataBaseConfig.DataBaseURL)
	if err != nil {
		logger.Error("failed to parse pgx database config", "error", err)
		return err
	}

	cfgPgx.MaxConnLifetime = cfg.DataBaseConfig.MaxConnLifetime
	cfgPgx.MaxConnIdleTime = cfg.DataBaseConfig.MaxConnIdleTime
	cfgPgx.MaxConns = cfg.DataBaseConfig.MaxConns
	cfgPgx.MinConns = cfg.DataBaseConfig.MinConns

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	pgxPool, err := pgxpool.NewWithConfig(ctx, cfgPgx)
	if err != nil {
		logger.Error("failed to create pgx pool", "error", err)
		return err
	}
	defer pgxPool.Close()

	if err := postgres.Migrate(pgxPool, logger); err != nil {
		logger.Error("failed to run migration", "error", err)
		return err
	}

	server := http.Server{
		Addr:         ":" + cfg.HTTPConfig.Port,
		ReadTimeout:  cfg.HTTPConfig.ReadTimeout,
		WriteTimeout: cfg.HTTPConfig.WriteTimeout,
		IdleTimeout:  cfg.HTTPConfig.IdleTimeout,
	}

	go func() {
		logger.Info("server started", "port", cfg.HTTPConfig.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen and serve error", "error", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown forced", "error", err)
		return err
	}
	logger.Info("server exited")

	return nil
}

// mustMakeLogger initializes the logger for further dependency assembly.
func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		log.Fatalf("unknown log level: %s", logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
