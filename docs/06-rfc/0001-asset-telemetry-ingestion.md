# RFC-0001: Asset Telemetry Ingestion Architecture

* **Author(s):** AI Architect
* **Status:** Under Review
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
3. **Forwarding Layer:** Continuously reads from SQLite and pushes to the Cloud Kafka/Redpanda cluster. If the network is down, the forwarder pauses and SQLite grows locally.

### 3.2 Cloud Storage (TimescaleDB)
Telemetry is time-series data. We will use **TimescaleDB** (PostgreSQL extension) with hypertables partitioned by time and `physical_asset_id` to store the raw streams for fast aggregation.

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
```

---

## 6. Event Lifecycle

1. **Edge** publishes `TelemetryPayload` to Kafka topic: `telemetry.raw.v1`
2. **Telemetry Service** consumes from Kafka.
3. **Telemetry Service** batch-inserts into TimescaleDB.
4. **Analytics Engine** queries TimescaleDB every 60 seconds to compute OEE and publishes event: `production.performance.oee_computed`.

---

## 7. Open Questions

- [ ] Do we need to compress payloads (e.g., Snappy/GZIP) at the Edge before sending to Kafka to save bandwidth?
- [ ] What is the data retention policy for raw telemetry in TimescaleDB? (e.g., drop raw data after 30 days, keep hourly roll-ups for 1 year).
