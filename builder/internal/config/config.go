// Package config provides loading configurations from ENV.
// It also provides structures for storing these configurations.
package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// DataBaseConfig holds PostgreSQL connection settings loaded from environment variables.
// Fields map to ENV variables used to configure the connection pool.
//
// DataBaseURL     - connection URL used by pgx
// MaxConns        - maximum open connections in the pool
// MinConns        - minimum connections to keep alive
// MaxConnLifetime - maximum lifetime of a connection
// MaxConnIdleTime - maximum idle time for a connection
type DataBaseConfig struct {
	DataBaseURL     string        `env:"DATABASE_URL"`
	MaxConns        int32         `env:"MAX_CONNS" env-default:"20"`
	MinConns        int32         `env:"MIN_CONNS" env-default:"5"`
	MaxConnLifetime time.Duration `env:"MAX_CONN_LIFETIME" env-default:"1h"`
	MaxConnIdleTime time.Duration `env:"MAX_CONN_IDLE_TIME" env-default:"30m"`
}

// HTTPConfig holds HTTP server configuration values loaded from environment variables.
// Port and timeouts control the HTTP server behavior used by the application.
// Timeouts are parsed as time.Duration (e.g., "5s", "1m").
type HTTPConfig struct {
	Port         string        `env:"PORT" env-default:"8080"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"10m"`
	IdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"120s"`
}

// Config aggregates application configuration: database, HTTP server, and runtime flags.
// Values are populated from the provided config file and environment variables.
// Use MustLoad to read and validate configuration at startup.
type Config struct {
	DataBaseConfig DataBaseConfig
	HTTPConfig     HTTPConfig
	LogLevel       string `env:"LOG_LEVEL" env-default:"debug"`
	RPS            int    `env:"RPS" env-default:"100"`
	BatchSize      int    `env:"BATCH_SIZE" env-default:"100"`
}

// MustLoad reads configuration from the ENV and exits the process on error.
// This is suitable for program startup where missing or invalid configuration should stop execution.
func MustLoad() *Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read config %s", err)
	}
	return &cfg
}
