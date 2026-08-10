package db

import (
	"context"
	"testing"
	"time"

	"github.com/iamraydoan/factoryos/services/analytics-engine/config"
)

func TestNewDB_InvalidURL(t *testing.T) {
	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	cfg.Database.URL = "invalid-url://bad-conn"
	_, err = NewDB(ctx, cfg.Database)
	if err == nil {
		t.Fatalf("expected error for invalid DB URL, got nil")
	}
}

func TestNewDB_UnreachableHost(t *testing.T) {
	ctx := context.Background()
	cfg := config.DatabaseConfig{
		URL:              "postgres://user:pass@127.0.0.1:1/nonexistent?sslmode=disable",
		TableName:        "raw_telemetry",
		MaxConns:         5,
		MinConns:         1,
		MaxConnLifetime:  time.Hour,
		MaxConnIdleTime:  15 * time.Minute,
		PingTimeout:      50 * time.Millisecond,
		RawRetentionDays: 30,
	}
	_, err := NewDB(ctx, cfg)
	if err == nil {
		t.Fatalf("expected ping timeout error for unreachable host, got nil")
	}
}

func TestDB_CloseNil(t *testing.T) {
	dbInstance := &DB{pool: nil}
	// Calling Close on nil pool should not panic
	dbInstance.Close()
}

func TestDB_AsBatchInserter(t *testing.T) {
	dbInstance := &DB{pool: nil}
	inserter := dbInstance.AsBatchInserter()
	if inserter != nil {
		t.Fatalf("expected nil BatchInserter for nil pool, got: %v", inserter)
	}
}
