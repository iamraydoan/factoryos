package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// MaterialDefinition represents a specific material or product with its specification.
type MaterialDefinition struct {
	ID              string
	MaterialClassID string
	Name            string
	PartNumber      string
	UnitOfMeasure   string
	Specification   *string // optional JSON spec (grade, tolerance, etc.)
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// MaterialDefinitionRepository defines data access methods for Material Definitions.
type MaterialDefinitionRepository interface {
	CreateMaterialDefinition(ctx context.Context, materialClassID, name, partNumber, unitOfMeasure string, specification *string) (*MaterialDefinition, error)
	GetMaterialDefinition(ctx context.Context, id string) (*MaterialDefinition, error)
	ListMaterialDefinitions(ctx context.Context, materialClassID string) ([]*MaterialDefinition, error)
}

// CreateMaterialDefinition inserts a new Material Definition and returns the generated row.
func (r *PostgresEquipmentRepository) CreateMaterialDefinition(ctx context.Context, materialClassID, name, partNumber, unitOfMeasure string, specification *string) (*MaterialDefinition, error) {
	query := `
		INSERT INTO material_definitions (material_class_id, name, part_number, unit_of_measure, specification)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, material_class_id, name, part_number, unit_of_measure, specification, created_at, updated_at`

	var md MaterialDefinition
	err := r.pool.QueryRow(ctx, query, materialClassID, name, partNumber, unitOfMeasure, specification).Scan(
		&md.ID, &md.MaterialClassID, &md.Name, &md.PartNumber, &md.UnitOfMeasure,
		&md.Specification, &md.CreatedAt, &md.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create material definition: %w", err)
	}
	return &md, nil
}

// GetMaterialDefinition retrieves a Material Definition by ID. Returns nil, nil if not found.
func (r *PostgresEquipmentRepository) GetMaterialDefinition(ctx context.Context, id string) (*MaterialDefinition, error) {
	query := `
		SELECT id, material_class_id, name, part_number, unit_of_measure, specification, created_at, updated_at
		FROM material_definitions
		WHERE id = $1`

	var md MaterialDefinition
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&md.ID, &md.MaterialClassID, &md.Name, &md.PartNumber, &md.UnitOfMeasure,
		&md.Specification, &md.CreatedAt, &md.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get material definition: %w", err)
	}
	return &md, nil
}

// ListMaterialDefinitions retrieves Material Definitions, optionally filtered by MaterialClassID.
// If materialClassID is empty, returns all material definitions ordered by name.
func (r *PostgresEquipmentRepository) ListMaterialDefinitions(ctx context.Context, materialClassID string) ([]*MaterialDefinition, error) {
	var rows pgx.Rows
	var err error

	if materialClassID != "" {
		query := `
			SELECT id, material_class_id, name, part_number, unit_of_measure, specification, created_at, updated_at
			FROM material_definitions
			WHERE material_class_id = $1
			ORDER BY name`
		rows, err = r.pool.Query(ctx, query, materialClassID)
	} else {
		query := `
			SELECT id, material_class_id, name, part_number, unit_of_measure, specification, created_at, updated_at
			FROM material_definitions
			ORDER BY name`
		rows, err = r.pool.Query(ctx, query)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list material definitions: %w", err)
	}
	defer rows.Close()

	var definitions []*MaterialDefinition
	for rows.Next() {
		var md MaterialDefinition
		if err := rows.Scan(&md.ID, &md.MaterialClassID, &md.Name, &md.PartNumber, &md.UnitOfMeasure,
			&md.Specification, &md.CreatedAt, &md.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan material definition: %w", err)
		}
		definitions = append(definitions, &md)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating material definitions: %w", err)
	}
	return definitions, nil
}
