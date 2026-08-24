# Changelog

All notable changes to the **FactoryOS** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- **MaterialClass & MaterialDefinition (`services/resource-service`):** Material category and material definition CRUD with `part_number` uniqueness, optional JSON `specification`, FK validation, and 27 unit tests.
- **Shift & ShiftAssignment (`services/resource-service`):** Shift definitions (TIME columns) and many-to-many shift assignments with upsert, 3-filter listing, FK validation, and 34 unit tests.

---

## [v0.3.0] - 2026-08-20

### Added
- **Person & PersonClass (`services/resource-service`):** Person/role CRUD with `person_status` enum, FK to `person_classes`, 24 unit tests.
- **Qualification Record (`services/resource-service`):** Certify persons for roles at work centers with `expires_at`, FK validation, 22 unit tests.
- **Qualification Expiry (`services/resource-service`):** `CheckExpiringQualifications` RPC with RFC3339 `before` filter, 6 unit tests.
- **OEE Threshold Alerting (`services/analytics-engine`):** Per-component and composite OEE alerts with configurable thresholds and cooldown dedup, 28 tests at 97.1% coverage.
- **Equipment Class & Capability (`services/resource-service`):** Capability type definitions and many-to-many work unit capability assignment with JSONB properties, 40 tests.
- **Physical Asset & Installation (`services/resource-service`):** Machine registry with state machine and time-bounded installation records with transactional install/uninstall, 47 tests.

---

## [v0.2.0] - 2026-08-12

### Added
- **ISA-95 Equipment Hierarchy (`services/resource-service`):** Sites → Areas → Work Centers → Work Units with Goose auto-migrations, gRPC CRUD, status state machine, and env-based config.
- **Analytics Engine (`services/analytics-engine`):** TimescaleDB `pgx.CopyFrom` batch writer, Kafka consumer with Snappy decompression, dynamic alert engine, real-time OEE aggregator, and Prometheus metrics.
- **ADR-0004:** High-throughput telemetry ingestion architecture (Go, Kafka Snappy, `pgx.CopyFrom`).
- **ADR-0006:** OEE streaming aggregation — fixed window with hard reset.
- **Edge Fleet Management idea ([INBOX.md](INBOX.md)):** Cloud-based monitoring dashboard and anti-spoofing device auth concept.

### Changed
- **CI Pipeline:** Build and test all services, per-module coverage gate (>80%).
- **Makefile:** Added `test-coverage` targets per module; removed `GO_ENV`.
- **RFC-0001:** Status → Approved, Snappy compression decision, 30-day retention policy, security & observability section.
- **EPIC-002:** Refined Telemetry & OEE task breakdown.
- **Docs governance:** Standardized all RFCs and ADRs to templates.

---

## [v0.1.0] - 2026-08-07

### Added
- **Edge Runtime SQLite Buffer (`platform/edge-runtime/buffer`):** WAL-mode offline-resilient outbox with Protobuf payload storage and retry tracking.
- **Telemetry Collector & Sync Worker (`platform/edge-runtime/collector`):** Sensor metric packaging with background cloud-flush and backoff on outages.
- **MQTT Subscriber (`platform/edge-runtime/mqtt`):** Async subscriber on `factoryos/telemetry/+/readings` with Protobuf and JSON dual decoder.
- **Mock PLC Simulator (`examples/mock-plc-simulator`):** Real-time telemetry generator with external JSON config.
- **GitHub Actions CI/CD:** Automated testing with >80% coverage gate and cross-platform release builds (linux/windows/darwin).
- **Edge Infrastructure:** Mosquitto MQTT broker in Docker Compose; developer backlog file ([INBOX.md](INBOX.md)).
