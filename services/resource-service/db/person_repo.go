package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Person represents an employee who can be assigned to work units.
type Person struct {
	ID            string
	PersonClassID string
	EmployeeID    string
	FirstName     string
	LastName      string
	Email         *string
	Status        string // "active", "inactive", "on_leave"
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// PersonRepository defines data access methods for Persons.
type PersonRepository interface {
	CreatePerson(ctx context.Context, personClassID, employeeID, firstName, lastName string, email *string) (*Person, error)
	GetPerson(ctx context.Context, id string) (*Person, error)
	ListPersons(ctx context.Context, personClassID string) ([]*Person, error)
}

// CreatePerson inserts a new Person and returns the generated row.
func (r *PostgresEquipmentRepository) CreatePerson(ctx context.Context, personClassID, employeeID, firstName, lastName string, email *string) (*Person, error) {
	query := `
		INSERT INTO persons (person_class_id, employee_id, first_name, last_name, email)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, person_class_id, employee_id, first_name, last_name, email, status, created_at, updated_at`

	var p Person
	err := r.pool.QueryRow(ctx, query, personClassID, employeeID, firstName, lastName, email).Scan(
		&p.ID, &p.PersonClassID, &p.EmployeeID, &p.FirstName, &p.LastName,
		&p.Email, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create person: %w", err)
	}
	return &p, nil
}

// GetPerson retrieves a Person by ID. Returns nil, nil if not found.
func (r *PostgresEquipmentRepository) GetPerson(ctx context.Context, id string) (*Person, error) {
	query := `
		SELECT id, person_class_id, employee_id, first_name, last_name, email, status, created_at, updated_at
		FROM persons
		WHERE id = $1`

	var p Person
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.PersonClassID, &p.EmployeeID, &p.FirstName, &p.LastName,
		&p.Email, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get person: %w", err)
	}
	return &p, nil
}

// ListPersons retrieves Persons, optionally filtered by PersonClassID.
// If personClassID is empty, returns all persons ordered by last name.
func (r *PostgresEquipmentRepository) ListPersons(ctx context.Context, personClassID string) ([]*Person, error) {
	var rows pgx.Rows
	var err error

	if personClassID != "" {
		query := `
			SELECT id, person_class_id, employee_id, first_name, last_name, email, status, created_at, updated_at
			FROM persons
			WHERE person_class_id = $1
			ORDER BY last_name, first_name`
		rows, err = r.pool.Query(ctx, query, personClassID)
	} else {
		query := `
			SELECT id, person_class_id, employee_id, first_name, last_name, email, status, created_at, updated_at
			FROM persons
			ORDER BY last_name, first_name`
		rows, err = r.pool.Query(ctx, query)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list persons: %w", err)
	}
	defer rows.Close()

	var persons []*Person
	for rows.Next() {
		var p Person
		if err := rows.Scan(&p.ID, &p.PersonClassID, &p.EmployeeID, &p.FirstName, &p.LastName,
			&p.Email, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan person: %w", err)
		}
		persons = append(persons, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating persons: %w", err)
	}
	return persons, nil
}
