package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// ShiftAssignment links a Person to a Shift at a Work Center.
type ShiftAssignment struct {
	ID            string
	PersonID      string
	ShiftID       string
	WorkCenterID  string
	EffectiveFrom time.Time
	EffectiveTo   *time.Time // nil = open-ended (currently assigned)
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ShiftAssignmentRepository defines data access methods for Shift Assignments.
type ShiftAssignmentRepository interface {
	AssignShift(ctx context.Context, personID, shiftID, workCenterID string, effectiveFrom time.Time) (*ShiftAssignment, error)
	GetShiftAssignment(ctx context.Context, id string) (*ShiftAssignment, error)
	ListShiftAssignments(ctx context.Context, personID, shiftID, workCenterID string) ([]*ShiftAssignment, error)
	UnassignShift(ctx context.Context, id string) (bool, error)
}

// AssignShift creates or updates a Shift Assignment (upsert on unique constraint).
func (r *PostgresEquipmentRepository) AssignShift(ctx context.Context, personID, shiftID, workCenterID string, effectiveFrom time.Time) (*ShiftAssignment, error) {
	query := `
		INSERT INTO shift_assignments (person_id, shift_id, work_center_id, effective_from)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (person_id, shift_id, work_center_id)
		DO UPDATE SET effective_from = EXCLUDED.effective_from, effective_to = NULL, updated_at = now()
		RETURNING id, person_id, shift_id, work_center_id, effective_from, effective_to, created_at, updated_at`

	var sa ShiftAssignment
	err := r.pool.QueryRow(ctx, query, personID, shiftID, workCenterID, effectiveFrom).Scan(
		&sa.ID, &sa.PersonID, &sa.ShiftID, &sa.WorkCenterID,
		&sa.EffectiveFrom, &sa.EffectiveTo, &sa.CreatedAt, &sa.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to assign shift: %w", err)
	}
	return &sa, nil
}

// GetShiftAssignment retrieves a Shift Assignment by ID. Returns nil, nil if not found.
func (r *PostgresEquipmentRepository) GetShiftAssignment(ctx context.Context, id string) (*ShiftAssignment, error) {
	query := `
		SELECT id, person_id, shift_id, work_center_id, effective_from, effective_to, created_at, updated_at
		FROM shift_assignments
		WHERE id = $1`

	var sa ShiftAssignment
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&sa.ID, &sa.PersonID, &sa.ShiftID, &sa.WorkCenterID,
		&sa.EffectiveFrom, &sa.EffectiveTo, &sa.CreatedAt, &sa.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get shift assignment: %w", err)
	}
	return &sa, nil
}

// ListShiftAssignments retrieves Shift Assignments, optionally filtered by personID, shiftID, and/or workCenterID.
// Empty strings mean "don't filter on that dimension".
func (r *PostgresEquipmentRepository) ListShiftAssignments(ctx context.Context, personID, shiftID, workCenterID string) ([]*ShiftAssignment, error) {
	query := `
		SELECT id, person_id, shift_id, work_center_id, effective_from, effective_to, created_at, updated_at
		FROM shift_assignments`

	var (
		conditions []string
		args       []any
	)

	if personID != "" {
		args = append(args, personID)
		conditions = append(conditions, fmt.Sprintf("person_id = $%d", len(args)))
	}

	if shiftID != "" {
		args = append(args, shiftID)
		conditions = append(conditions, fmt.Sprintf("shift_id = $%d", len(args)))
	}

	if workCenterID != "" {
		args = append(args, workCenterID)
		conditions = append(conditions, fmt.Sprintf("work_center_id = $%d", len(args)))
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY effective_from DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list shift assignments: %w", err)
	}
	defer rows.Close()

	var assignments []*ShiftAssignment
	for rows.Next() {
		var sa ShiftAssignment
		if err := rows.Scan(
			&sa.ID, &sa.PersonID, &sa.ShiftID, &sa.WorkCenterID,
			&sa.EffectiveFrom, &sa.EffectiveTo, &sa.CreatedAt, &sa.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan shift assignment: %w", err)
		}
		assignments = append(assignments, &sa)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating shift assignments: %w", err)
	}
	return assignments, nil
}

// UnassignShift deletes a Shift Assignment by ID. Returns true if deleted.
func (r *PostgresEquipmentRepository) UnassignShift(ctx context.Context, id string) (bool, error) {
	query := `DELETE FROM shift_assignments WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("failed to unassign shift: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}
