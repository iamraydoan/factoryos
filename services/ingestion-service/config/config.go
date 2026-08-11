package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var validate = validator.New()

// Config holds all configuration parameters for the Ingestion Service.
type Config struct {
	Server ServerConfig
	Kafka  KafkaConfig
}

// ServerConfig holds gRPC and HTTP metrics server parameters.
type ServerConfig struct {
	GRPCPort        string        `env:"GRPC_PORT" envDefault:":50051" validate:"required"`
	MetricsPort     string        `env:"METRICS_PORT" envDefault:":8083" validate:"required"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"5s" validate:"gt=0"`
}

// KafkaConfig holds Kafka producer parameters for telemetry streaming.
type KafkaConfig struct {
	Brokers      []string      `env:"KAFKA_BROKERS" envSeparator:"," envDefault:"localhost:9092" validate:"required,min=1,dive,required"`
	Topic        string        `env:"KAFKA_TOPIC" envDefault:"telemetry.raw.v1" validate:"required"`
	BatchSize    int           `env:"KAFKA_PRODUCER_BATCH_SIZE" envDefault:"500" validate:"gt=0"`
	BatchTimeout time.Duration `env:"KAFKA_PRODUCER_BATCH_TIMEOUT" envDefault:"10ms" validate:"gt=0"`
	RequiredAcks int           `env:"KAFKA_REQUIRED_ACKS" envDefault:"1" validate:"gte=-1,lte=1"`
	Async        bool          `env:"KAFKA_PRODUCER_ASYNC" envDefault:"false"`
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
