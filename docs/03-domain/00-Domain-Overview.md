# Domain Model Overview — Bounded Context Map

FactoryOS is structured into **5 ISA-95 Level 3 functional domains** plus a cross-cutting **Resource Management** layer. Each domain is a self-contained Bounded Context with its own data model, state machines, and events.

---

## ISA-95 Level 3 Module Map

```text
┌─────────────────────────────────────────────────────────────────┐
│                      ERP / Level 4                              │
│              (Production Orders, BOMs, Demand)                  │
└───────────────────────────┬─────────────────────────────────────┘
                            │  Production Request
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                   01 · Resource Management                      │
│        Equipment Hierarchy · Personnel · Material Definition    │
└──────┬──────────────────────┬──────────────────────────────────-┘
       │                      │  (shared master data)
       ▼                      ▼
┌──────────────┐   ┌──────────────────────┐   ┌─────────────────┐
│     02       │   │         03           │   │       04        │
│  Production  │──▶│      Quality         │──▶│   Maintenance   │
│  Operations  │   │      Operations      │   │   Operations    │
└──────┬───────┘   └──────────────────────┘   └─────────────────┘
       │  (material consumption)
       ▼
┌──────────────┐
│     05       │
│  Inventory   │
│  Operations  │
└──────────────┘
```

---

## Module Descriptions

| Module | ISA-95 Reference | Responsibility |
|---|---|---|
| **01 Resource Management** | ISA-95 Part 2 — Resource Model | Defines the shared master data: Equipment Hierarchy, Personnel roles & skills, Material & Product definitions. Consumed by all other modules. |
| **02 Production Operations** | ISA-95 Part 3 — §5 Production Ops | Receives Production Requests from ERP, creates Schedules, dispatches Work Orders to the floor, tracks Execution and Performance (OEE). |
| **03 Quality Operations** | ISA-95 Part 3 — §6 Quality Ops | Defines Test Specifications, records Inspection results, manages NCR lifecycle (Disposition → CAPA). |
| **04 Maintenance Operations** | ISA-95 Part 3 — §7 Maintenance Ops | Manages Physical Asset registry, responds to Maintenance Requests, schedules Preventive Maintenance, tracks Downtime. |
| **05 Inventory Operations** | ISA-95 Part 3 — §8 Inventory Ops | Manages Material Lot lifecycle (receive → stage → consume), records material movements, builds the As-Built Genealogy tree. |

---

## Cross-Domain Event Contracts

| Publisher | Event | Subscriber(s) |
|---|---|---|
| Production Operations | `production.work_order.released` | Inventory Operations (stage materials) |
| Production Operations | `production.operation.completed` | Quality Operations (trigger inspection) |
| Inventory Operations | `inventory.lot.staged` | Production Operations (unblock operation) |
| Quality Operations | `quality.ncr.raised` | Inventory Operations (quarantine lot), Production Operations (hold station) |
| Maintenance Operations | `maintenance.work_order.completed` | Production Operations (release station) |
| Production Operations | `production.asset.downtime_started` | Maintenance Operations (create request) |
