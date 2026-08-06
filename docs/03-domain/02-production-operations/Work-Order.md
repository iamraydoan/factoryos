# Production Operations — Work Order

> **ISA-95 Reference:** Part 2, Section 4 — Job Order (Operations Schedule Segment)

## 1. Overview
A **Work Order** (ISA-95: Job Order) is the primary execution unit dispatched to the shop floor. It instructs the floor to produce a specific quantity of a MaterialDefinition, following a defined Routing, at a target Work Center, within a scheduled time window.

---

## 2. Entity Definitions

| Entity | ISA-95 Term | Description |
|---|---|---|
| **WorkOrder** | Job Order | The floor-level execution directive: product, quantity, routing, Work Center, and schedule window |
| **WorkOrderLine** | Job Order Segment | A sub-quantity breakdown when a WO is split across multiple shifts or Work Units |

---

## 3. Work Order State Machine

```text
[ DRAFT ]
    │
    ▼
[ RELEASED ] ──────────────────────────────┐
    │                                      │
    ▼                                      │
[ DISPATCHED ]  (assigned to Work Unit)    │ Emergency
    │                                      │ re-priority
    ▼                                      │
[ IN_PROGRESS ] ◀─────────────────────────┘
    │         │
    │         ▼
    │    [ HELD ]  (quality hold, material shortage, maintenance fault)
    │         │
    │         ▼
    │    [ RESUMED ]
    │
    ▼
[ COMPLETED ]
    │
    ▼
[ CLOSED ]  (confirmed & reported back to ERP)
```

---

## 4. Dispatch Rules
- A Work Order may only be dispatched to a Work Unit that belongs to a Work Center with the required **EquipmentClass**.
- The assigned Operator must hold a valid **QualificationRecord** for the required PersonClass.
- All required material lots must be **staged** at the Work Unit before dispatch.

---

## 5. Key Events

| Event | Description |
|---|---|
| `production.work_order.released` | WO is approved and ready for dispatch |
| `production.work_order.dispatched` | WO is assigned to a specific Work Unit and Operator |
| `production.work_order.held` | WO execution is paused (reason recorded) |
| `production.work_order.completed` | All Operations finished; quantity confirmed |
| `production.work_order.closed` | WO results reported back to ERP |
