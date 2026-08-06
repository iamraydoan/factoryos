# Maintenance Operations — Work Order

> **ISA-95 Reference:** Part 3, §7 — Maintenance Work Order

## 1. Overview
A **Maintenance Work Order** is the execution directive for a maintenance task — either Corrective (in response to a fault) or Preventive (scheduled). It dispatches a Technician to a Physical Asset with specific task instructions.

---

## 2. Work Order Types

| Type | Trigger | Description |
|---|---|---|
| **Corrective (CM)** | Maintenance Request (fault or anomaly) | Unplanned repair to restore asset to operational status |
| **Preventive (PM)** | PM Schedule (time or runtime trigger) | Planned inspection and servicing to prevent failure |

---

## 3. Work Order State Machine

```text
[ CREATED ] ──▶ [ ASSIGNED ] ──▶ [ IN_PROGRESS ] ──▶ [ COMPLETED ] ──▶ [ CLOSED ]
                                        │
                                        ▼
                                  [ ON_HOLD ]  (waiting for spare parts)
                                        │
                                        ▼
                                  [ RESUMED ]
```

---

## 4. Task Execution
Each Maintenance Work Order contains:
- **Assigned Technician** and required **skill certification**.
- **Task Instructions** (checklist of maintenance steps).
- **Spare Parts** required (linked to Inventory Operations for parts reservation).
- **Estimated Duration** (from PM Schedule or historical CM average).

---

## 5. Key Events

| Event | Description |
|---|---|
| `maintenance.work_order.created` | WO generated from a Request or PM Schedule trigger |
| `maintenance.work_order.assigned` | Technician dispatched |
| `maintenance.work_order.completed` | Maintenance task finished; Physical Asset status set to ACTIVE |
| `maintenance.work_order.closed` | Work Order results recorded; MTTR updated |
