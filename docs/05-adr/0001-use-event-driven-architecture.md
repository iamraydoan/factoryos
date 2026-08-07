# ADR-0001: Adoption of Event-Driven Architecture (EDA)

* **Status:** Accepted
* **Date:** 2026-08-05
* **Authors:** Founding CTO, Principal Software Architect

---

## 1. Context

FactoryOS needs to orchestrate physical manufacturing operations across multi-tenant enterprise cloud and local factory edge nodes. Legacy MES solutions use synchronous RPC or relational database polling, which scales poorly when handling thousands of machine telemetry feeds and triggers cascading service failures when network partitions occur.

### Decision Drivers
* Need for sub-second reactive response to machine failures and quality defects.
* Requirement for offline edge resilience during cloud connectivity loss.
* Decoupling domain microservices (Production, Maintenance, Quality, Warehouse).
* Regulatory demand for complete, immutable historical audit logs.

---

## 2. Decision

**Option 2: Event-Driven Architecture (EDA)** with append-only event streaming backbone (Kafka/Redpanda).

---

## 3. Consequences

### Positive Impacts
* High throughput telemetry ingestion decoupled from transactional services.
* Edge nodes buffer domain events locally during WAN outages and replay them upon reconnection.
* New capabilities (e.g., Energy Analytics) can consume existing event streams without modifying core production code.

### Negative Impacts & Trade-offs
* Increased system complexity (eventual consistency, idempotent event handlers, dead-letter queues).
* Schema evolution requires strict schema registry governance.

---

## 4. Alternatives Considered

* **Synchronous REST / gRPC microservices architecture:** Rejected due to high coupling and cascading failure risk during edge network drops.
* **Monolithic shared database architecture:** Rejected due to database lock contention under high-frequency telemetry load and lack of offline edge buffering capabilities.

