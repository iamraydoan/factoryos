# RFC-0003: Bidirectional Material Lot Genealogy & Traceability Engine

* **Author(s):** Ray Doan
* **Status:** Draft
* **Target Release:** Milestone 2
* **Created Date:** 2026-08-06

---

## 1. Problem Statement

Regulatory compliance and quality recalls require 100% bidirectional traceability: tracing raw material lot numbers down to individual assembled serial units (Where-Used & As-Built tree).

---

## 2. Goals & Non-Goals

### Goals
- [ ] Record lot allocation at staging bins and station loading.
- [ ] Build DAG (Directed Acyclic Graph) for As-Built material component tree.
- [ ] Provide instant recall query API (find all affected serialized units by raw lot ID).

### Non-Goals
- Warehouse 3D spatial bin optimization.
- Manual paper batch ticket scanning OCR.

---

## 3. Proposed Solution Architecture

*(To be detailed during Milestone 2 design phase)*

---

## 4. API & Schema Contracts

```protobuf
// Protobuf schema definitions for Genealogy events will be added here
```

---

## 5. Data Model & Database Migrations

*(To be detailed: material_lots, serial_genealogy_dag)*

---

## 6. Event Lifecycle

*(To be detailed: warehouse.lot.allocated, quality.genealogy.linked)*

---

## 7. Security & Observability

*(To be detailed: Traceability query performance SLAs, audit log encryption)*

---

## 8. Open Questions

- [ ] Should genealogy graph be stored in PostgreSQL Recursive CTEs or graph database (Neo4j / Memgraph)?

