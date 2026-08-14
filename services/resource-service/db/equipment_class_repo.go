package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// EquipmentClass represents a capability type (e.g., "CNC Lathe ≥ 5-axis").
type EquipmentClass struct {
	ID          string
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// EquipmentClassRepository defines data access methods for Equipment Classes.
type EquipmentClassRepository interface {
	CreateEquipmentClass(ctx context.Context, name string, description *string) (*EquipmentClass, error)
	GetEquipmentClass(ctx context.Context, id string) (*EquipmentClass, error)
	ListEquipmentClasses(ctx context.Context) ([]*EquipmentClass, error)
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
