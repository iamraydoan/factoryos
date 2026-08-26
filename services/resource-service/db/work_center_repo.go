package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// WorkCenter represents a row in the work_centers table.
type WorkCenter struct {
	ID              string
	AreaID          string
	Name            string
	EquipmentClass  string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// WorkCenterRepository defines data access methods for Work Centers.
type WorkCenterRepository interface {
	GetWorkCenter(ctx context.Context, id string) (*WorkCenter, error)
}

// GetWorkCenter retrieves a Work Center by ID. Returns nil, nil if not found.
func (r *PostgresEquipmentRepository) GetWorkCenter(ctx context.Context, id string) (*WorkCenter, error) {
	query := `
		SELECT id, area_id, name, equipment_class, created_at, updated_at
		FROM work_centers
		WHERE id = $1`

	var wc WorkCenter
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&wc.ID, &wc.AreaID, &wc.Name, &wc.EquipmentClass, &wc.CreatedAt, &wc.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get work center: %w", err)
	}
	return &wc, nil
}
