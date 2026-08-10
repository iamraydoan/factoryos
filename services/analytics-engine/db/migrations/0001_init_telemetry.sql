-- Migration: 0001_init_telemetry.sql
-- Description: Create raw_telemetry hypertable, indexes, retention policy, and continuous aggregates.

-- 0. Migration version tracking table
-- Records applied migrations so future scripts can skip already-executed ones.
CREATE TABLE IF NOT EXISTS schema_migrations (
    version     TEXT        PRIMARY KEY,
    applied_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO schema_migrations (version) VALUES ('0001') ON CONFLICT DO NOTHING;

-- Enable TimescaleDB extension if available
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- 1. Create raw telemetry table
CREATE TABLE IF NOT EXISTS raw_telemetry (
    time TIMESTAMPTZ NOT NULL,
    physical_asset_id UUID NOT NULL,
    metric_name TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    quality TEXT
);

-- 2. Convert to TimescaleDB hypertable partitioned by time (7 days per chunk) if TimescaleDB is loaded
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
        IF NOT EXISTS (SELECT 1 FROM timescaledb_information.hypertables WHERE hypertable_name = 'raw_telemetry') THEN
            PERFORM create_hypertable('raw_telemetry', 'time', chunk_time_interval => INTERVAL '7 days', if_not_exists => TRUE);
        END IF;
    END IF;
END $$;

-- 3. Composite index for fast asset-time queries
CREATE INDEX IF NOT EXISTS ix_asset_time ON raw_telemetry (physical_asset_id, time DESC);

-- 4. Automated 30-day Retention Policy
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
        -- Add retention policy if not already attached
        BEGIN
            PERFORM add_retention_policy('raw_telemetry', INTERVAL '30 days', if_not_exists => TRUE);
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'Retention policy notice: %', SQLERRM;
        END;
    END IF;
END $$;

-- 5. Hourly Continuous Aggregate Materialized View (kept for 365 days)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
        IF NOT EXISTS (SELECT 1 FROM pg_matviews WHERE matviewname = 'telemetry_hourly_summary') THEN
            CREATE MATERIALIZED VIEW telemetry_hourly_summary
            WITH (timescaledb.continuous) AS
            SELECT 
                time_bucket('1 hour', time) AS bucket,
                physical_asset_id,
                metric_name,
                AVG(value) AS avg_value,
                MIN(value) AS min_value,
                MAX(value) AS max_value,
                COUNT(*) AS sample_count
            FROM raw_telemetry
            GROUP BY bucket, physical_asset_id, metric_name;

            PERFORM add_continuous_aggregate_policy('telemetry_hourly_summary',
                start_offset => INTERVAL '3 hours',
                end_offset => INTERVAL '1 hour',
                schedule_interval => INTERVAL '1 hour',
                if_not_exists => TRUE);

            PERFORM add_retention_policy('telemetry_hourly_summary', INTERVAL '365 days', if_not_exists => TRUE);
        END IF;
    END IF;
END $$;
