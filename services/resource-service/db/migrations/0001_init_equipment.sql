-- +goose Up
-- ============================================================================
-- Migration: 0001_init_equipment
-- Description: Create the ISA-95 Equipment Hierarchy tables for FactoryOS.
--
-- The hierarchy follows the ISA-95 standard:
--   Site → Area → Work Center → Work Unit
--
-- References:
--   - EPIC-002: Resource Management & Real-Time Telemetry
--   - api/contracts/resource/v1/equipment.proto
-- ============================================================================


-- ----------------------------------------------------------------------------
-- Sites: Top-level physical factory locations.
-- Example: "Austin Gigafactory", "Berlin Plant"
-- ----------------------------------------------------------------------------
CREATE TABLE sites (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),  -- Unique identifier (UUIDv7 recommended)
    name        TEXT        NOT NULL,                               -- Human-readable site name
    location    TEXT        NOT NULL DEFAULT '',                    -- Physical address or region
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),                 -- Row creation timestamp
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()                  -- Last modification timestamp
);


-- ----------------------------------------------------------------------------
-- Areas: Logical subdivisions within a Site.
-- Example: "Stamping Floor", "Assembly Line", "Quality Lab"
-- ----------------------------------------------------------------------------
CREATE TABLE areas (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),  -- Unique identifier
    site_id     UUID        NOT NULL REFERENCES sites(id) ON DELETE CASCADE,  -- Parent site
    name        TEXT        NOT NULL,                               -- Human-readable area name
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),                 -- Row creation timestamp
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()                  -- Last modification timestamp
);


-- ----------------------------------------------------------------------------
-- Work Centers: Groups of Work Units sharing a capability class.
-- Example: "Press Line 1", "CNC Milling Bay", "Welding Station"
-- ----------------------------------------------------------------------------
CREATE TABLE work_centers (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),  -- Unique identifier
    area_id         UUID        NOT NULL REFERENCES areas(id) ON DELETE CASCADE,  -- Parent area
    name            TEXT        NOT NULL,                               -- Human-readable work center name
    equipment_class TEXT        NOT NULL DEFAULT '',                    -- Capability spec (e.g., "Heavy Press", "5-Axis CNC")
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),                 -- Row creation timestamp
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()                  -- Last modification timestamp
);


-- ----------------------------------------------------------------------------
-- Work Unit Status: Valid operational states for a Work Unit.
-- State machine transitions:
--   available → allocated → in_production → faulted → available
-- ----------------------------------------------------------------------------
CREATE TYPE work_unit_status AS ENUM (
    'available',      -- Idle, ready for allocation
    'allocated',      -- Assigned to a work order but not yet running
    'in_production',  -- Actively producing
    'faulted'         -- Down due to breakdown or maintenance
);


-- ----------------------------------------------------------------------------
-- Work Units: Individual machines or stations on the shop floor.
-- Example: "Press Station 1A", "CNC Mill #3", "Welding Robot #2"
-- ----------------------------------------------------------------------------
CREATE TABLE work_units (
    id                UUID              PRIMARY KEY DEFAULT gen_random_uuid(),  -- Unique identifier
    work_center_id    UUID              NOT NULL REFERENCES work_centers(id) ON DELETE CASCADE,  -- Parent work center
    name              TEXT              NOT NULL,                               -- Human-readable work unit name
    status            work_unit_status  NOT NULL DEFAULT 'available',           -- Current operational status
    physical_asset_id UUID              NULL,                                   -- Optional link to a Physical Asset
    created_at        TIMESTAMPTZ       NOT NULL DEFAULT now(),                 -- Row creation timestamp
    updated_at        TIMESTAMPTZ       NOT NULL DEFAULT now()                  -- Last modification timestamp
);


-- ----------------------------------------------------------------------------
-- Indexes: Optimize foreign key joins and status filtering.
-- ----------------------------------------------------------------------------
CREATE INDEX idx_areas_site_id          ON areas(site_id);            -- FK lookup: areas by site
CREATE INDEX idx_work_centers_area_id   ON work_centers(area_id);     -- FK lookup: work centers by area
CREATE INDEX idx_work_units_work_center ON work_units(work_center_id); -- FK lookup: work units by work center
CREATE INDEX idx_work_units_status      ON work_units(status);        -- Filter: query work units by status


-- +goose Down
-- ============================================================================
-- Rollback: Drop tables in reverse dependency order.
-- ============================================================================

DROP TABLE IF EXISTS work_units;
DROP TYPE IF EXISTS work_unit_status;
DROP TABLE IF EXISTS work_centers;
DROP TABLE IF EXISTS areas;
DROP TABLE IF EXISTS sites;
