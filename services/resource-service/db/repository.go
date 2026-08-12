package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WorkUnit represents a row in the work_units table.
type WorkUnit struct {
	ID              string
	WorkCenterID    string
	Name            string
	Status          string
	PhysicalAssetID *string    // Nullable — a Work Unit may not have a Physical Asset linked
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// EquipmentRepository defines the data access methods for the ISA-95 Equipment Hierarchy.
type EquipmentRepository interface {
	CreateWorkUnit(ctx context.Context, workCenterID, name string) (*WorkUnit, error)
	GetWorkUnit(ctx context.Context, id string) (*WorkUnit, error)
	ListWorkUnits(ctx context.Context, workCenterID string) ([]*WorkUnit, error)
	UpdateWorkUnitStatus(ctx context.Context, id, expectedStatus, newStatus string) (*WorkUnit, error)
}

// PostgresEquipmentRepository implements EquipmentRepository using PostgreSQL.
type PostgresEquipmentRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresEquipmentRepository creates a new repository with the given connection pool.
func NewPostgresEquipmentRepository(pool *pgxpool.Pool) *PostgresEquipmentRepository {
	return &PostgresEquipmentRepository{pool: pool}
}

// CreateWorkUnit inserts a new Work Unit and returns the generated row.
func (r *PostgresEquipmentRepository) CreateWorkUnit(ctx context.Context, workCenterID, name string) (*WorkUnit, error) {
	query := `
		INSERT INTO work_units (work_center_id, name)
		VALUES ($1, $2)
		RETURNING id, work_center_id, name, status, physical_asset_id, created_at, updated_at`

	var wu WorkUnit
	err := r.pool.QueryRow(ctx, query, workCenterID, name).Scan(
		&wu.ID, &wu.WorkCenterID, &wu.Name, &wu.Status,
		&wu.PhysicalAssetID, &wu.CreatedAt, &wu.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create work unit: %w", err)
	}

	return &wu, nil
}

// GetWorkUnit retrieves a Work Unit by ID. Returns nil, nil if not found.
func (r *PostgresEquipmentRepository) GetWorkUnit(ctx context.Context, id string) (*WorkUnit, error) {
	query := `
		SELECT id, work_center_id, name, status, physical_asset_id, created_at, updated_at
		FROM work_units
		WHERE id = $1`

	var wu WorkUnit
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&wu.ID, &wu.WorkCenterID, &wu.Name, &wu.Status,
		&wu.PhysicalAssetID, &wu.CreatedAt, &wu.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get work unit: %w", err)
	}

	return &wu, nil
}

// ListWorkUnits retrieves all Work Units in a Work Center, ordered by creation time.
func (r *PostgresEquipmentRepository) ListWorkUnits(ctx context.Context, workCenterID string) ([]*WorkUnit, error) {
	query := `
		SELECT id, work_center_id, name, status, physical_asset_id, created_at, updated_at
		FROM work_units
		WHERE work_center_id = $1
		ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query, workCenterID)
	if err != nil {
		return nil, fmt.Errorf("failed to list work units: %w", err)
	}
	defer rows.Close()

	var workUnits []*WorkUnit

	for rows.Next() {
		var wu WorkUnit
		if err := rows.Scan(
			&wu.ID, &wu.WorkCenterID, &wu.Name, &wu.Status,
			&wu.PhysicalAssetID, &wu.CreatedAt, &wu.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan work unit: %w", err)
		}
		workUnits = append(workUnits, &wu)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating work units: %w", err)
	}

	return workUnits, nil
}

// UpdateWorkUnitStatus atomically changes a Work Unit's status only if the current status matches expectedStatus.
// Returns nil, nil if the Work Unit is not found or the current status doesn't match.
func (r *PostgresEquipmentRepository) UpdateWorkUnitStatus(ctx context.Context, id, expectedStatus, newStatus string) (*WorkUnit, error) {
	query := `
		UPDATE work_units
		SET status = $1, updated_at = now()
		WHERE id = $2 AND status = $3
		RETURNING id, work_center_id, name, status, physical_asset_id, created_at, updated_at`

	var wu WorkUnit
	err := r.pool.QueryRow(ctx, query, newStatus, id, expectedStatus).Scan(
		&wu.ID, &wu.WorkCenterID, &wu.Name, &wu.Status,
		&wu.PhysicalAssetID, &wu.CreatedAt, &wu.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to update work unit status: %w", err)
	}

	return &wu, nil
}
