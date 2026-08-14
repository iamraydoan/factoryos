package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PhysicalAssetInstallation records a time-bounded link between a Physical Asset and a Work Unit.
type PhysicalAssetInstallation struct {
	ID              string
	PhysicalAssetID string
	WorkUnitID      string
	InstalledAt     time.Time
	RemovedAt       *time.Time // nil = currently installed
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// InstallationRepository defines data access methods for Physical Asset installations.
type InstallationRepository interface {
	InstallAsset(ctx context.Context, physicalAssetID, workUnitID string) (*PhysicalAssetInstallation, error)
	UninstallAsset(ctx context.Context, workUnitID string) (*PhysicalAssetInstallation, error)
	GetCurrentInstallation(ctx context.Context, workUnitID string) (*PhysicalAssetInstallation, error)
	ListInstallations(ctx context.Context, workUnitID string) ([]*PhysicalAssetInstallation, error)
}

// InstallAsset installs a Physical Asset at a Work Unit within a transaction.
// Enforces: one asset per work unit, one work unit per asset (at a time).
// Uses SELECT ... FOR UPDATE for defense-in-depth; the database-level partial
// unique indexes (migration 0004) are the authoritative guard against races.
func (r *PostgresEquipmentRepository) InstallAsset(ctx context.Context, physicalAssetID, workUnitID string) (*PhysicalAssetInstallation, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // No-op after successful commit

	// Guard: asset must not be installed elsewhere (FOR UPDATE prevents concurrent races)
	var occupiedWU *string
	err = tx.QueryRow(ctx,
		`SELECT work_unit_id FROM physical_asset_installations
		 WHERE physical_asset_id = $1 AND removed_at IS NULL
		 FOR UPDATE`,
		physicalAssetID,
	).Scan(&occupiedWU)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing installation: %w", err)
	}
	if occupiedWU != nil {
		return nil, fmt.Errorf("asset is already installed at work unit %s", *occupiedWU)
	}

	// Guard: work unit must not already have an active installation (FOR UPDATE prevents concurrent races)
	var occupiedAsset *string
	err = tx.QueryRow(ctx,
		`SELECT physical_asset_id FROM physical_asset_installations
		 WHERE work_unit_id = $1 AND removed_at IS NULL
		 FOR UPDATE`,
		workUnitID,
	).Scan(&occupiedAsset)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to check work unit installation: %w", err)
	}
	if occupiedAsset != nil {
		return nil, fmt.Errorf("work unit already has asset %s installed", *occupiedAsset)
	}

	// Create installation record
	var inst PhysicalAssetInstallation
	err = tx.QueryRow(ctx,
		`INSERT INTO physical_asset_installations (physical_asset_id, work_unit_id)
		 VALUES ($1, $2)
		 RETURNING id, physical_asset_id, work_unit_id, installed_at, removed_at, created_at, updated_at`,
		physicalAssetID, workUnitID,
	).Scan(&inst.ID, &inst.PhysicalAssetID, &inst.WorkUnitID,
		&inst.InstalledAt, &inst.RemovedAt, &inst.CreatedAt, &inst.UpdatedAt)
	if err != nil {
		// Handle unique constraint violation from migration 0004 as a business error
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("concurrent installation detected for asset %s or work unit %s", physicalAssetID, workUnitID)
		}
		return nil, fmt.Errorf("failed to create installation record: %w", err)
	}

	// Update denormalized pointer on work_units
	_, err = tx.Exec(ctx,
		`UPDATE work_units SET physical_asset_id = $1, updated_at = now() WHERE id = $2`,
		physicalAssetID, workUnitID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update work unit pointer: %w", err)
	}

	// Update denormalized installed_at on physical_assets
	_, err = tx.Exec(ctx,
		`UPDATE physical_assets SET installed_at = now(), updated_at = now() WHERE id = $1`,
		physicalAssetID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update asset installed_at: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &inst, nil
}

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// UninstallAsset removes the currently installed Physical Asset from a Work Unit.
// Returns nil, nil if no active installation exists.
func (r *PostgresEquipmentRepository) UninstallAsset(ctx context.Context, workUnitID string) (*PhysicalAssetInstallation, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // No-op after successful commit

	// Close the active installation by setting removed_at
	var inst PhysicalAssetInstallation
	err = tx.QueryRow(ctx,
		`UPDATE physical_asset_installations
		 SET removed_at = now(), updated_at = now()
		 WHERE work_unit_id = $1 AND removed_at IS NULL
		 RETURNING id, physical_asset_id, work_unit_id, installed_at, removed_at, created_at, updated_at`,
		workUnitID,
	).Scan(&inst.ID, &inst.PhysicalAssetID, &inst.WorkUnitID,
		&inst.InstalledAt, &inst.RemovedAt, &inst.CreatedAt, &inst.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to uninstall asset: %w", err)
	}

	// Clear denormalized pointer on work_units
	_, err = tx.Exec(ctx,
		`UPDATE work_units SET physical_asset_id = NULL, updated_at = now() WHERE id = $1`,
		workUnitID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to clear work unit pointer: %w", err)
	}

	// Clear denormalized installed_at on physical_assets
	_, err = tx.Exec(ctx,
		`UPDATE physical_assets SET installed_at = NULL, updated_at = now() WHERE id = $1`,
		inst.PhysicalAssetID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to clear asset installed_at: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &inst, nil
}

// GetCurrentInstallation returns the active installation for a Work Unit.
// Returns nil, nil if no asset is currently installed.
func (r *PostgresEquipmentRepository) GetCurrentInstallation(ctx context.Context, workUnitID string) (*PhysicalAssetInstallation, error) {
	query := `
		SELECT id, physical_asset_id, work_unit_id, installed_at, removed_at, created_at, updated_at
		FROM physical_asset_installations
		WHERE work_unit_id = $1 AND removed_at IS NULL`

	var inst PhysicalAssetInstallation
	err := r.pool.QueryRow(ctx, query, workUnitID).Scan(
		&inst.ID, &inst.PhysicalAssetID, &inst.WorkUnitID,
		&inst.InstalledAt, &inst.RemovedAt, &inst.CreatedAt, &inst.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get current installation: %w", err)
	}
	return &inst, nil
}

// ListInstallations returns the full installation history for a Work Unit.
func (r *PostgresEquipmentRepository) ListInstallations(ctx context.Context, workUnitID string) ([]*PhysicalAssetInstallation, error) {
	query := `
		SELECT id, physical_asset_id, work_unit_id, installed_at, removed_at, created_at, updated_at
		FROM physical_asset_installations
		WHERE work_unit_id = $1
		ORDER BY installed_at DESC`

	rows, err := r.pool.Query(ctx, query, workUnitID)
	if err != nil {
		return nil, fmt.Errorf("failed to list installations: %w", err)
	}
	defer rows.Close()

	var installations []*PhysicalAssetInstallation
	for rows.Next() {
		var inst PhysicalAssetInstallation
		if err := rows.Scan(
			&inst.ID, &inst.PhysicalAssetID, &inst.WorkUnitID,
			&inst.InstalledAt, &inst.RemovedAt, &inst.CreatedAt, &inst.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan installation: %w", err)
		}
		installations = append(installations, &inst)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating installations: %w", err)
	}
	return installations, nil
}
