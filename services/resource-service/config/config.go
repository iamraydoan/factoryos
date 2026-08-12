package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

// validate is the shared validator instance for all config structs.
var validate = validator.New()

// Config holds all configuration parameters for the Resource Service,
// parsed via struct tags from Environment Variables and optional .env file.
type Config struct {
	Server   ServerConfig   `envPrefix:"SERVER_"`
	Database DatabaseConfig `envPrefix:"DB_"`
}

// ServerConfig holds gRPC server parameters.
type ServerConfig struct {
	Port            string        `env:"PORT" envDefault:":50052" validate:"required"`
	MetricsPort     string        `env:"METRICS_PORT" envDefault:":9092" validate:"required"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"5s" validate:"gt=0"`
}

// DatabaseConfig holds PostgreSQL connection pool tuning parameters.
type DatabaseConfig struct {
	// URL must be set explicitly via DB_URL environment variable or .env file.
	// No default is provided to prevent accidental use of hardcoded credentials.
	URL             string        `env:"URL" validate:"required"`
	MaxConns        int32         `env:"MAX_CONNS" envDefault:"10" validate:"gt=0"`
	MinConns        int32         `env:"MIN_CONNS" envDefault:"2" validate:"gte=0,ltefield=MaxConns"`
	MaxConnLifetime time.Duration `env:"MAX_CONN_LIFETIME" envDefault:"1h" validate:"gt=0"`
	MaxConnIdleTime time.Duration `env:"MAX_CONN_IDLE_TIME" envDefault:"15m" validate:"gt=0"`
	PingTimeout     time.Duration `env:"PING_TIMEOUT" envDefault:"5s" validate:"gt=0"`
}

// Validate verifies all configuration parameters using go-playground/validator.
func (c *Config) Validate() error {
	return validate.Struct(c)
}

// LoadConfig reads .env (if present), parses env variables with defaults, and validates.
func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}
