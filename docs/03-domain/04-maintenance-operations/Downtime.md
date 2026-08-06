# Maintenance Operations — Downtime

> **ISA-95 Reference:** Part 3, §7 — Equipment Downtime Tracking

## 1. Overview
**Downtime** records every period during which a Work Unit is not producing. Downtime records feed directly into the OEE Availability calculation and MTBF/MTTR analytics. Accurate categorization is critical for root cause and improvement decisions.

---

## 2. Entity Definitions

| Entity | Description |
|---|---|
| **DowntimeRecord** | A timestamped record of a Work Unit stoppage: start time, end time, duration, category |
| **DowntimeCategory** | Classification of the stoppage reason |
| **DowntimeLedger** | The aggregated downtime log for a Work Unit over a given period |

---

## 3. Downtime Categories (SEMI E10 / OEE Standard)

| Category | Code | Description |
|---|---|---|
| **Unscheduled Downtime** | UD | Unexpected machine failure or fault |
| **Scheduled Downtime** | SD | Planned maintenance, changeover, shift breaks |
| **Engineering Downtime** | ED | Setup, first-article inspection, process development |
| **Standby** | SB | Work Unit available but no Work Order assigned |
| **Non-Scheduled** | NS | Work Unit not required for production (off-shift) |

---

## 4. MTBF & MTTR Formulas

```
MTBF = Total Run Time / Number of Failures
MTTR = Total Downtime (UD only) / Number of Failures
```

These metrics are computed per Physical Asset over a rolling 30-day window by default, configurable per reporting period.

---

## 5. Key Events

| Event | Description |
|---|---|
| `maintenance.downtime.started` | Work Unit transitions out of ACTIVE state; downtime clock starts |
| `maintenance.downtime.categorized` | Technician or system assigns the downtime category |
| `maintenance.downtime.ended` | Work Unit returns to ACTIVE; downtime duration recorded |
| `maintenance.downtime.mtbf_updated` | MTBF metric recalculated for the Physical Asset |
