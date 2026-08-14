-- +goose Up
-- ============================================================================
-- Migration: 0003_physical_assets
-- Description: Create the Physical Asset registry and installation history.
--
-- ISA-95 distinguishes Physical Assets (real machines with serial numbers)
-- from Work Units (logical scheduling slots). A Physical Asset is "installed
-- at" a Work Unit — if the machine is replaced, the Work Unit persists and
-- only the installation record changes.
--
-- References:
--   - EPIC-002: Resource Management & Real-Time Telemetry
--   - docs/03-domain/04-maintenance-operations/Physical-Asset.md
-- ============================================================================


-- ----------------------------------------------------------------------------
-- Physical Asset Status: Valid operational states for a Physical Asset.
-- State machine transitions:
--   active → faulted → under_maintenance → active / decommissioned
-- ----------------------------------------------------------------------------
CREATE TYPE physical_asset_status AS ENUM (
    'active',             -- Operational and installed
    'faulted',            -- Down due to failure
    'under_maintenance',  -- Being repaired
    'decommissioned'      -- Permanently retired
);


-- ----------------------------------------------------------------------------
-- Physical Assets: Real machines tracked by serial number.
-- Example: "Haas VF-2 #SN-48291", "Fanuc Robot R-2000iC #FR-77421"
-- ----------------------------------------------------------------------------
CREATE TABLE physical_assets (
    id              UUID                  PRIMARY KEY DEFAULT gen_random_uuid(),  -- Unique identifier (UUIDv7 recommended)
    name            TEXT                  NOT NULL,                               -- Human-readable asset name
    serial_number   VARCHAR(255)          NOT NULL UNIQUE,                        -- Manufacturer serial number (unique per machine)
    manufacturer    TEXT                  NOT NULL DEFAULT '',                    -- e.g., "Haas Automation", "Fanuc"
    model           TEXT                  NOT NULL DEFAULT '',                    -- e.g., "VF-2", "R-2000iC/165F"
    asset_type      TEXT                  NOT NULL DEFAULT '',                    -- e.g., "CNC Mill", "Press", "Robot"
    status          physical_asset_status NOT NULL DEFAULT 'active',              -- Current operational status
    installed_at    TIMESTAMPTZ           NULL,                                   -- Denormalized: when current installation began (NULL = not installed)
    created_at      TIMESTAMPTZ           NOT NULL DEFAULT now(),                 -- Row creation timestamp
    updated_at      TIMESTAMPTZ           NOT NULL DEFAULT now()                  -- Last modification timestamp
);


-- ----------------------------------------------------------------------------
-- Physical Asset Installations: Time-bounded link between asset and work unit.
-- Tracks which Physical Asset is (or was) installed at which Work Unit.
-- A Work Unit can have at most ONE active installation (removed_at IS NULL).
-- Example: "Haas VF-2 #SN-48291 installed at CNC Station 1A from Jan 2025"
-- ----------------------------------------------------------------------------
CREATE TABLE physical_asset_installations (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),  -- Unique identifier
    physical_asset_id   UUID        NOT NULL REFERENCES physical_assets(id) ON DELETE CASCADE,  -- Installed asset
    work_unit_id        UUID        NOT NULL REFERENCES work_units(id) ON DELETE CASCADE,       -- Target work unit
    installed_at        TIMESTAMPTZ NOT NULL DEFAULT now(),                 -- When the asset was installed
    removed_at          TIMESTAMPTZ NULL,                                   -- NULL = currently installed; set on uninstall
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),                 -- Row creation timestamp
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()                  -- Last modification timestamp
);


-- ----------------------------------------------------------------------------
-- Indexes: Optimize FK lookups, status filtering, and active installation queries.
-- ----------------------------------------------------------------------------
CREATE INDEX idx_physical_assets_serial    ON physical_assets(serial_number);                              -- Lookup by serial number
CREATE INDEX idx_physical_assets_status    ON physical_assets(status);                                      -- Filter: query assets by status
CREATE INDEX idx_installations_asset       ON physical_asset_installations(physical_asset_id);              -- FK lookup: installations by asset
CREATE INDEX idx_installations_work_unit   ON physical_asset_installations(work_unit_id);                   -- FK lookup: installations by work unit
CREATE INDEX idx_installations_active      ON physical_asset_installations(work_unit_id, removed_at)        -- Partial index: fast lookup of current
    WHERE removed_at IS NULL;                                                                               -- installation per Work Unit


-- +goose Down
-- ============================================================================
-- Rollback: Drop tables in reverse dependency order.
-- ============================================================================

DROP TABLE IF EXISTS physical_asset_installations;
DROP TABLE IF EXISTS physical_assets;
DROP TYPE IF EXISTS physical_asset_status;
