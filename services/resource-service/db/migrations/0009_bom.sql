-- +goose Up
CREATE TABLE IF NOT EXISTS bill_of_materials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    material_definition_id UUID NOT NULL REFERENCES material_definitions(id) ON DELETE RESTRICT,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(material_definition_id, version)
);

CREATE TABLE IF NOT EXISTS bom_components (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bom_id UUID NOT NULL REFERENCES bill_of_materials(id) ON DELETE CASCADE,
    material_definition_id UUID NOT NULL REFERENCES material_definitions(id) ON DELETE RESTRICT,
    quantity VARCHAR(50) NOT NULL,       -- decimal string to avoid floating-point precision issues
    unit_of_measure VARCHAR(50) NOT NULL, -- e.g., "kg", "pcs", "liters"
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_bom_material ON bill_of_materials(material_definition_id);
CREATE INDEX idx_bom_components_bom ON bom_components(bom_id);
CREATE INDEX idx_bom_components_material ON bom_components(material_definition_id);

-- +goose Down
DROP TABLE IF EXISTS bom_components;
DROP TABLE IF EXISTS bill_of_materials;
