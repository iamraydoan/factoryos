-- +goose Up
-- ============================================================================
-- Migration: 0004_installation_unique_constraints
-- Description: Add partial unique indexes to enforce one-active-installation
--              constraints at the database level.
--
-- Without these constraints, concurrent InstallAsset requests can race past
-- the application-level SELECT guards and create duplicate active installations.
-- The partial unique indexes make this impossible at the DB level.
--
-- References:
--   - Code review: Critical #1 — TOCTOU race in InstallAsset
-- ============================================================================

-- One active installation per Physical Asset (cannot be installed at two Work Units simultaneously)
CREATE UNIQUE INDEX idx_one_active_install_per_asset
    ON physical_asset_installations (physical_asset_id)
    WHERE removed_at IS NULL;

-- One active installation per Work Unit (cannot have two assets installed simultaneously)
CREATE UNIQUE INDEX idx_one_active_install_per_work_unit
    ON physical_asset_installations (work_unit_id)
    WHERE removed_at IS NULL;


-- +goose Down
-- ============================================================================
-- Rollback: Drop the unique partial indexes.
-- ============================================================================

DROP INDEX IF EXISTS idx_one_active_install_per_asset;
DROP INDEX IF EXISTS idx_one_active_install_per_work_unit;
