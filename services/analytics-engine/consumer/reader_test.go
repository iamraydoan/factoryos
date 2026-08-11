package consumer

import (
	"testing"
	"time"

	"github.com/iamraydoan/factoryos/services/analytics-engine/config"
)

func TestNewKafkaReader(t *testing.T) {
	cfg := config.KafkaConfig{
		Brokers:        []string{"localhost:9092"},
		Topic:          "telemetry.raw.v1",
		GroupID:        "test-group",
		MinBytes:       10240,
		MaxBytes:       10485760,
		MaxWait:        100 * time.Millisecond,
		CommitInterval: 500 * time.Millisecond,
	}
	reader := NewKafkaReader(cfg)
	if reader == nil {
		t.Fatalf("expected non-nil kafka.Reader")
	}
	defer reader.Close()

	readerCfg := reader.Config()
	if readerCfg.Topic != "telemetry.raw.v1" {
		t.Fatalf("expected topic telemetry.raw.v1, got: %s", readerCfg.Topic)
	}
	if readerCfg.GroupID != "test-group" {
		t.Fatalf("expected groupID test-group, got: %s", readerCfg.GroupID)
	}
}

func TestNewKafkaReader_Defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb?sslmode=disable")

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load default config: %v", err)
	}
	reader := NewKafkaReader(cfg.Kafka)
	if reader == nil {
		t.Fatalf("expected non-nil reader on default config")
	}
	defer reader.Close()
}
