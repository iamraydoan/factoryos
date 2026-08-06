# Production Operations — Execution

> **ISA-95 Reference:** Part 2, Section 4 — Job Response (Production Response)

## 1. Overview
**Execution** captures the actual production response as it unfolds on the floor — recording what actually happened versus what was planned. This data feeds directly into the Performance module for OEE and variance analysis.

---

## 2. Entity Definitions

| Entity | ISA-95 Term | Description |
|---|---|---|
| **JobResponse** | Job Response | Actual results recorded for a Work Order: quantity good, quantity rejected, actual start/end times |
| **JobElementResponse** | Job Element Response | Actual results for each individual Operation: duration, operator, output, material consumed |
| **ProductionEvent** | — | A timestamped event during execution (start, stop, hold, fault, resume) used for downtime analysis |

---

## 3. Actual vs. Planned Variance

| Metric | Planned Source | Actual Source |
|---|---|---|
| Quantity | Work Order target qty | JobResponse.quantity_good |
| Start Time | Production Schedule | First `operation.started` event |
| End Time | Production Schedule | `work_order.completed` event |
| Cycle Time | Routing standard time | JobElementResponse.actual_duration |

---

## 4. Key Events

| Event | Description |
|---|---|
| `production.execution.production_event_recorded` | A timestamped floor event is captured (used for downtime ledger) |
| `production.execution.quantity_reported` | Operator submits actual good/reject quantity for an Operation |
| `production.execution.variance_detected` | Actual cycle time or quantity deviates beyond tolerance from plan |
