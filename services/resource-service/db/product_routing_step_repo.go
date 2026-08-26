package db

import (
	"context"
	"fmt"
	"time"
)

// ProductRoutingStep represents a line item in a Product Routing Spec.
type ProductRoutingStep struct {
	ID                string
	RoutingSpecID     string
	WorkCenterID      string
	StepNumber        int32
	EstimatedDuration string
	Description       *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ProductRoutingStepRepository defines data access methods for Product Routing Steps.
type ProductRoutingStepRepository interface {
	AddRoutingStep(ctx context.Context, routingSpecID, workCenterID string, stepNumber int32, estimatedDuration string, description *string) (*ProductRoutingStep, error)
	ListRoutingSteps(ctx context.Context, routingSpecID string) ([]*ProductRoutingStep, error)
}

// AddRoutingStep inserts a new Product Routing Step and returns the generated row.
func (r *PostgresEquipmentRepository) AddRoutingStep(ctx context.Context, routingSpecID, workCenterID string, stepNumber int32, estimatedDuration string, description *string) (*ProductRoutingStep, error) {
	query := `
		INSERT INTO product_routing_steps (routing_spec_id, work_center_id, step_number, estimated_duration, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, routing_spec_id, work_center_id, step_number, estimated_duration, description, created_at, updated_at`

	var step ProductRoutingStep
	err := r.pool.QueryRow(ctx, query, routingSpecID, workCenterID, stepNumber, estimatedDuration, description).Scan(
		&step.ID, &step.RoutingSpecID, &step.WorkCenterID, &step.StepNumber,
		&step.EstimatedDuration, &step.Description, &step.CreatedAt, &step.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add routing step: %w", err)
	}
	return &step, nil
}

// ListRoutingSteps retrieves all steps for a given Routing Spec, ordered by step_number.
func (r *PostgresEquipmentRepository) ListRoutingSteps(ctx context.Context, routingSpecID string) ([]*ProductRoutingStep, error) {
	query := `
		SELECT id, routing_spec_id, work_center_id, step_number, estimated_duration, description, created_at, updated_at
		FROM product_routing_steps
		WHERE routing_spec_id = $1
		ORDER BY step_number`

	rows, err := r.pool.Query(ctx, query, routingSpecID)
	if err != nil {
		return nil, fmt.Errorf("failed to list routing steps: %w", err)
	}
	defer rows.Close()

	var steps []*ProductRoutingStep
	for rows.Next() {
		var step ProductRoutingStep
		if err := rows.Scan(&step.ID, &step.RoutingSpecID, &step.WorkCenterID, &step.StepNumber,
			&step.EstimatedDuration, &step.Description, &step.CreatedAt, &step.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan routing step: %w", err)
		}
		steps = append(steps, &step)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating routing steps: %w", err)
	}
	return steps, nil
}
