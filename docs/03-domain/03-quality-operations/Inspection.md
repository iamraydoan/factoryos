# Quality Operations — Inspection

> **ISA-95 Reference:** Part 3, §6 — Quality Test (runtime instance of a Test Specification)

## 1. Overview
An **Inspection** is the runtime execution of a Test Specification for a specific production unit, lot, or Operation. It records the actual measured values and the resulting Pass/Fail verdict.

---

## 2. Entity Definitions

| Entity | Description |
|---|---|
| **Inspection** | A runtime quality test instance linked to a Work Order, Operation, and Lot |
| **InspectionResult** | Recorded actual value for each TestCriteria (numeric, boolean, image attachment) |
| **InspectionVerdict** | Overall Pass / Fail / Conditional conclusion for the Inspection |

---

## 3. Inspection State Machine

```text
[ PENDING ] ──▶ [ IN_PROGRESS ] ──▶ [ PASS ] ──▶ production continues
                      │
                      ▼
                  [ FAIL ] ──▶ triggers Non-Conformance creation
```

---

## 4. SPC Data Recording
For numeric criteria, each recorded value is stored as a **data point** in the time-series historian. SPC control charts (X-bar, R-chart) are computed from these streams to detect process drift before defects occur.

---

## 5. Key Events

| Event | Description |
|---|---|
| `quality.inspection.triggered` | A production event (Operation completed) creates a pending Inspection |
| `quality.inspection.started` | Inspector begins recording results |
| `quality.inspection.passed` | All criteria met; production flow continues |
| `quality.inspection.failed` | One or more criteria breached; NCR creation initiated |
