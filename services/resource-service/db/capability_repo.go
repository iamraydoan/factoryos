package db

import (
	"context"
	"fmt"
	"time"
)

// WorkUnitCapability links a Work Unit to an Equipment Class with optional properties.
type WorkUnitCapability struct {
	ID               string
	WorkUnitID       string
	EquipmentClassID string
	Properties       map[string]interface{}
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// CapabilityRepository defines data access methods for Work Unit capabilities.
type CapabilityRepository interface {
	AssignCapability(ctx context.Context, workUnitID, equipmentClassID string, properties map[string]interface{}) (*WorkUnitCapability, error)
	ListWorkUnitCapabilities(ctx context.Context, workUnitID string) ([]*WorkUnitCapability, error)
	RemoveCapability(ctx context.Context, workUnitID, equipmentClassID string) (bool, error)
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
