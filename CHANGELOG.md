# Changelog

All notable changes to the **FactoryOS** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

---

## [v0.4.1] - 2026-10-07

### Fixed
- **Release pipeline (`release.yml`):** `v0.4.0`'s tag pointed at a workflow that pinned Go 1.22 and never generated the gitignored protobuf SDK, so `edge-runtime` failed to build (`does not contain package .../go/gen/telemetry/v1`). The job now takes its toolchain from `go.work` and runs `make proto-gen` before cross-compiling.
- **Go toolchain and protobuf/grpc versions (`go.work`, all `go.mod`):** CI's Go pin (1.22) was below `go.work`'s `go 1.26.5`, which makes Go ignore the workspace and silently drop the cross-module `replace` directives. CI and release now use `go-version-file: go.work`. Protobuf and gRPC were also pinned inconsistently across the six modules (protobuf v1.34.1–v1.36.12, grpc v1.64.0–v1.82.1); all now use protobuf v1.36.12 and gRPC v1.84.0, matching the codegen plugins.

---

## [v0.4.0] - 2026-10-07

### Added
- **Shared error taxonomy & transport adapters (`libs/factoryos-common-error`, `libs/factoryos-common-error-adapters`):** Category-based taxonomy with `DomainException` / `CommonErrorCode`, and one adapter per protocol — RFC 9457 over HTTP, `google.rpc.Status` over gRPC. No transport dependency in the taxonomy module; adapters self-register via Spring Boot auto-configuration, with per-service gRPC overrides through `GrpcStatusOverrideProvider`.
- **Error taxonomy adoption (`services/production-service`):** `WorkOrderErrorCode`; cursor failures split into `INVALID_CURSOR` / `UNSUPPORTED_CURSOR_VERSION` / `CURSOR_SORT_KEY_MISMATCH`; page-size and malformed-UUID failures classified as client errors. Fixes the four defects in ADR-0007 §1.
- **Gateway CORS (`deploy/traefik/dynamic/cors.yaml`, `production.yaml`, `docker-compose.yml`):** `api-cors` middleware — Traefik answers `OPTIONS` preflight itself. Env-driven allow-list `CORS_ALLOW_ORIGIN_LIST`, credentials on, `Max-Age 3600`. Attached to `production-rest`.
- **Three env profiles (`Makefile`, `.env.*.example`):** `make setup-env-host` | `setup-env-docker` | `setup-env-devcontainer` — each installs root `.env` **and** `services/production-service/.env` in one step.
- **Developer tool installation (`Makefile`, `.devcontainer/devcontainer.json`):** `make install-tools` for protobuf, Buf, OpenAPI generator, and Redocly CLI.
- **Dev-only Swagger UI (`docker-compose.yml`, `Makefile`):** `swagger-ui` on the `docs` profile, `make docs-up` / `docs-down`.
- **Swagger UI serves any domain spec (`docker-compose.docs.yml`):** `SWAGGER_SPEC_FILE` selects the spec path relative to the mounted `/spec/`.
- **Work Order (`services/production-service`):** JPA entity with UUIDv7 identifiers, Flyway `V001`/`V002`, and `ListWorkOrders` over gRPC with cursor pagination.
- **Cursor & offset pagination library (`services/production-service`):** Keyset cursors (`SortKey`, `SortCriteria`, `Cursor`, `KeysetCondition`, `CursorPage`, `CursorPageRequest`) and page-number pagination (`OffsetPage`, `OffsetPageRequest`). 116 unit tests.
- **REST pagination DTOs (`services/production-service`):** `CursorPageResponse` / `OffsetPageResponse` on the `{ data, meta }` envelope, with REST/gRPC key mapping (`limit`/`cursor` vs `pageSize`/`pageToken`).
- **Work Order REST API (`services/production-service`):** `GET /api/v1/work-orders` with page + cursor pagination and `work_center_id` / `state` filters.
- **Production Work Order contract (`api/contracts/openapi/production/v1/`):** REST projection of `ProductionService` — `POST/GET /work-orders`, `GET /work-orders/{id}`, 7 transition actions, RFC 9457 errors. New shared `ConflictError` (409).
- **Error taxonomy contract (`api/contracts/openapi/common/v1/schemas/errors.yaml`):** `ErrorResponse` / `ErrorDetail` per RFC 9457 with `code`, `errors[]`, `retryable`, `traceId`.
- **Error handling design standard (`docs/07-api/ERROR_HANDLING.md`):** Language-neutral spec — nine categories, HTTP/gRPC mappings, wire formats, naming rules, logging contract.
- **ADR-0007:** Global error handling standard — errors classified by category, adapters own the transport mapping.
- **ProductRoutingSpec & ProductRoutingStep (`services/resource-service`):** Versioned routing with step sequencing and FK validation. 28 unit tests.
- **BOM & BOMComponent (`services/resource-service`):** Bill of Materials with versioning (`material_definition_id + version` unique) and child components. 26 unit tests.
- **MaterialClass & MaterialDefinition (`services/resource-service`):** Material CRUD with `part_number` uniqueness and optional JSON `specification`. 27 unit tests.
- **Shift & ShiftAssignment (`services/resource-service`):** Shift definitions and many-to-many assignments with upsert and 3-filter listing. 34 unit tests.
- **Collection query standard (`docs/07-api/QUERY_CONVENTIONS.md`, `common/v1/parameters/query.yaml`):** filtering (`eq`/`ne`/`gt`/`gte`/`lt`/`lte`/`IN`/`like`), `-`-prefix multi-field `sort`, `q`, sparse `fields`, and shared parameter components.
- **Production service gateway routing (`docker-compose.yml`, `deploy/traefik/`, `Dockerfile`):** `production-service` containerized (multi-stage Maven on `eclipse-temurin:21`) and routed via Traefik file-provider config. REST and gRPC routes verified end-to-end.
- **Env-templated gateway backends (`docker-compose.yml`, `deploy/traefik/dynamic/*.yaml`):** Traefik backend URLs templated with `{{env}}` (`PRODUCTION_REST_BACKEND`, `PRODUCTION_GRPC_BACKEND`).
- **Telemetry collection queries (`api/contracts/openapi/telemetry/v1/`):** `latest` / `alerts` gain filter/sort/`fields` on `{ data, meta }`; `history` gains `sort` + `fields` with a max-buckets guard.

