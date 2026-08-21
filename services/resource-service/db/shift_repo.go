package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Shift defines a named work shift with start and end times.
type Shift struct {
	ID          string
	Name        string
	StartTime   string
	EndTime     string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ShiftRepository defines data access methods for Shifts.
type ShiftRepository interface {
	CreateShift(ctx context.Context, name, startTime, endTime string, description *string) (*Shift, error)
	GetShift(ctx context.Context, id string) (*Shift, error)
	ListShifts(ctx context.Context) ([]*Shift, error)
}

// CreateShift inserts a new Shift and returns the generated row.
func (r *PostgresEquipmentRepository) CreateShift(ctx context.Context, name, startTime, endTime string, description *string) (*Shift, error) {
	query := `
		INSERT INTO shifts (name, start_time, end_time, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, start_time, end_time, description, created_at, updated_at`

	var s Shift
	err := r.pool.QueryRow(ctx, query, name, startTime, endTime, description).Scan(
		&s.ID, &s.Name, &s.StartTime, &s.EndTime, &s.Description, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create shift: %w", err)
	}
	return &s, nil
}

// GetShift retrieves a Shift by ID. Returns nil, nil if not found.
func (r *PostgresEquipmentRepository) GetShift(ctx context.Context, id string) (*Shift, error) {
	query := `
		SELECT id, name, start_time, end_time, description, created_at, updated_at
		FROM shifts
		WHERE id = $1`

	var s Shift
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.StartTime, &s.EndTime, &s.Description, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get shift: %w", err)
	}
	return &s, nil
}

// ListShifts retrieves all Shifts, ordered by name.
func (r *PostgresEquipmentRepository) ListShifts(ctx context.Context) ([]*Shift, error) {
	query := `
		SELECT id, name, start_time, end_time, description, created_at, updated_at
		FROM shifts
		ORDER BY created_at`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list shifts: %w", err)
	}
	defer rows.Close()

	var shifts []*Shift
	for rows.Next() {
		var s Shift
		if err := rows.Scan(&s.ID, &s.Name, &s.StartTime, &s.EndTime, &s.Description, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan shift: %w", err)
		}
		shifts = append(shifts, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating shifts: %w", err)
	}
	return shifts, nil
}
