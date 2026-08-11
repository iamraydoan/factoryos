# CLAUDE.md - FactoryOS AI Guidelines for Claude Code

You are pair-programming as a Principal Software Architect & Senior Engineer on **FactoryOS**, an enterprise event-driven manufacturing operations platform (IIoT / MES / MOM).

---

## 1. Essential Commands & Workflows

### Testing & Verification
```bash
# Run all unit tests across the workspace (Edge Runtime, Platform SDK, Analytics Engine)
make test

# Run all tests explicitly
make test-all

# Run tests for specific components
make test-analytics    # Analytics Engine with -race, -cover, and timeout
make test-ingestion    # Telemetry Ingestion Gateway Service (gRPC)
make test-edge         # Edge Runtime collector, mqtt, forwarder, and store-and-forward buffer
make test-sdk          # Platform SDK (Protobuf stubs, OpenAPI models, Swagger UI)

# Generate code coverage for Analytics Engine
make test-coverage
```

### Build & Execution
```bash
# Build all Go binaries into bin/
make build             # alias for make build-all
make build-analytics   # Compiles to bin/analytics-engine
make build-ingestion   # Compiles to bin/ingestion-service
make build-edge        # Compiles to bin/edge-runtime
make build-simulator   # Compiles to bin/mock-plc-simulator

# Run services locally
make run-analytics
make run-ingestion
make run-edge
make run-simulator
```

### Schema, Protocol Buffers (Buf) & OpenAPI (Swagger)
```bash
# Lint Protobuf schemas
make proto-lint        # Runs buf lint in api/contracts/

# Generate Go stubs and SDK contracts from Protobuf
make proto-gen         # Runs buf generate in api/contracts/

# [Step 1] Bundle all multi-file OpenAPI domain specs into dist/ (via Redocly)
make openapi-bundle    # Auto-scans api/contracts/openapi/**/openapi.yaml
                       # Output: api/contracts/openapi/<domain>/v1/dist/openapi.bundled.yaml

# [Step 2] Generate Go SDK for every domain from their bundled spec
make openapi-gen       # Runs openapi-bundle first, then oapi-codegen for each domain
                       # Output: platform/platform-sdk/go/gen/openapi/<domain>/v1/<domain>.gen.go
                       # Package names: telemetryv1, resourcev1, productionv1, etc.
```

> **Note:** `dist/` and `platform-sdk/go/gen/` are both gitignored — they are
> build artifacts regenerated automatically by `make openapi-gen` in every CI run.

### Local Infrastructure (Docker Compose)
```bash
make infra-up          # Start PostgreSQL (TimescaleDB), Kafka, Redpanda Console, Mosquitto MQTT
make infra-down        # Stop all containers
make infra-ps          # Inspect container status
make infra-logs        # Follow container logs
make clean             # Remove build artifacts (bin/), coverage files, and test binaries
```

---

## 2. Architecture & Tech Stack

* **Repository Model:** Go Monorepo using Go 1.22+ Workspaces (`go.work`).
  * `platform/platform-sdk`: Shared domain models, Protobuf generated code, and platform utilities.
  * `platform/edge-runtime`: Shopfloor signal collector (OPC-UA / MQTT) with offline SQLite store-and-forward buffer.
  * `services/analytics-engine`: Real-time telemetry ingestion (Kafka), threshold rule evaluation, batch DB writer (TimescaleDB / PostgreSQL).
  * `examples/mock-plc-simulator`: High-frequency sensor and PLC telemetry generator.
  * `api/contracts/`: Source of truth for all Protobuf / gRPC schemas and event definitions.
* **Storage & Messaging:**
  * Kafka (telemetry streaming & transactional outbox events).
  * TimescaleDB / PostgreSQL (time-series metrics and relational entities).
  * Mosquitto MQTT (edge IoT device communication).
  * Embedded SQLite (edge-runtime local resilient buffering).

---

## 3. Golden Development Lifecycle Workflow

Always follow this lifecycle before implementing new features:

1. **Check Strategy & Roadmap:** Review `docs/08-roadmap/MILESTONES.md` to identify active milestone goals.
2. **Verify/Author RFC:** Ensure an RFC exists under `docs/06-rfc/` using `0000-rfc-template.md`.
3. **Verify/Author ADR:** If making architectural or infrastructure choices, record an ADR under `docs/05-adr/` using `0000-adr-template.md`.
4. **Verify/Author Epic:** Ensure actionable tasks are documented in `docs/09-epics/`.
5. **Schema-First Contracts:** Define Protobuf schemas in `api/contracts/` and run `make proto-gen` before writing service code.
6. **Implement & Test:** Write unit and integration tests (target > 80% coverage).
7. **Update Checklist:** Mark completed tasks as `[x]` in the corresponding Epic under `docs/09-epics/`.

---

## 4. Engineering & Code Standards

* **Schema-First Mandate:** Never write microservice endpoints, consumers, or data models without matching definitions in `api/contracts/`.
* **Event Naming Convention:** Strictly format event topics and types as `<domain>.<entity>.<past_tense_action>` (e.g., `production.work_order.completed`, `quality.ncr.raised`).
* **Entity Identifiers:** Always use **UUIDv7** (time-ordered UUIDs) for entity primary keys.
* **Edge Resilience:** Code in `platform/edge-runtime` must be resilient to network disconnections using local SQLite store-and-forward queues.
* **No Swallowing Errors:** Log explicit stack traces, context, and error messages. Never ignore errors or use silent fallbacks.
* **Bounded Context Isolation:** Direct cross-service database queries are prohibited. Use asynchronous events or gRPC RPCs.
* **Go Idioms:**
  * Accept interfaces, return concrete structs.
  * Respect `context.Context` cancellation and propagation in all network and DB operations.
  * Keep concurrency safe (use mutexes, atomics, or channels correctly; run tests with `-race`).

---

## 5. Documentation & File Governance

* Keep documentation folders sequential:
  `docs/00-governance`, `docs/01-overview`, `docs/02-business`, `docs/03-domain`, `docs/04-architecture`, `docs/05-adr`, `docs/06-rfc`, `docs/07-api`, `docs/08-roadmap`, `docs/09-epics`, `docs/10-developer-guide`.
* Use semantic uppercase for core files (`PROJECT_BIBLE.md`, `PRODUCT_VISION.md`).
* Use serial IDs: RFCs (`0001-xxxx.md`), ADRs (`0001-xxxx.md`), Epics (`EPIC-001-xxxx.md`).
