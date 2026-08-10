# RFC-0001: Asset Telemetry Ingestion Architecture

* **Author(s):** AI Architect
* **Status:** Approved
* **Target Release:** Milestone 1
* **Created Date:** 2026-08-06

---

## 1. Problem Statement

To compute real-time OEE and trigger automated maintenance alerts, FactoryOS needs to ingest high-frequency telemetry data (temperature, vibration, cycle counts, machine state) from thousands of Physical Assets on the shop floor. 

The shop floor environment is prone to intermittent network drops. If the cloud connection goes down, telemetry data must not be lost. When connectivity returns, the system must handle the sudden burst of buffered data without crashing.

---

## 2. Goals & Non-Goals

### Goals
- [x] Ingest high-frequency telemetry (up to 100Hz per sensor) with low latency.
- [x] Ensure zero data loss during network partitions lasting up to 24 hours.
- [x] Standardize the telemetry payload regardless of the underlying machine protocol (OPC-UA, Modbus, MQTT).

### Non-Goals
- Real-time control (Level 2 SCADA). FactoryOS will *read* telemetry, but will not *write* control signals back to the PLC in milliseconds.
- Complex Event Processing (CEP) at the edge. The edge only buffers and forwards; analytics happen in the cloud.

---

## 3. Proposed Solution Architecture

### 3.1 The Store-and-Forward Edge Runtime
We will deploy a lightweight **Go-based Edge Runtime** on a local IPC (Industrial PC) inside the factory network. 

1. **Ingestion Layer:** Connects to PLCs via OPC-UA or local MQTT brokers.
2. **Buffer Layer:** Writes all incoming payloads to a local embedded **SQLite** database.
3. **Forwarding Layer:** Continuously reads from SQLite and pushes to the Cloud Kafka/Redpanda cluster using native Snappy compression. If the network is down, the forwarder pauses and SQLite grows locally.

### 3.2 Cloud Ingestion & TimescaleDB Storage
A dedicated **Go-based Telemetry & Analytics Engine (`services/analytics-engine`)** consumes from Kafka topic `telemetry.raw.v1`:
1. **In-Memory Stream Processing:** Evaluates sensor alarm thresholds in RAM for immediate anomaly alerting.
2. **High-Throughput Micro-Batching:** Accumulates 500–1,000 readings (or flushes every 200ms) and uses **`pgx.CopyFrom` (PostgreSQL Binary COPY Protocol)** to stream batches directly into TimescaleDB hypertables (`raw_telemetry`), achieving 100,000+ writes/sec without SQL parser overhead.
3. **Automated Lifecycle Policies:** Drops raw data after 30 days and maintains 1-year hourly continuous aggregate roll-ups (`telemetry_hourly_summary`).


---

## 4. API & Schema Contracts

All telemetry payloads emitted by the Edge Runtime to the Cloud must conform to the following Protobuf schema. 

> *Note: This schema will be stored in `api/contracts/telemetry/v1/ingestion.proto`*

```protobuf
syntax = "proto3";

package factoryos.telemetry.v1;

import "google/protobuf/timestamp.proto";

// The standardized payload sent from the Edge to the Cloud
message TelemetryPayload {
  // UUIDv7 of the Physical Asset emitting the data
  string physical_asset_id = 1;
  
  // The exact time the reading was taken at the edge (NOT cloud receive time)
  google.protobuf.Timestamp edge_timestamp = 2;
  
  // Array of sensor readings to support batching
  repeated SensorReading readings = 3;
}

message SensorReading {
  // E.g., "temperature_celsius", "spindle_vibration", "cycle_count"
  string metric_name = 1;
  
  // The value of the reading
  double value = 2;
  
  // Optional: Data quality flag (e.g., "GOOD", "BAD", "UNCERTAIN")
  string quality = 3;
}
```

---

## 5. Data Model & Database Migrations

### TimescaleDB Hypertable
```sql
CREATE TABLE raw_telemetry (
    time TIMESTAMPTZ NOT NULL,
    physical_asset_id UUID NOT NULL,
    metric_name TEXT NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    quality TEXT
);

-- Convert to TimescaleDB hypertable partitioned by time (7 days per chunk)
SELECT create_hypertable('raw_telemetry', 'time', chunk_time_interval => INTERVAL '7 days');

-- Create index for fast querying by asset
CREATE INDEX ix_asset_time ON raw_telemetry (physical_asset_id, time DESC);

-- Automated Data Retention Policy (Default: Drop raw data after 30 days)
-- Configurable via ENV: TELEMETRY_RAW_RETENTION_DAYS (default: 30 days)
SELECT add_retention_policy('raw_telemetry', INTERVAL '30 days');

-- Continuous Aggregate for Hourly Roll-ups (Kept for 1 year)
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

SELECT add_continuous_aggregate_policy('telemetry_hourly_summary',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');

SELECT add_retention_policy('telemetry_hourly_summary', INTERVAL '365 days');
```

---

## 6. Event Lifecycle

1. **Edge** publishes `TelemetryPayload` to Kafka topic: `telemetry.raw.v1` using native **Snappy / Zstd compression** (`compression.type=snappy`).
2. **Telemetry Service** consumes from Kafka (transparent decompression by Kafka client).
3. **Telemetry Service** batch-inserts into TimescaleDB hypertable `raw_telemetry`.
4. **TimescaleDB Background Workers** continuously run retention policy (purge raw chunks older than 30 days) and update hourly continuous aggregates.
5. **Analytics Engine** queries TimescaleDB every 60 seconds to compute OEE and publishes event: `production.performance.oee_computed`.

---

## 7. Security & Observability

### Security & Authentication
- **Transport Security:** Edge-to-Cloud telemetry streaming over Kafka/Redpanda is encrypted using TLS 1.3.
- **Authentication:** Edge IPC instances authenticate using X.509 client certificates (mTLS) or SASL/SCRAM tokens stored in Edge secure enclave / environment secrets.
- **Data Integrity:** Protobuf payloads include edge-generated SHA256 checksums to verify message integrity during store-and-forward replay.

### Observability & Metrics
- **Prometheus Metrics:**
  - `telemetry_ingested_messages_total`: Count of ingested messages labeled by `asset_id` and `quality`.
  - `telemetry_ingestion_latency_seconds`: Histogram of latency from `edge_timestamp` to TimescaleDB commit time.
  - `edge_buffer_depth_records`: Number of records queued in local SQLite buffer when offline.
- **Distributed Tracing:** OpenTelemetry trace contexts (`traceparent`) injected into Kafka record headers by Edge Runtime for end-to-end tracing.

---

## 8. Open Questions

- [x] **Payload Compression at Edge:**
  - **Decision:** Enable native Kafka producer-level compression using **Snappy** (or **Zstd**).
  - **Rationale:** Native Kafka compression occurs at the transport batch layer before pushing to broker. It achieves 60-80% bandwidth reduction with minimal CPU overhead on Edge IPCs and requires zero custom application decompression code in downstream consumers.
- [x] **Data Retention Policy in TimescaleDB:**
  - **Decision:** Raw telemetry data is retained for **30 days** (`TELEMETRY_RAW_RETENTION_DAYS=30d`), while hourly continuous aggregates (`telemetry_hourly_summary`) are retained for **365 days** (`TELEMETRY_HOURLY_RETENTION_DAYS=365d`).
  - **Configurability:** Controlled via environment variables during database migration/service startup scripts (`TELEMETRY_RAW_RETENTION_DAYS`, `TELEMETRY_HOURLY_RETENTION_DAYS`).



