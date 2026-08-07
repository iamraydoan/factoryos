# RFC-0002: Guided Work Order Execution & Station State Machine Engine

* **Author(s):** Ray Doan
* **Status:** Draft
* **Target Release:** Milestone 2
* **Created Date:** 2026-08-06

---

## 1. Problem Statement

Operators on the shop floor need interactive guidance for work order execution, step routings, component assembly validation, and automated station hold triggers upon defect detection.

---

## 2. Goals & Non-Goals

### Goals
- [ ] Implement durable state machine engine for Work Orders & Operations.
- [ ] Build Web UI Operator Terminal for shop-floor workstations.
- [ ] Support barcode/QR code scan verification for part loading.

### Non-Goals
- Real-time CAD parsing inside operator browser.
- Direct ERP financial ledger postings (handled via domain events).

---

## 3. Proposed Solution Architecture

*(To be detailed during Milestone 2 design phase)*

---

## 4. API & Schema Contracts

```protobuf
// Protobuf schema definitions for Work Order execution will be added here
```

---

## 5. Data Model & Database Migrations

*(To be detailed: work_orders, operation_steps, station_state_logs)*

---

## 6. Event Lifecycle

*(To be detailed: production.work_order.started, production.work_order.completed)*

---

## 7. Security & Observability

*(To be detailed: RBAC station authorization, operation duration metrics)*

---

## 8. Open Questions

- [ ] Should work order state persistence utilize Temporal durable workflow or PostgreSQL state machine table?

