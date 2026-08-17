package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var validate = validator.New()

// Config holds all configuration parameters for the Analytics Engine,
// parsed via struct tags from Environment Variables and optional .env file.
type Config struct {
	Server    ServerConfig
	Kafka     KafkaConfig
	Database  DatabaseConfig
	Ingestion IngestionConfig
	OEE       OEEConfig
}

// ServerConfig holds HTTP server parameters.
type ServerConfig struct {
	MetricsPort     string        `env:"METRICS_PORT" envDefault:":8082" validate:"required"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"5s" validate:"gt=0"`
}

// KafkaConfig holds Kafka connection and consumer tuning parameters.
type KafkaConfig struct {
	Brokers        []string      `env:"KAFKA_BROKERS" envSeparator:"," envDefault:"localhost:9092" validate:"required,min=1,dive,required"`
	Topic          string        `env:"KAFKA_TOPIC" envDefault:"telemetry.raw.v1" validate:"required"`
	GroupID        string        `env:"KAFKA_GROUP_ID" envDefault:"analytics-engine-group" validate:"required"`
	MinBytes       int           `env:"KAFKA_MIN_BYTES" envDefault:"10240" validate:"gt=0"`
	MaxBytes       int           `env:"KAFKA_MAX_BYTES" envDefault:"10485760" validate:"gt=0,gtefield=MinBytes"`
	MaxWait        time.Duration `env:"KAFKA_MAX_WAIT" envDefault:"100ms" validate:"gt=0"`
	CommitInterval time.Duration `env:"KAFKA_COMMIT_INTERVAL" envDefault:"500ms" validate:"gt=0"`
	RetryBackoff   time.Duration `env:"KAFKA_RETRY_BACKOFF" envDefault:"50ms" validate:"gt=0"`
}

// DatabaseConfig holds TimescaleDB connection pool tuning parameters.
type DatabaseConfig struct {
	// URL must be set explicitly via DATABASE_URL environment variable or .env file.
	// See .env.example for the expected format. No default is provided to prevent
	// accidental use of hardcoded credentials in staging or production environments.
	URL              string        `env:"DATABASE_URL" validate:"required"`
	TableName        string        `env:"DATABASE_TABLE_NAME" envDefault:"raw_telemetry" validate:"required"`
	MaxConns         int32         `env:"DATABASE_MAX_CONNS" envDefault:"30" validate:"gt=0"`
	MinConns         int32         `env:"DATABASE_MIN_CONNS" envDefault:"5" validate:"gte=0,ltefield=MaxConns"`
	MaxConnLifetime  time.Duration `env:"DATABASE_MAX_CONN_LIFETIME" envDefault:"1h" validate:"gt=0"`
	MaxConnIdleTime  time.Duration `env:"DATABASE_MAX_CONN_IDLE_TIME" envDefault:"15m" validate:"gt=0"`
	PingTimeout      time.Duration `env:"DATABASE_PING_TIMEOUT" envDefault:"5s" validate:"gt=0"`
	RawRetentionDays int           `env:"TELEMETRY_RAW_RETENTION_DAYS" envDefault:"30" validate:"gt=0"`
}

// IngestionConfig holds batch buffer tuning parameters.
type IngestionConfig struct {
	BatchSize                 int           `env:"BATCH_SIZE" envDefault:"500" validate:"gt=0"`
	FlushInterval             time.Duration `env:"FLUSH_INTERVAL" envDefault:"200ms" validate:"gt=0"`
	ChannelCapacityMultiplier int           `env:"CHANNEL_CAPACITY_MULTIPLIER" envDefault:"4" validate:"gt=0"`
}

// OEEAlertConfig holds OEE threshold alerting parameters.
type OEEAlertConfig struct {
	OEEWarningThreshold        float64 `env:"OEE_WARN" envDefault:"0.85"`
	OEECriticalThreshold       float64 `env:"OEE_CRIT" envDefault:"0.70"`
	ComponentWarningThreshold  float64 `env:"COMP_WARN" envDefault:"0.90"`
	ComponentCriticalThreshold float64 `env:"COMP_CRIT" envDefault:"0.75"`
	CooldownMinutes            int     `env:"COOLDOWN_MIN" envDefault:"5"`
}

// OEEConfig holds parameters for the real-time OEE streaming aggregator.
type OEEConfig struct {
	WindowDuration        time.Duration  `env:"OEE_WINDOW_DURATION" envDefault:"1h" validate:"gt=0"`
	SnapshotInterval      time.Duration  `env:"OEE_SNAPSHOT_INTERVAL" envDefault:"15s" validate:"gt=0"`
	DefaultIdealCycleTime time.Duration  `env:"OEE_DEFAULT_IDEAL_CYCLE_TIME" envDefault:"30s" validate:"gt=0"`
	OEEAlert              OEEAlertConfig `envPrefix:"OEE_ALERT_"`
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
