# FactoryOS Project Bible

> **The North Star for Engineering Governance, Architecture Principles, & Quality Standards**

---

## 1. Vision & Core Philosophy

FactoryOS is an enterprise-grade, event-driven operational system of action designed to orchestrate modern manufacturing shop floors. 

* **Mental Model:** SAP Context + Palantir Physical Ontology + Datadog Real-time Telemetry + Stripe Developer-first Extensibility.
* **Core Axiom:** *The shop floor never stops.* All systems must tolerate intermittent network partitions, high-frequency signal ingestion, and deterministic offline buffering.

---

## 2. Architectural Principles

1. **Schema-First Specification:** All synchronous RPCs and asynchronous events are defined in `api/contracts/` (Protobuf, AsyncAPI, OpenAPI) before implementation begins.
2. **Event-Driven & Outbox Pattern:** Cross-domain communication occurs asynchronously via events published through the Transactional Outbox pattern.
3. **Domain-Driven Design (DDD):** Domain logic must remain clean, uncoupled from transport protocols, and scoped strictly within defined Bounded Contexts.
4. **Resilient Edge Architecture:** Edge runtimes (`platform/edge-runtime`) operate autonomously with local SQLite store-and-forward buffers.

---

## 3. DDD & Domain Rules

* **Bounded Context Isolation:** Direct database queries across microservice boundaries are strictly prohibited.
* **Event Naming Convention:** `<domain>.<entity>.<past_tense_action>` (e.g., `production.work_order.completed`, `quality.ncr.raised`).
* **Entity Identifiers:** All entities must use UUIDv7 (time-ordered UUIDs) for primary keys.

---

## 4. Engineering & Code Standards

* **Languages:** Go / Java for high-throughput edge and platform microservices; TypeScript (React/Next.js) for Web UI.
* **API Style:** gRPC for internal low-latency RPCs; REST/GraphQL for frontend integration; AsyncAPI/Kafka for event streaming.
* **Error Handling:** Standardized gRPC status codes & RFC-7807 Problem Details for REST APIs.

---

## 5. Security & Performance SLAs

* **Ingestion Throughput:** Ingest 10,000+ OPC-UA/MQTT signals/sec with p99 latency < 50ms at Edge.
* **Command Latency:** Station execution operator commands p99 < 100ms.
* **Security:** Zero-trust mTLS for Edge-to-Cloud communication, RBAC/ABAC role enforcement on all API routes.

---

## 6. Definition of Done (DoD)

A feature or RFC is considered **Done** only when:
- [ ] RFC/ADR is approved and merged into `docs/`.
- [ ] Schema contracts (Protobuf/AsyncAPI) are committed to `api/contracts/`.
- [ ] Unit and integration tests pass with > 80% code coverage.
- [ ] Telemetry metrics (Prometheus/OpenTelemetry) and logs are emitted.
- [ ] Deployment manifests (Helm/Terraform) are updated.