### Changed
- **API design skill (`.agents/skills/api-design/SKILL.md`):** Aligned to house conventions — `{ data, meta }`, RFC 9457 errors with a stable `code`, camelCase wire names, `422` for validation and `400` for malformed input.
- **Error taxonomy — split `VALIDATION` into `MALFORMED` and `VALIDATION` (`docs/07-api/ERROR_HANDLING.md`, `ADR-0007`):** unparseable → **400**, parsed-but-invalid → **422**; `INVALID_CURSOR` stays at 400. gRPC unchanged (`INVALID_ARGUMENT`).
- **OpenAPI contracts (`api/contracts/openapi/`):** Added `UnprocessableEntityError` and wired `422` into work-order mutations; `BadRequestError` clarified as malformed-input-only.
- **Error-handling tutorial (`docs/07-api/error-handling-tutorial/`):** `ErrorCategory` gains `MALFORMED(400, 3)`, `VALIDATION` moves to `422`; samples updated to match.
- **Agent instructions consolidated (`AGENTS.md`, `CLAUDE.md`):** `AGENTS.md` is the single source of truth; `CLAUDE.md` is now a git symlink to it.
- **Unified platform OpenAPI (`api/contracts/openapi/openapi.yaml`):** Single contract entrypoint via `$ref`; per-domain wrappers deleted, one `platformv1` SDK; overlay renamed to `overlays/gateway.overlay.yaml`.
- **Business APIs only in the platform contract:** `/healthz`, `/ready`, `/stats` removed — internal ops endpoints. Merged spec holds only `/work-orders*` + `/telemetry/*`.
- **Collection envelope key `pagination` → `meta` (breaking, pre-release):** REST collections are now `{ data, meta }`.
- **Wire page numbers are 1-indexed (breaking, pre-release):** `?page=1` is the first page; below 1 is `PAGE_NUMBER_OUT_OF_RANGE` (400). Internals stay 0-indexed.
- **Telemetry query params to camelCase + pagination-aligned (breaking, pre-release):** `asset_id` → `assetId`, `metric_name` → `metricName`; `latest` limit 50 → 20, max 1000 → 100.
- **OpenAPI contract vs gateway auth split (`api/contracts/openapi/`, `Makefile`):** Sources are auth-free for `oapi-codegen`; gateway auth lives in the overlay, applied by `make openapi-gateway`. See ADR-0008.
- **Java services migrated to Spring gRPC (`services/*/pom.xml`):** `net.devh` starter replaced by Boot 4.1.1's `spring-boot-starter-grpc-server` (all versions managed); `protobuf-maven-plugin` swapped to `io.github.ascopes`.
- **All five Java services now inherit `services/pom.xml`:** Spring Boot 3.3.0 → 4.1.1; `spring-boot-starter-web` → `-webmvc`; the parent no longer forces `data-jpa`, `spring-kafka`, `postgresql`, and `lombok` on every service.

### Removed
- **Go `swaggerui` placeholder handler (`platform/platform-sdk/go/swaggerui/`):** Deleted — no service mounted it. Interactive docs are served only by the Docker `swagger-ui` service.
- **Per-domain `ErrorResponse` copies (`telemetry/v1/schemas/common.yaml`):** Deleted — all specs `$ref` the canonical `common/v1/schemas/errors.yaml`.

### Fixed
- **`services/production-service`:** Out-of-range `limit`/`page`, malformed cursor, and malformed `workCenterId` now return classified client errors (422 for page size, 400 for cursor/UUID; gRPC `INVALID_ARGUMENT`) instead of HTTP 500 / gRPC `INTERNAL`, each with a machine-readable `code`.
- **`docs/07-api/PAGINATION_DESIGN.md`:** Rewritten as a language-neutral standard; corrected the Java section (missing `SortCriteria`, `CursorPageResponse`, `OffsetPageResponse`) and a malformed code fence.

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
