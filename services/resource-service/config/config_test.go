package config

import (
	"os"
	"testing"
)

// TestLoadConfig_Success verifies that valid environment variables are parsed correctly.
func TestLoadConfig_Success(t *testing.T) {
	// Set required env vars
	os.Setenv("DB_URL", "postgres://factoryos:factoryos_password@localhost:5432/factoryos?sslmode=disable")
	defer os.Unsetenv("DB_URL")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() = %v, want nil", err)
	}

	// Verify database config
	if cfg.Database.URL == "" {
		t.Error("Database.URL is empty")
	}
	if cfg.Database.MaxConns != 10 {
		t.Errorf("Database.MaxConns = %d, want 10", cfg.Database.MaxConns)
	}
	if cfg.Database.MinConns != 2 {
		t.Errorf("Database.MinConns = %d, want 2", cfg.Database.MinConns)
	}

	// Verify server defaults
	if cfg.Server.Port != ":50052" {
		t.Errorf("Server.Port = %s, want :50052", cfg.Server.Port)
	}
	if cfg.Server.MetricsPort != ":9092" {
		t.Errorf("Server.MetricsPort = %s, want :9092", cfg.Server.MetricsPort)
	}
}

// TestLoadConfig_MissingURL verifies that missing DB_URL returns an error.
func TestLoadConfig_MissingURL(t *testing.T) {
	os.Unsetenv("DB_URL")

	_, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig() = nil, want error for missing DB_URL")
	}
}

// TestLoadConfig_WithEnvFile verifies that .env file is loaded if present.
func TestLoadConfig_CustomValues(t *testing.T) {
	os.Setenv("DB_URL", "postgres://custom:5432/test")
	os.Setenv("DB_MAX_CONNS", "25")
	os.Setenv("SERVER_PORT", ":9999")
	defer func() {
		os.Unsetenv("DB_URL")
		os.Unsetenv("DB_MAX_CONNS")
		os.Unsetenv("SERVER_PORT")
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() = %v, want nil", err)
	}

	if cfg.Database.MaxConns != 25 {
		t.Errorf("Database.MaxConns = %d, want 25", cfg.Database.MaxConns)
	}
	if cfg.Server.Port != ":9999" {
		t.Errorf("Server.Port = %s, want :9999", cfg.Server.Port)
	}
}
