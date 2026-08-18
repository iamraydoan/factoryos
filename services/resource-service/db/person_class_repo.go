package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// PersonClass represents a role category (e.g., "Operator", "Technician").
type PersonClass struct {
	ID          string
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PersonClassRepository defines data access methods for Person Classes.
type PersonClassRepository interface {
	CreatePersonClass(ctx context.Context, name string, description *string) (*PersonClass, error)
	GetPersonClass(ctx context.Context, id string) (*PersonClass, error)
	ListPersonClasses(ctx context.Context) ([]*PersonClass, error)
}

// CreatePersonClass inserts a new Person Class and returns the generated row.
func (r *PostgresEquipmentRepository) CreatePersonClass(ctx context.Context, name string, description *string) (*PersonClass, error) {
	query := `
		INSERT INTO person_classes (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at`

	var pc PersonClass
	err := r.pool.QueryRow(ctx, query, name, description).Scan(
		&pc.ID, &pc.Name, &pc.Description, &pc.CreatedAt, &pc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create person class: %w", err)
	}
	return &pc, nil
}

// GetPersonClass retrieves a Person Class by ID. Returns nil, nil if not found.
func (r *PostgresEquipmentRepository) GetPersonClass(ctx context.Context, id string) (*PersonClass, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM person_classes
		WHERE id = $1`

	var pc PersonClass
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&pc.ID, &pc.Name, &pc.Description, &pc.CreatedAt, &pc.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get person class: %w", err)
	}
	return &pc, nil
}

// ListPersonClasses retrieves all Person Classes, ordered by name.
func (r *PostgresEquipmentRepository) ListPersonClasses(ctx context.Context) ([]*PersonClass, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM person_classes
		ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list person classes: %w", err)
	}
	defer rows.Close()

	var classes []*PersonClass
	for rows.Next() {
		var pc PersonClass
		if err := rows.Scan(&pc.ID, &pc.Name, &pc.Description, &pc.CreatedAt, &pc.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan person class: %w", err)
		}
		classes = append(classes, &pc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating person classes: %w", err)
	}
	return classes, nil
}
