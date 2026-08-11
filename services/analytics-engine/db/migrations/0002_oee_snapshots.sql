-- +goose Up
-- Description: Create OEE snapshots hypertable for real-time Overall Equipment Effectiveness tracking.

-- 1. Create OEE snapshots table
CREATE TABLE IF NOT EXISTS oee_snapshots (
    time            TIMESTAMPTZ     NOT NULL,
    asset_id        UUID            NOT NULL,
    availability    DOUBLE PRECISION NOT NULL,
    performance     DOUBLE PRECISION NOT NULL,
    quality         DOUBLE PRECISION NOT NULL,
    oee             DOUBLE PRECISION NOT NULL,
    run_time_ms     BIGINT          NOT NULL,
    planned_time_ms BIGINT          NOT NULL,
    total_output    BIGINT          NOT NULL,
    good_output     BIGINT          NOT NULL
);

-- 2. Convert to TimescaleDB hypertable (1-day chunks for OEE data)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
        IF NOT EXISTS (SELECT 1 FROM timescaledb_information.hypertables WHERE hypertable_name = 'oee_snapshots') THEN
            PERFORM create_hypertable('oee_snapshots', 'time', chunk_time_interval => INTERVAL '1 day', if_not_exists => TRUE);
        END IF;
    END IF;
END $$;

-- 3. Composite index for fast asset-time queries
-- Includes time in the unique constraint to prevent duplicate snapshots within the
-- same interval (e.g., from timer drift or snapshot loop double-fire).
CREATE UNIQUE INDEX IF NOT EXISTS ix_oee_asset_time ON oee_snapshots (asset_id, time DESC);

-- 4. Retention policy: keep OEE data for 90 days
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
        BEGIN
            PERFORM add_retention_policy('oee_snapshots', INTERVAL '90 days', if_not_exists => TRUE);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'OEE retention policy notice: %', SQLERRM;
        END;
    END IF;
END $$;

-- +goose Down
DROP TABLE IF EXISTS oee_snapshots;
