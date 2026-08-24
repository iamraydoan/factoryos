-- +goose Up
CREATE TABLE IF NOT EXISTS material_classes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS material_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    material_class_id UUID NOT NULL REFERENCES material_classes(id) ON DELETE RESTRICT,
    name VARCHAR(255) NOT NULL,
    part_number VARCHAR(100) NOT NULL UNIQUE,
    unit_of_measure VARCHAR(50) NOT NULL,
    specification TEXT,                                 -- optional JSON spec (grade, tolerance, etc.)
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_material_definitions_class ON material_definitions(material_class_id);
CREATE INDEX idx_material_definitions_part_number ON material_definitions(part_number);

-- +goose Down
DROP TABLE IF EXISTS material_definitions;
DROP TABLE IF EXISTS material_classes;
