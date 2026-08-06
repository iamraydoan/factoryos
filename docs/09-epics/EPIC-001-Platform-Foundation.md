# EPIC-001: Platform Architecture & Infrastructure Foundation

* **Milestone:** Milestone 1
* **Status:** In Progress

---

## Task Checklist

### Core Setup & Governance
- [x] Create `PROJECT_BIBLE.md` and repository standards.
- [x] Define Protobuf / AsyncAPI contracts directory hierarchy.
- [ ] Configure PostgreSQL with Transactional Outbox table schema.
- [ ] Setup Apache Kafka event streaming infrastructure.
- [ ] Setup Zitadel for Identity & Access Management (OIDC).
- [ ] Setup Traefik API Gateway for gRPC routing.
- [ ] Setup Valkey (Redis Drop-in Replacement) for distributed caching.
- [ ] Setup Observability stack (OpenTelemetry, Prometheus, Jaeger, Loki, Grafana).

### Platform Services
- [ ] Implement `platform-sdk` event producer/consumer wrappers (Java & Go).
- [ ] Setup Temporal.io cluster for Durable Workflow Engine.
