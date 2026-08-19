-- +goose Up
CREATE TABLE IF NOT EXISTS qualification_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    person_id UUID NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    person_class_id UUID NOT NULL REFERENCES person_classes(id) ON DELETE RESTRICT,
    work_center_id UUID NOT NULL REFERENCES work_centers(id) ON DELETE CASCADE,
    certified_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,                          -- NULL = no expiry
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(person_id, person_class_id, work_center_id)
);

CREATE INDEX idx_qualification_records_person ON qualification_records(person_id);
CREATE INDEX idx_qualification_records_work_center ON qualification_records(work_center_id);
CREATE INDEX idx_qualification_records_expires ON qualification_records(expires_at) WHERE expires_at IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS qualification_records;
