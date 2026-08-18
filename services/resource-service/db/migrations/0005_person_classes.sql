-- +goose Up
CREATE TABLE IF NOT EXISTS person_classes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE person_status AS ENUM ('active', 'inactive', 'on_leave');

CREATE TABLE IF NOT EXISTS persons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    person_class_id UUID NOT NULL REFERENCES person_classes(id) ON DELETE RESTRICT,
    employee_id VARCHAR(100) NOT NULL UNIQUE,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    status person_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_persons_person_class ON persons(person_class_id);
CREATE INDEX idx_persons_employee_id ON persons(employee_id);

-- +goose Down
DROP TABLE IF EXISTS persons;
DROP TYPE IF EXISTS person_status;
DROP TABLE IF EXISTS person_classes;