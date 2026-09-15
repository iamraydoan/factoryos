# Changelog

All notable changes to the **FactoryOS** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- **Work Order (`services/production-service`):** JPA entity with UUIDv7 identifiers, Flyway migrations `V001` (table + indexes) and `V002` (composite pagination index), and `ListWorkOrders` over gRPC with cursor-based pagination.
- **Cursor & offset pagination library (`services/production-service`):** Keyset pagination with versioned opaque cursors (`SortKey`, `SortCriteria`, `Cursor`, `KeysetCondition`, `CursorPage`, `CursorPageRequest`) and page-number pagination (`OffsetPage`, `OffsetPageRequest`). 116 unit tests.
- **REST pagination DTOs (`services/production-service`):** `CursorPageResponse` and `OffsetPageResponse` implementing the `{ data, pagination }` envelope from [PAGINATION_DESIGN.md](docs/07-api/PAGINATION_DESIGN.md), with REST/gRPC key-name mapping (`limit`/`cursor` vs `pageSize`/`pageToken`).
- **Work Order REST API (`services/production-service`):** `GET /api/v1/work-orders` with page-based and cursor-based pagination, plus `work_center_id` and `state` filters.
- **Error taxonomy contract (`api/contracts/openapi/common/v1/schemas/errors.yaml`):** `ErrorResponse` and `ErrorDetail` schemas implementing RFC 9457 Problem Details with FactoryOS extension members (`code`, `errors[]`, `retryable`, `traceId`). The previously-orphaned common schema is now canonical.
- **Error handling design standard ([docs/07-api/ERROR_HANDLING.md](docs/07-api/ERROR_HANDLING.md)):** Language-neutral specification — the nine error categories, their HTTP and gRPC mappings, both wire formats, the code taxonomy and naming rules, and the logging contract.
- **ADR-0007:** Global error handling standard — errors classified by category rather than transport status, adapters own the transport mapping, RFC 9457 with a `code` extension.
- **ProductRoutingSpec & ProductRoutingStep (`services/resource-service`):** Versioned routing with step sequencing, FK validation to WorkCenter and MaterialDefinition, and 28 unit tests. Completes EPIC-002 Material Definition section.
- **BOM & BOMComponent (`services/resource-service`):** Bill of Materials with versioning (`material_definition_id + version` unique), child component linking with quantity/unit-of-measure, FK validation, and 26 unit tests.
- **MaterialClass & MaterialDefinition (`services/resource-service`):** Material category and material definition CRUD with `part_number` uniqueness, optional JSON `specification`, FK validation, and 27 unit tests.
- **Shift & ShiftAssignment (`services/resource-service`):** Shift definitions (TIME columns) and many-to-many shift assignments with upsert, 3-filter listing, FK validation, and 34 unit tests.

### Removed
- **Per-domain `ErrorResponse` copies (`telemetry/v1/schemas/common.yaml`):** The local `ErrorResponse` schema copy is deleted — all specs now `$ref` the canonical `common/v1/schemas/errors.yaml`. The shape itself was kept and expanded with RFC 7807 fields.

### Fixed
- **`docs/07-api/PAGINATION_DESIGN.md`:** Rewritten as a language-neutral standard — implementation code removed in favour of wire formats, token format, validation rules, and language-agnostic implementation requirements. Corrected the earlier Java section, which omitted `SortCriteria`, `CursorPageResponse`, and `OffsetPageResponse` and used signatures that do not match the shipped library. Also fixed a malformed code fence that had caused everything after it to render as a single code block.

### Documentation Note
- The error handling standard above is **design-first**: the taxonomy, mappings, and contract schemas are fixed, while the implementation in each language is not yet done. Implementations must satisfy [ERROR_HANDLING.md](docs/07-api/ERROR_HANDLING.md) rather than define their own vocabulary.

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
