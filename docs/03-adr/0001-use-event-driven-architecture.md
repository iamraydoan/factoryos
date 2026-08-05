# ADR 0001: Adoption of Event-Driven Architecture (EDA)

* **Status:** Accepted
* **Deciders:** Founding CTO, Principal Software Architect
* **Date:** 2026-08-05

## Context & Problem Statement
FactoryOS needs to orchestrate physical manufacturing operations across multi-tenant enterprise cloud and local factory edge nodes. Legacy MES solutions use synchronous RPC or relational database polling, which scales poorly when handling thousands of machine telemetry feeds and triggers cascading service failures when network partitions occur.

## Decision Drivers
* Need for sub-second reactive response to machine failures and quality defects.
* Requirement for offline edge resilience during cloud connectivity loss.
* Decoupling domain microservices (Production, Maintenance, Quality, Warehouse).
* Regulatory demand for complete, immutable historical audit logs.

## Considered Options
1. Synchronous REST / gRPC microservices architecture with relational database storage.
2. Event-Driven Architecture (EDA) with append-only event streaming backbone.
3. Monolithic shared database architecture.

## Decision Outcome
**Option 2: Event-Driven Architecture (EDA)**.

### Positive Consequences
* High throughput telemetry ingestion decoupled from transactional services.
* Edge nodes buffer domain events locally during WAN outages and replay them upon reconnection.
* New capabilities (e.g., Energy Analytics) can consume existing event streams without modifying core production code.

### Negative Consequences / Tradeoffs
* Increased system complexity (eventual consistency, idempotent event handlers, dead-letter queues).
* Schema evolution requires strict schema registry governance.
