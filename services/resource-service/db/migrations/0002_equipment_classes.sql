-- +goose Up
CREATE TABLE IF NOT EXISTS equipment_classes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS work_unit_capabilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    work_unit_id UUID NOT NULL REFERENCES work_units(id) ON DELETE CASCADE,
    equipment_class_id UUID NOT NULL REFERENCES equipment_classes(id) ON DELETE CASCADE,
    properties JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(work_unit_id, equipment_class_id)
);

CREATE INDEX idx_work_unit_capabilities_work_unit ON work_unit_capabilities(work_unit_id);
CREATE INDEX idx_work_unit_capabilities_equipment_class ON work_unit_capabilities(equipment_class_id);

-- +goose Down
DROP TABLE IF EXISTS work_unit_capabilities;
DROP TABLE IF EXISTS equipment_classes;
