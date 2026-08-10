package db

import (
	"context"
	"testing"

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

func TestDB_CloseNil(t *testing.T) {
	dbInstance := &DB{Pool: nil}
	// Calling Close on nil pool should not panic
	dbInstance.Close()
}
