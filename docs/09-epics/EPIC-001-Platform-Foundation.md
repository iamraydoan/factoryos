# EPIC-001: Platform Architecture & Infrastructure Foundation

* **Milestone:** Milestone 1
* **Status:** In Progress

---

## Task Checklist

### Core Setup & Governance
- [x] Create `PROJECT_BIBLE.md` and repository standards.
- [x] Define Protobuf / AsyncAPI contracts directory hierarchy.

### Base Project & Build Pipeline Scaffolding
- [x] Configure Protobuf compilation pipeline using `buf` (`buf.yaml`, `buf.gen.yaml`) for Go and Java.
- [x] Setup Go workspace (`go.work` / `go.mod`) for `platform/edge-runtime`, `platform/platform-sdk`, and `services/analytics-engine`.
- [x] Setup Java / Spring Boot base project structure for core microservices (`services/production-service`, `quality-service`, `warehouse-service`, `maintenance-service`).

### Infrastructure & Event Streaming
- [ ] Configure PostgreSQL with Transactional Outbox table schema & Logical Replication (`wal_level = logical`).
- [ ] Setup Debezium Kafka Connect service for Change Data Capture (CDC) outbox streaming.
- [ ] Setup Apache Kafka event streaming infrastructure.
- [ ] Setup Zitadel for Identity & Access Management (OIDC).
- [ ] Setup Traefik API Gateway for gRPC routing.
- [ ] Setup Valkey (Redis Drop-in Replacement) for distributed caching.
- [ ] Setup Observability stack (OpenTelemetry, Prometheus, Jaeger, Loki, Grafana).

### Platform Services & SDK
- [ ] Implement `platform-sdk` event producer/consumer wrappers with transactional outbox helpers (Java & Go).
- [ ] Setup Temporal.io cluster for Durable Workflow Engine.
