# Production Operations — Performance

> **ISA-95 Reference:** Part 3, §5 — Production Performance Analysis

## 1. Overview
Aggregates execution data into **Key Performance Indicators (KPIs)** for the facility. The primary metric is OEE (Overall Equipment Effectiveness), computed per Work Unit, per Shift, and per Work Center.

---

## 2. OEE Calculation Model

```
OEE = Availability × Performance × Quality
```

| Component | Formula | Data Source |
|---|---|---|
| **Availability** | Actual Run Time / Planned Production Time | Production Events (run, downtime) |
| **Performance** | (Actual Output × Ideal Cycle Time) / Actual Run Time | JobElementResponse + Routing standard times |
| **Quality** | Good Units / Total Units Started | JobResponse (good qty vs. reject qty) |

---

## 3. KPI Aggregation Levels

```text
Work Unit (Station) OEE
    └── Work Center OEE  (average of Work Units)
            └── Area OEE
                    └── Site OEE
```

KPIs are computed at configurable intervals: **real-time (streaming)**, shift summary, daily, and weekly.

---

## 4. Entity Definitions

| Entity | Description |
|---|---|
| **OEESnapshot** | Point-in-time OEE reading for a Work Unit at a specific timestamp |
| **ShiftPerformanceSummary** | Aggregated OEE + output + downtime report for a completed Shift |
| **ProductionKPIResult** | ISA-95 KPI Result record for a defined measurement period |

---

## 5. Key Events

| Event | Description |
|---|---|
| `production.performance.oee_computed` | OEE snapshot calculated for a Work Unit (emitted every N seconds) |
| `production.performance.shift_summary_generated` | End-of-shift performance report published |
| `production.performance.oee_below_threshold` | OEE drops below configured alert level; Supervisor notified |
