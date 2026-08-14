package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// PhysicalAsset represents a real machine with a serial number.
type PhysicalAsset struct {
	ID           string
	Name         string
	SerialNumber string
	Manufacturer string
	Model        string
	AssetType    string
	Status       string     // "active", "faulted", "under_maintenance", "decommissioned"
	InstalledAt  *time.Time // Denormalized: when current installation began (nil = not installed)
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// PhysicalAssetRepository defines data access methods for Physical Assets.
type PhysicalAssetRepository interface {
	CreatePhysicalAsset(ctx context.Context, asset *PhysicalAsset) (*PhysicalAsset, error)
	GetPhysicalAsset(ctx context.Context, id string) (*PhysicalAsset, error)
	ListPhysicalAssets(ctx context.Context) ([]*PhysicalAsset, error)
}

// CreatePhysicalAsset inserts a new Physical Asset and returns the generated row.
func (r *PostgresEquipmentRepository) CreatePhysicalAsset(ctx context.Context, asset *PhysicalAsset) (*PhysicalAsset, error) {
	query := `
		INSERT INTO physical_assets (name, serial_number, manufacturer, model, asset_type, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, serial_number, manufacturer, model, asset_type, status, installed_at, created_at, updated_at`

	var pa PhysicalAsset
	err := r.pool.QueryRow(ctx, query,
		asset.Name, asset.SerialNumber, asset.Manufacturer,
		asset.Model, asset.AssetType, asset.Status,
	).Scan(
		&pa.ID, &pa.Name, &pa.SerialNumber, &pa.Manufacturer,
		&pa.Model, &pa.AssetType, &pa.Status, &pa.InstalledAt,
		&pa.CreatedAt, &pa.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create physical asset: %w", err)
	}
	return &pa, nil
}

// GetPhysicalAsset retrieves a Physical Asset by ID. Returns nil, nil if not found.
func (r *PostgresEquipmentRepository) GetPhysicalAsset(ctx context.Context, id string) (*PhysicalAsset, error) {
	query := `
		SELECT id, name, serial_number, manufacturer, model, asset_type, status, installed_at, created_at, updated_at
		FROM physical_assets
		WHERE id = $1`

	var pa PhysicalAsset
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&pa.ID, &pa.Name, &pa.SerialNumber, &pa.Manufacturer,
		&pa.Model, &pa.AssetType, &pa.Status, &pa.InstalledAt,
		&pa.CreatedAt, &pa.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get physical asset: %w", err)
	}
	return &pa, nil
}

// ListPhysicalAssets retrieves all Physical Assets, ordered by name.
func (r *PostgresEquipmentRepository) ListPhysicalAssets(ctx context.Context) ([]*PhysicalAsset, error) {
	query := `
		SELECT id, name, serial_number, manufacturer, model, asset_type, status, installed_at, created_at, updated_at
		FROM physical_assets
		ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list physical assets: %w", err)
	}
	defer rows.Close()

	var assets []*PhysicalAsset
	for rows.Next() {
		var pa PhysicalAsset
		if err := rows.Scan(
			&pa.ID, &pa.Name, &pa.SerialNumber, &pa.Manufacturer,
			&pa.Model, &pa.AssetType, &pa.Status, &pa.InstalledAt,
			&pa.CreatedAt, &pa.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan physical asset: %w", err)
		}
		assets = append(assets, &pa)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating physical assets: %w", err)
	}
	return assets, nil
}
