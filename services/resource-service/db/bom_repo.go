package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// BillOfMaterials represents a versioned list of components for a parent material.
type BillOfMaterials struct {
	ID                   string
	MaterialDefinitionID string
	Version              string
	Description          *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// BOMRepository defines data access methods for Bills of Materials.
type BOMRepository interface {
	CreateBOM(ctx context.Context, materialDefinitionID, version string, description *string) (*BillOfMaterials, error)
	GetBOM(ctx context.Context, id string) (*BillOfMaterials, error)
	ListBOMs(ctx context.Context, materialDefinitionID string) ([]*BillOfMaterials, error)
}

// CreateBOM inserts a new Bill of Materials and returns the generated row.
func (r *PostgresEquipmentRepository) CreateBOM(ctx context.Context, materialDefinitionID, version string, description *string) (*BillOfMaterials, error) {
	query := `
		INSERT INTO bill_of_materials (material_definition_id, version, description)
		VALUES ($1, $2, $3)
		RETURNING id, material_definition_id, version, description, created_at, updated_at`

	var bom BillOfMaterials
	err := r.pool.QueryRow(ctx, query, materialDefinitionID, version, description).Scan(
		&bom.ID, &bom.MaterialDefinitionID, &bom.Version, &bom.Description,
		&bom.CreatedAt, &bom.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create BOM: %w", err)
	}
	return &bom, nil
}

// GetBOM retrieves a Bill of Materials by ID. Returns nil, nil if not found.
func (r *PostgresEquipmentRepository) GetBOM(ctx context.Context, id string) (*BillOfMaterials, error) {
	query := `
		SELECT id, material_definition_id, version, description, created_at, updated_at
		FROM bill_of_materials
		WHERE id = $1`

	var bom BillOfMaterials
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&bom.ID, &bom.MaterialDefinitionID, &bom.Version, &bom.Description,
		&bom.CreatedAt, &bom.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get BOM: %w", err)
	}
	return &bom, nil
}

// ListBOMs retrieves Bills of Materials, optionally filtered by MaterialDefinitionID.
// If materialDefinitionID is empty, returns all BOMs ordered by creation date.
func (r *PostgresEquipmentRepository) ListBOMs(ctx context.Context, materialDefinitionID string) ([]*BillOfMaterials, error) {
	var rows pgx.Rows
	var err error

	if materialDefinitionID != "" {
		query := `
			SELECT id, material_definition_id, version, description, created_at, updated_at
			FROM bill_of_materials
			WHERE material_definition_id = $1
			ORDER BY created_at DESC`
		rows, err = r.pool.Query(ctx, query, materialDefinitionID)
	} else {
		query := `
			SELECT id, material_definition_id, version, description, created_at, updated_at
			FROM bill_of_materials
			ORDER BY created_at DESC`
		rows, err = r.pool.Query(ctx, query)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list BOMs: %w", err)
	}
	defer rows.Close()

	var boms []*BillOfMaterials
	for rows.Next() {
		var bom BillOfMaterials
		if err := rows.Scan(&bom.ID, &bom.MaterialDefinitionID, &bom.Version, &bom.Description,
			&bom.CreatedAt, &bom.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan BOM: %w", err)
		}
		boms = append(boms, &bom)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating BOMs: %w", err)
	}
	return boms, nil
}
