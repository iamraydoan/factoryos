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
	PhysicalAssetID *string // Nullable — a Work Unit may not have a Physical Asset linked
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// EquipmentClass represents a capability type (e.g., "CNC Lathe ≥ 5-axis").
type EquipmentClass struct {
	ID          string
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// WorkUnitCapability links a Work Unit to an Equipment Class with optional properties.
type WorkUnitCapability struct {
	ID               string
	WorkUnitID       string
	EquipmentClassID string
	Properties       map[string]interface{}
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// EquipmentRepository defines the data access methods for the ISA-95 Equipment Hierarchy.
type EquipmentRepository interface {
	CreateWorkUnit(ctx context.Context, workCenterID, name string) (*WorkUnit, error)
	GetWorkUnit(ctx context.Context, id string) (*WorkUnit, error)
	ListWorkUnits(ctx context.Context, workCenterID string) ([]*WorkUnit, error)
	UpdateWorkUnitStatus(ctx context.Context, id, expectedStatus, newStatus string) (*WorkUnit, error)

	CreateEquipmentClass(ctx context.Context, name string, description *string) (*EquipmentClass, error)
	GetEquipmentClass(ctx context.Context, id string) (*EquipmentClass, error)
	ListEquipmentClasses(ctx context.Context) ([]*EquipmentClass, error)

	AssignCapability(ctx context.Context, workUnitID, equipmentClassID string, properties map[string]interface{}) (*WorkUnitCapability, error)
	ListWorkUnitCapabilities(ctx context.Context, workUnitID string) ([]*WorkUnitCapability, error)
	RemoveCapability(ctx context.Context, workUnitID, equipmentClassID string) (bool, error)
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

// CreateEquipmentClass inserts a new Equipment Class and returns the generated row.
func (r *PostgresEquipmentRepository) CreateEquipmentClass(ctx context.Context, name string, description *string) (*EquipmentClass, error) {
	query := `
		INSERT INTO equipment_classes (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at`

	var ec EquipmentClass
	err := r.pool.QueryRow(ctx, query, name, description).Scan(
		&ec.ID, &ec.Name, &ec.Description, &ec.CreatedAt, &ec.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create equipment class: %w", err)
	}
	return &ec, nil
}

// GetEquipmentClass retrieves an Equipment Class by ID. Returns nil, nil if not found.
func (r *PostgresEquipmentRepository) GetEquipmentClass(ctx context.Context, id string) (*EquipmentClass, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM equipment_classes
		WHERE id = $1`

	var ec EquipmentClass
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ec.ID, &ec.Name, &ec.Description, &ec.CreatedAt, &ec.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get equipment class: %w", err)
	}
	return &ec, nil
}
// ListEquipmentClasses retrieves all Equipment Classes, ordered by name.
func (r *PostgresEquipmentRepository) ListEquipmentClasses(ctx context.Context) ([]*EquipmentClass, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM equipment_classes
		ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list equipment classes: %w", err)
	}
	defer rows.Close()

	var classes []*EquipmentClass
	for rows.Next() {
		var ec EquipmentClass
		if err := rows.Scan(&ec.ID, &ec.Name, &ec.Description, &ec.CreatedAt, &ec.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan equipment class: %w", err)
		}
		classes = append(classes, &ec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating equipment classes: %w", err)
	}
	return classes, nil
}

// AssignCapability links a Work Unit to an Equipment Class with optional properties.
func (r *PostgresEquipmentRepository) AssignCapability(ctx context.Context, workUnitID, equipmentClassID string, properties map[string]interface{}) (*WorkUnitCapability, error) {
	query := `
		INSERT INTO work_unit_capabilities (work_unit_id, equipment_class_id, properties)
		VALUES ($1, $2, $3)
		RETURNING id, work_unit_id, equipment_class_id, properties, created_at, updated_at`

	var wuc WorkUnitCapability
	err := r.pool.QueryRow(ctx, query, workUnitID, equipmentClassID, properties).Scan(
		&wuc.ID, &wuc.WorkUnitID, &wuc.EquipmentClassID, &wuc.Properties, &wuc.CreatedAt, &wuc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to assign capability: %w", err)
	}
	return &wuc, nil
}

// ListWorkUnitCapabilities retrieves all capabilities for a Work Unit, ordered by creation time.
func (r *PostgresEquipmentRepository) ListWorkUnitCapabilities(ctx context.Context, workUnitID string) ([]*WorkUnitCapability, error) {
	query := `
		SELECT id, work_unit_id, equipment_class_id, properties, created_at, updated_at
		FROM work_unit_capabilities
		WHERE work_unit_id = $1
		ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query, workUnitID)
	if err != nil {
		return nil, fmt.Errorf("failed to list capabilities: %w", err)
	}
	defer rows.Close()

	var capabilities []*WorkUnitCapability
	for rows.Next() {
		var wuc WorkUnitCapability
		if err := rows.Scan(
			&wuc.ID, &wuc.WorkUnitID, &wuc.EquipmentClassID, &wuc.Properties, &wuc.CreatedAt, &wuc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan capability: %w", err)
		}
		capabilities = append(capabilities, &wuc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating capabilities: %w", err)
	}
	return capabilities, nil
}
// RemoveCapability removes a Work Unit ↔ Equipment Class link. Returns true if removed.
func (r *PostgresEquipmentRepository) RemoveCapability(ctx context.Context, workUnitID, equipmentClassID string) (bool, error) {
	query := `
		DELETE FROM work_unit_capabilities
		WHERE work_unit_id = $1 AND equipment_class_id = $2`

	tag, err := r.pool.Exec(ctx, query, workUnitID, equipmentClassID)
	if err != nil {
		return false, fmt.Errorf("failed to remove capability: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}
