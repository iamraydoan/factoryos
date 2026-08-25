package db

import (
	"context"
	"fmt"
	"time"
)

// BOMComponent represents a line item in a Bill of Materials.
type BOMComponent struct {
	ID                   string
	BOMID                string
	MaterialDefinitionID string
	Quantity             string
	UnitOfMeasure        string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// BOMComponentRepository defines data access methods for BOM Components.
type BOMComponentRepository interface {
	AddBOMComponent(ctx context.Context, bomID, materialDefinitionID, quantity, unitOfMeasure string) (*BOMComponent, error)
	ListBOMComponents(ctx context.Context, bomID string) ([]*BOMComponent, error)
}

// AddBOMComponent inserts a new BOM Component and returns the generated row.
func (r *PostgresEquipmentRepository) AddBOMComponent(ctx context.Context, bomID, materialDefinitionID, quantity, unitOfMeasure string) (*BOMComponent, error) {
	query := `
		INSERT INTO bom_components (bom_id, material_definition_id, quantity, unit_of_measure)
		VALUES ($1, $2, $3, $4)
		RETURNING id, bom_id, material_definition_id, quantity, unit_of_measure, created_at, updated_at`

	var comp BOMComponent
	err := r.pool.QueryRow(ctx, query, bomID, materialDefinitionID, quantity, unitOfMeasure).Scan(
		&comp.ID, &comp.BOMID, &comp.MaterialDefinitionID, &comp.Quantity,
		&comp.UnitOfMeasure, &comp.CreatedAt, &comp.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add BOM component: %w", err)
	}
	return &comp, nil
}

// ListBOMComponents retrieves all components for a given BOM, ordered by creation date.
func (r *PostgresEquipmentRepository) ListBOMComponents(ctx context.Context, bomID string) ([]*BOMComponent, error) {
	query := `
		SELECT id, bom_id, material_definition_id, quantity, unit_of_measure, created_at, updated_at
		FROM bom_components
		WHERE bom_id = $1
		ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query, bomID)
	if err != nil {
		return nil, fmt.Errorf("failed to list BOM components: %w", err)
	}
	defer rows.Close()

	var components []*BOMComponent
	for rows.Next() {
		var comp BOMComponent
		if err := rows.Scan(&comp.ID, &comp.BOMID, &comp.MaterialDefinitionID, &comp.Quantity,
			&comp.UnitOfMeasure, &comp.CreatedAt, &comp.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan BOM component: %w", err)
		}
		components = append(components, &comp)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating BOM components: %w", err)
	}
	return components, nil
}
