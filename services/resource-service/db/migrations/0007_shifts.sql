-- +goose Up
CREATE TABLE IF NOT EXISTS shifts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS shift_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    person_id UUID NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    shift_id UUID NOT NULL REFERENCES shifts(id) ON DELETE CASCADE,
    work_center_id UUID NOT NULL REFERENCES work_centers(id) ON DELETE CASCADE,
    effective_from TIMESTAMPTZ NOT NULL DEFAULT now(),
    effective_to TIMESTAMPTZ,                      -- NULL = open-ended (currently assigned)
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(person_id, shift_id, work_center_id)
);

CREATE INDEX idx_shift_assignments_person ON shift_assignments(person_id);
CREATE INDEX idx_shift_assignments_shift ON shift_assignments(shift_id);
CREATE INDEX idx_shift_assignments_work_center ON shift_assignments(work_center_id);
CREATE INDEX idx_shift_assignments_effective ON shift_assignments(effective_to) WHERE effective_to IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS shift_assignments;
DROP TABLE IF EXISTS shifts;
