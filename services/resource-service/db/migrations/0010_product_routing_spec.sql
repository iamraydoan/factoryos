-- +goose Up
CREATE TABLE IF NOT EXISTS product_routing_specs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    material_definition_id UUID NOT NULL REFERENCES material_definitions(id) ON DELETE RESTRICT,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(material_definition_id, version)
);

CREATE TABLE IF NOT EXISTS product_routing_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    routing_spec_id UUID NOT NULL REFERENCES product_routing_specs(id) ON DELETE CASCADE,
    work_center_id UUID NOT NULL REFERENCES work_centers(id) ON DELETE RESTRICT,
    step_number INTEGER NOT NULL,          -- sequence order (1, 2, 3, ...); unique per routing spec
    estimated_duration VARCHAR(50) NOT NULL, -- human-readable duration (e.g., "45m", "1h30m")
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(routing_spec_id, step_number)
);

CREATE INDEX idx_routing_spec_material ON product_routing_specs(material_definition_id);
CREATE INDEX idx_routing_steps_spec ON product_routing_steps(routing_spec_id);
CREATE INDEX idx_routing_steps_work_center ON product_routing_steps(work_center_id);

-- +goose Down
DROP TABLE IF EXISTS product_routing_steps;
DROP TABLE IF EXISTS product_routing_specs;
