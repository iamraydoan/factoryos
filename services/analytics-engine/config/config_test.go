package config

import (
	"testing"
	"time"
)

func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb?sslmode=disable")

	// Clear relevant envs (t.Setenv restores original values after test)
	t.Setenv("KAFKA_BROKERS", "")
	t.Setenv("KAFKA_TOPIC", "")
	t.Setenv("BATCH_SIZE", "")
	t.Setenv("FLUSH_INTERVAL", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if len(cfg.Kafka.Brokers) != 1 || cfg.Kafka.Brokers[0] != "localhost:9092" {
		t.Fatalf("expected default broker localhost:9092, got: %v", cfg.Kafka.Brokers)
	}
	if cfg.Kafka.Topic != "telemetry.raw.v1" {
		t.Fatalf("expected default topic telemetry.raw.v1, got: %s", cfg.Kafka.Topic)
	}
	if cfg.Ingestion.BatchSize != 500 {
		t.Fatalf("expected default batch size 500, got: %d", cfg.Ingestion.BatchSize)
	}
	if cfg.Ingestion.FlushInterval != 200*time.Millisecond {
		t.Fatalf("expected default flush interval 200ms, got: %v", cfg.Ingestion.FlushInterval)
	}
	if cfg.Database.TableName != "raw_telemetry" {
		t.Fatalf("expected default table name raw_telemetry, got: %s", cfg.Database.TableName)
	}
}

func TestLoadConfig_CustomEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb?sslmode=disable")
	t.Setenv("KAFKA_BROKERS", "kafka-1:9092,kafka-2:9092")
	t.Setenv("KAFKA_TOPIC", "custom.telemetry.v1")
	t.Setenv("BATCH_SIZE", "1000")
	t.Setenv("FLUSH_INTERVAL", "500ms")
	t.Setenv("TELEMETRY_RAW_RETENTION_DAYS", "60")
	t.Setenv("DATABASE_TABLE_NAME", "custom_telemetry")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if len(cfg.Kafka.Brokers) != 2 || cfg.Kafka.Brokers[1] != "kafka-2:9092" {
		t.Fatalf("expected 2 brokers, got: %v", cfg.Kafka.Brokers)
	}
	if cfg.Kafka.Topic != "custom.telemetry.v1" {
		t.Fatalf("expected custom.telemetry.v1, got: %s", cfg.Kafka.Topic)
	}
	if cfg.Ingestion.BatchSize != 1000 {
		t.Fatalf("expected batch size 1000, got: %d", cfg.Ingestion.BatchSize)
	}
	if cfg.Ingestion.FlushInterval != 500*time.Millisecond {
		t.Fatalf("expected flush interval 500ms, got: %v", cfg.Ingestion.FlushInterval)
	}
	if cfg.Database.RawRetentionDays != 60 {
		t.Fatalf("expected retention days 60, got: %d", cfg.Database.RawRetentionDays)
	}
	if cfg.Database.TableName != "custom_telemetry" {
		t.Fatalf("expected table custom_telemetry, got: %s", cfg.Database.TableName)
	}
}

func TestConfig_Validate(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb?sslmode=disable")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error loading default config: %v", err)
	}

	// Case 1: Missing MetricsPort
	cfg.Server.MetricsPort = ""
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected error for empty metrics port")
	}

	// Case 2: Missing Kafka Brokers
	cfg, _ = LoadConfig()
	cfg.Kafka.Brokers = []string{}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected error for empty brokers")
	}

	// Case 3: Missing Kafka Topic
	cfg, _ = LoadConfig()
	cfg.Kafka.Topic = ""
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected error for empty topic")
	}

	// Case 4: Missing Database URL
	cfg, _ = LoadConfig()
	cfg.Database.URL = ""
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected error for empty database URL")
	}

	// Case 5: MinConns > MaxConns
	cfg, _ = LoadConfig()
	cfg.Database.MinConns = 100
	cfg.Database.MaxConns = 10
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected error for min_conns > max_conns")
	}

	// Case 6: BatchSize <= 0
	cfg, _ = LoadConfig()
	cfg.Ingestion.BatchSize = 0
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error when MaxBytes < MinBytes, got nil")
	}
}

func TestLoadConfig_MissingRequiredEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	_, err := LoadConfig()
	if err == nil {
		t.Fatalf("expected error when DATABASE_URL is missing, got nil")
	}
}

func TestConfig_ValidateErrors(t *testing.T) {
	// Invalid: batch size <= 0
	cfg := &Config{
		Server: ServerConfig{MetricsPort: ":8082", ShutdownTimeout: 5 * time.Second},
		Kafka: KafkaConfig{
			Brokers:        []string{"localhost:9092"},
			Topic:          "topic",
			GroupID:        "group",
			MinBytes:       1024,
			MaxBytes:       2048,
			MaxWait:        100 * time.Millisecond,
			CommitInterval: 500 * time.Millisecond,
			RetryBackoff:   50 * time.Millisecond,
		},
		Database: DatabaseConfig{
			URL:              "postgres://localhost:5432/db",
			TableName:        "raw_telemetry",
			MaxConns:         10,
			MinConns:         2,
			MaxConnLifetime:  time.Hour,
			MaxConnIdleTime:  15 * time.Minute,
			PingTimeout:      5 * time.Second,
			RawRetentionDays: 30,
		},
		Ingestion: IngestionConfig{
			BatchSize:                 0, // Invalid: gt=0
			FlushInterval:             200 * time.Millisecond,
			ChannelCapacityMultiplier: 4,
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error for BatchSize=0, got nil")
	}
}
