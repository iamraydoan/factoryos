# EPIC-002: Resource Management & Real-Time Telemetry

* **Milestone:** Milestone 1
* **Status:** In Progress
* **Domain:** `docs/03-domain/01-resource-management/`
* **RFC:** `RFC-0001`

---

## Task Checklist

### Equipment Hierarchy (ISA-95 Resource Model)
- [x] Implement CRUD API for Site, Area, Work Center, Work Unit entities.
- [x] Implement Equipment Class definition and Work Unit capability assignment.
- [x] Implement Work Unit status state machine: Available → Allocated → In Production → Faulted.
- [x] Link Work Unit to Physical Asset (installation record).

### Personnel & Qualification
- [x] Implement Person and PersonClass (role) management.
- [x] Implement QualificationRecord: certify a Person for a PersonClass at a Work Center.
- [x] Implement QualificationRecord expiry tracking and alert.
- [x] Implement Shift and ShiftAssignment management.

### Material Definition
- [ ] Implement MaterialClass and MaterialDefinition CRUD API.
- [ ] Implement Bill of Materials (BOM) with versioning support.
- [ ] Implement ProductRoutingSpec: link MaterialDefinition to Work Center sequence.

### Telemetry & OEE
- [x] Build Go Edge Runtime OPC-UA collector (`platform/edge-runtime`).
- [x] Configure SQLite store-and-forward buffer for offline tolerance.
- [x] Implement MQTT ingestion endpoint for sensor telemetry.
- [x] Define Protobuf `TelemetryIngestionService` and `RecordBatch` contracts (`api/contracts/telemetry/v1/ingestion.proto`).
- [x] Implement Go Edge Runtime gRPC Forwarder with SQLite store-and-forward drain loop (`platform/edge-runtime/forwarder`).
- [x] Implement Cloud Ingestion Gateway Service (`TelemetryIngestionService` gRPC handler producing to Kafka `telemetry.raw.v1`).
- [x] Implement TimescaleDB `raw_telemetry` hypertable migration, 30-day retention policy & continuous aggregate roll-up (`services/analytics-engine/db`).
- [x] Implement high-throughput Go batch writer using `pgx.CopyFrom` in `services/analytics-engine`.
- [x] Implement Kafka stream consumer with Snappy decompression in `services/analytics-engine`.
- [x] Implement real-time in-memory threshold alert evaluator.
- [x] Implement real-time OEE streaming aggregator (Availability × Performance × Quality) per Work Unit.
- [x] Implement OEE alert when Work Unit drops below configured threshold.


