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

	// Physical Asset methods
	CreatePhysicalAsset(ctx context.Context, asset *PhysicalAsset) (*PhysicalAsset, error)
	GetPhysicalAsset(ctx context.Context, id string) (*PhysicalAsset, error)
	ListPhysicalAssets(ctx context.Context) ([]*PhysicalAsset, error)

	// Installation methods
	InstallAsset(ctx context.Context, physicalAssetID, workUnitID string) (*PhysicalAssetInstallation, error)
	UninstallAsset(ctx context.Context, workUnitID string) (*PhysicalAssetInstallation, error)
	GetCurrentInstallation(ctx context.Context, workUnitID string) (*PhysicalAssetInstallation, error)
	ListInstallations(ctx context.Context, workUnitID string) ([]*PhysicalAssetInstallation, error)
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

// ============================================================================
// Physical Asset Methods
// ============================================================================

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

// ============================================================================
// Installation Methods
// ============================================================================

// InstallAsset installs a Physical Asset at a Work Unit within a transaction.
// Enforces: one asset per work unit, one work unit per asset (at a time).
func (r *PostgresEquipmentRepository) InstallAsset(ctx context.Context, physicalAssetID, workUnitID string) (*PhysicalAssetInstallation, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // No-op after successful commit

	// Guard: asset must not be installed elsewhere
	var occupiedWU *string
	err = tx.QueryRow(ctx,
		`SELECT work_unit_id FROM physical_asset_installations
		 WHERE physical_asset_id = $1 AND removed_at IS NULL`,
		physicalAssetID,
	).Scan(&occupiedWU)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing installation: %w", err)
	}
	if occupiedWU != nil {
		return nil, fmt.Errorf("asset is already installed at work unit %s", *occupiedWU)
	}

	// Guard: work unit must not already have an active installation
	var occupiedAsset *string
	err = tx.QueryRow(ctx,
		`SELECT physical_asset_id FROM physical_asset_installations
		 WHERE work_unit_id = $1 AND removed_at IS NULL`,
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
