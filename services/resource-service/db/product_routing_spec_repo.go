package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ProductRoutingSpec represents a versioned sequence of Work Centers for producing a material.
type ProductRoutingSpec struct {
	ID                   string
	MaterialDefinitionID string
	Version              string
	Description          *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// ProductRoutingSpecRepository defines data access methods for Product Routing Specs.
type ProductRoutingSpecRepository interface {
	CreateRoutingSpec(ctx context.Context, materialDefinitionID, version string, description *string) (*ProductRoutingSpec, error)
	GetRoutingSpec(ctx context.Context, id string) (*ProductRoutingSpec, error)
	ListRoutingSpecs(ctx context.Context, materialDefinitionID string) ([]*ProductRoutingSpec, error)
}

// CreateRoutingSpec inserts a new Product Routing Spec and returns the generated row.
func (r *PostgresEquipmentRepository) CreateRoutingSpec(ctx context.Context, materialDefinitionID, version string, description *string) (*ProductRoutingSpec, error) {
	query := `
		INSERT INTO product_routing_specs (material_definition_id, version, description)
		VALUES ($1, $2, $3)
		RETURNING id, material_definition_id, version, description, created_at, updated_at`

	var spec ProductRoutingSpec
	err := r.pool.QueryRow(ctx, query, materialDefinitionID, version, description).Scan(
		&spec.ID, &spec.MaterialDefinitionID, &spec.Version, &spec.Description,
		&spec.CreatedAt, &spec.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create routing spec: %w", err)
	}
	return &spec, nil
}

// GetRoutingSpec retrieves a Product Routing Spec by ID. Returns nil, nil if not found.
func (r *PostgresEquipmentRepository) GetRoutingSpec(ctx context.Context, id string) (*ProductRoutingSpec, error) {
	query := `
		SELECT id, material_definition_id, version, description, created_at, updated_at
		FROM product_routing_specs
		WHERE id = $1`

	var spec ProductRoutingSpec
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&spec.ID, &spec.MaterialDefinitionID, &spec.Version, &spec.Description,
		&spec.CreatedAt, &spec.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get routing spec: %w", err)
	}
	return &spec, nil
}

// ListRoutingSpecs retrieves Product Routing Specs, optionally filtered by MaterialDefinitionID.
// If materialDefinitionID is empty, returns all routing specs ordered by creation date.
func (r *PostgresEquipmentRepository) ListRoutingSpecs(ctx context.Context, materialDefinitionID string) ([]*ProductRoutingSpec, error) {
	var rows pgx.Rows
	var err error

	if materialDefinitionID != "" {
		query := `
			SELECT id, material_definition_id, version, description, created_at, updated_at
			FROM product_routing_specs
			WHERE material_definition_id = $1
			ORDER BY created_at DESC`
		rows, err = r.pool.Query(ctx, query, materialDefinitionID)
	} else {
		query := `
			SELECT id, material_definition_id, version, description, created_at, updated_at
			FROM product_routing_specs
			ORDER BY created_at DESC`
		rows, err = r.pool.Query(ctx, query)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list routing specs: %w", err)
	}
	defer rows.Close()

	var specs []*ProductRoutingSpec
	for rows.Next() {
		var spec ProductRoutingSpec
		if err := rows.Scan(&spec.ID, &spec.MaterialDefinitionID, &spec.Version, &spec.Description,
			&spec.CreatedAt, &spec.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan routing spec: %w", err)
		}
		specs = append(specs, &spec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating routing specs: %w", err)
	}
	return specs, nil
}
