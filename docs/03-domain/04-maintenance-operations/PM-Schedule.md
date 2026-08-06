# Maintenance Operations — PM Schedule

> **ISA-95 Reference:** Part 3, §7 — Preventive Maintenance Scheduling

## 1. Overview
A **PM Schedule** defines the recurring maintenance tasks for a Physical Asset class. It triggers Preventive Maintenance Work Orders automatically based on time elapsed or runtime hours accumulated.

---

## 2. Entity Definitions

| Entity | Description |
|---|---|
| **PMSchedule** | A recurring maintenance plan linked to a PhysicalAssetClass |
| **PMTask** | An individual maintenance step within a PM plan (e.g., "Replace filter", "Lubricate bearing") |
| **PMTrigger** | The condition that fires the schedule: calendar interval, runtime hours, or cycle count |
| **PMDue** | A pending PM event created when a trigger condition is met |

---

## 3. Trigger Types

| Trigger Type | Example | Source |
|---|---|---|
| **Calendar** | Every 30 days | System clock |
| **Runtime Hours** | Every 500 operating hours | Asset telemetry (OPC-UA runtime counter) |
| **Cycle Count** | Every 10,000 production cycles | JobElementResponse cycle count |
| **Condition-Based** | Vibration trend exceeds baseline | Asset sensor stream (predictive, Phase 4+) |

---

## 4. PM Schedule State Machine

```text
[ ACTIVE ] ──▶ [ DUE ] ──▶ [ WORK_ORDER_CREATED ] ──▶ [ COMPLETED ] ──▶ [ ACTIVE ]
                              (next interval resets)
```

---

## 5. Key Events

| Event | Description |
|---|---|
| `maintenance.pm_schedule.due` | Trigger condition met; PM Work Order auto-created |
| `maintenance.pm_schedule.overdue` | PM due date passed without completion; escalation alert sent |
| `maintenance.pm_schedule.completed` | PM Work Order closed; next trigger interval restarted |
