package db

import (
	"context"
	_ "embed"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/iamraydoan/factoryos/services/analytics-engine/config"
)

//go:embed migrations/0001_init_telemetry.sql
var initMigrationSQL string

// DB wraps the TimescaleDB connection pool.
type DB struct {
	pool *pgxpool.Pool
	cfg  config.DatabaseConfig
}

// NewDB creates and tests a connection pool to TimescaleDB/PostgreSQL using validated configuration parameters.
func NewDB(ctx context.Context, cfg config.DatabaseConfig) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Apply configuration parameters directly without redundant defensive checks
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database pool: %w", err)
	}

	// Ping database with configured timeout
	pingCtx, cancel := context.WithTimeout(ctx, cfg.PingTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	log.Printf("[TimescaleDB] Successfully connected to pool (MaxConns: %d, MinConns: %d, Table: %s).",
		cfg.MaxConns, cfg.MinConns, cfg.TableName)
	return &DB{pool: pool, cfg: cfg}, nil
}

// AsBatchInserter returns the underlying pool as a BatchInserter interface,
// allowing the BatchWriter to use it without exposing the full pgxpool.Pool.
func (d *DB) AsBatchInserter() BatchInserter {
	if d == nil || d.pool == nil {
		return nil
	}
	return d.pool
}

// RunMigrations executes the initial schema and hypertable setup.
func (d *DB) RunMigrations(ctx context.Context) error {
	log.Println("[TimescaleDB] Running database migrations for raw_telemetry hypertable...")
	if _, err := d.pool.Exec(ctx, initMigrationSQL); err != nil {
		return fmt.Errorf("failed to execute migration script: %w", err)
	}
	log.Println("[TimescaleDB] Migrations applied successfully.")
	return nil
}

// Close gracefully closes the connection pool.
func (d *DB) Close() {
	if d.pool != nil {
		d.pool.Close()
		log.Println("[TimescaleDB] Database connection pool closed.")
	}
}
