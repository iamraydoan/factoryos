package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type MaterialClass struct {
	ID          string
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MaterialClassRepository interface {
	CreateMaterialClass(ctx context.Context, name string, description *string) (*MaterialClass, error)
	GetMaterialClass(ctx context.Context, id string) (*MaterialClass, error)
	ListMaterialClasses(ctx context.Context) ([]*MaterialClass, error)
}

func (r *PostgresEquipmentRepository) CreateMaterialClass(ctx context.Context, name string, description *string) (*MaterialClass, error) {
	query := `
		INSERT INTO material_classes (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at`

	var mc MaterialClass
	err := r.pool.QueryRow(ctx, query, name, description).Scan(
		&mc.ID, &mc.Name, &mc.Description, &mc.CreatedAt, &mc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create material class: %w", err)
	}
	return &mc, nil
}

// GetMaterialClass retrieves a Material Class by ID. Returns nil, nil if not found.
func (r *PostgresEquipmentRepository) GetMaterialClass(ctx context.Context, id string) (*MaterialClass, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM material_classes
		WHERE id = $1`

	var mc MaterialClass
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&mc.ID, &mc.Name, &mc.Description, &mc.CreatedAt, &mc.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get material class: %w", err)
	}
	return &mc, nil
}

// ListMaterialClasses retrieves all Material Classes, ordered by name.
func (r *PostgresEquipmentRepository) ListMaterialClasses(ctx context.Context) ([]*MaterialClass, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM material_classes
		ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list material classes: %w", err)
	}
	defer rows.Close()

	var classes []*MaterialClass
	for rows.Next() {
		var mc MaterialClass
		if err := rows.Scan(&mc.ID, &mc.Name, &mc.Description, &mc.CreatedAt, &mc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan material class: %w", err)
		}
		classes = append(classes, &mc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating material classes: %w", err)
	}
	return classes, nil
}
