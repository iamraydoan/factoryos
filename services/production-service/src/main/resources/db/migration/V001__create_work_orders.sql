CREATE TABLE IF NOT EXISTS work_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    material_definition_id UUID NOT NULL,       -- FK to MaterialDefinition (cross-service reference)
    routing_spec_id UUID NOT NULL,              -- FK to ProductRoutingSpec (cross-service reference)
    work_center_id UUID NOT NULL,               -- FK to WorkCenter (cross-service reference)
    target_quantity VARCHAR(50) NOT NULL,        -- decimal string to avoid floating-point precision issues
    unit_of_measure VARCHAR(50) NOT NULL,        -- e.g., "pcs", "kg", "liters"
    state VARCHAR(20) NOT NULL DEFAULT 'draft',  -- state machine: draft, released, dispatched, in_progress, held, completed, closed
    priority VARCHAR(20) NOT NULL DEFAULT 'medium', -- priority level: low, medium, high, urgent
    description TEXT,
    due_date TIMESTAMPTZ,                        -- when the WO must be completed (nullable)
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_work_orders_state ON work_orders(state);
CREATE INDEX idx_work_orders_work_center ON work_orders(work_center_id);
CREATE INDEX idx_work_orders_material ON work_orders(material_definition_id);
CREATE INDEX idx_work_orders_due_date ON work_orders(due_date);
