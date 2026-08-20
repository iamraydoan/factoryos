package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// QualificationRecord certifies a Person for a PersonClass at a Work Center.
type QualificationRecord struct {
	ID            string
	PersonID      string
	PersonClassID string
	WorkCenterID  string
	CertifiedAt   time.Time
	ExpiresAt     *time.Time // nil = no expiry (permanent certification)
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// QualificationRepository defines data access methods for Qualification Records.
type QualificationRepository interface {
	QualifyPerson(ctx context.Context, personID, personClassID, workCenterID string, expiresAt *time.Time) (*QualificationRecord, error)
	GetQualification(ctx context.Context, id string) (*QualificationRecord, error)
	ListQualifications(ctx context.Context, personID, workCenterID string) ([]*QualificationRecord, error)
	RevokeQualification(ctx context.Context, id string) (bool, error)
	CheckExpiringQualifications(ctx context.Context, before time.Time) ([]*QualificationRecord, error)
}

// QualifyPerson creates or updates a Qualification Record (upsert on unique constraint).
func (r *PostgresEquipmentRepository) QualifyPerson(ctx context.Context, personID, personClassID, workCenterID string, expiresAt *time.Time) (*QualificationRecord, error) {
	query := `
		INSERT INTO qualification_records (person_id, person_class_id, work_center_id, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (person_id, person_class_id, work_center_id)
		DO UPDATE SET expires_at = EXCLUDED.expires_at, updated_at = now()
		RETURNING id, person_id, person_class_id, work_center_id, certified_at, expires_at, created_at, updated_at`

	var qr QualificationRecord
	err := r.pool.QueryRow(ctx, query, personID, personClassID, workCenterID, expiresAt).Scan(
		&qr.ID, &qr.PersonID, &qr.PersonClassID, &qr.WorkCenterID,
		&qr.CertifiedAt, &qr.ExpiresAt, &qr.CreatedAt, &qr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to qualify person: %w", err)
	}
	return &qr, nil
}

// GetQualification retrieves a Qualification Record by ID. Returns nil, nil if not found.
func (r *PostgresEquipmentRepository) GetQualification(ctx context.Context, id string) (*QualificationRecord, error) {
	query := `
		SELECT id, person_id, person_class_id, work_center_id, certified_at, expires_at, created_at, updated_at
		FROM qualification_records
		WHERE id = $1`

	var qr QualificationRecord
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&qr.ID, &qr.PersonID, &qr.PersonClassID, &qr.WorkCenterID,
		&qr.CertifiedAt, &qr.ExpiresAt, &qr.CreatedAt, &qr.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get qualification: %w", err)
	}
	return &qr, nil
}

// ListQualifications retrieves Qualification Records, optionally filtered by personID and/or workCenterID.
// Empty strings mean "don't filter on that dimension".
func (r *PostgresEquipmentRepository) ListQualifications(ctx context.Context, personID, workCenterID string) ([]*QualificationRecord, error) {
	var (
		rows pgx.Rows
		err  error
	)

	switch {
	case personID != "" && workCenterID != "":
		query := `
			SELECT id, person_id, person_class_id, work_center_id, certified_at, expires_at, created_at, updated_at
			FROM qualification_records
			WHERE person_id = $1 AND work_center_id = $2
			ORDER BY certified_at DESC`
		rows, err = r.pool.Query(ctx, query, personID, workCenterID)
	case personID != "":
		query := `
			SELECT id, person_id, person_class_id, work_center_id, certified_at, expires_at, created_at, updated_at
			FROM qualification_records
			WHERE person_id = $1
			ORDER BY certified_at DESC`
		rows, err = r.pool.Query(ctx, query, personID)
	case workCenterID != "":
		query := `
			SELECT id, person_id, person_class_id, work_center_id, certified_at, expires_at, created_at, updated_at
			FROM qualification_records
			WHERE work_center_id = $1
			ORDER BY certified_at DESC`
		rows, err = r.pool.Query(ctx, query, workCenterID)
	default:
		query := `
			SELECT id, person_id, person_class_id, work_center_id, certified_at, expires_at, created_at, updated_at
			FROM qualification_records
			ORDER BY certified_at DESC`
		rows, err = r.pool.Query(ctx, query)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list qualifications: %w", err)
	}
	defer rows.Close()

	var records []*QualificationRecord
	for rows.Next() {
		var qr QualificationRecord
		if err := rows.Scan(
			&qr.ID, &qr.PersonID, &qr.PersonClassID, &qr.WorkCenterID,
			&qr.CertifiedAt, &qr.ExpiresAt, &qr.CreatedAt, &qr.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan qualification: %w", err)
		}
		records = append(records, &qr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate qualifications: %w", err)
	}
	return records, nil
}

// RevokeQualification deletes a Qualification Record by ID. Returns true if deleted.
func (r *PostgresEquipmentRepository) RevokeQualification(ctx context.Context, id string) (bool, error) {
	query := `DELETE FROM qualification_records WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("failed to revoke qualification: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// CheckExpiringQualifications returns all qualification records whose expires_at
// is non-null and falls before the given time, ordered by expiry ascending.
func (r *PostgresEquipmentRepository) CheckExpiringQualifications(ctx context.Context, before time.Time) ([]*QualificationRecord, error) {
	query := `
		SELECT id, person_id, person_class_id, work_center_id, certified_at, expires_at, created_at, updated_at
		FROM qualification_records
		WHERE expires_at IS NOT NULL AND expires_at < $1
		ORDER BY expires_at ASC`

	rows, err := r.pool.Query(ctx, query, before)
	if err != nil {
		return nil, fmt.Errorf("failed to list expiring qualifications: %w", err)
	}
	defer rows.Close()

	var records []*QualificationRecord
	for rows.Next() {
		var qr QualificationRecord
		if err := rows.Scan(
			&qr.ID, &qr.PersonID, &qr.PersonClassID, &qr.WorkCenterID,
			&qr.CertifiedAt, &qr.ExpiresAt, &qr.CreatedAt, &qr.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan qualification: %w", err)
		}
		records = append(records, &qr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate qualifications: %w", err)
	}
	return records, nil
}
