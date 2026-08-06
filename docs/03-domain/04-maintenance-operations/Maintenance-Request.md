# Maintenance Operations — Maintenance Request

> **ISA-95 Reference:** Part 3, §7 — Maintenance Request (trigger for Maintenance Work Order)

## 1. Overview
A **Maintenance Request** is the signal that something is wrong with a Physical Asset and requires maintenance attention. It is the entry point that creates a Maintenance Work Order. Requests may originate automatically from telemetry thresholds or manually from an Operator or Technician.

---

## 2. Request Origins

| Source | Trigger Mechanism |
|---|---|
| **Automated — Telemetry** | Asset sensor reading exceeds configured threshold (e.g., vibration > 7.5 mm/s) |
| **Automated — OEE Degradation** | Work Unit OEE drops below threshold for N consecutive cycles |
| **Manual — Operator** | Operator reports an anomaly via the terminal one-tap "Report Issue" action |
| **Manual — Technician** | Technician identifies a problem during a PM inspection |

---

## 3. Maintenance Request State Machine

```text
[ RAISED ] ──▶ [ ACKNOWLEDGED ] ──▶ [ WORK_ORDER_CREATED ] ──▶ [ CLOSED ]
                     │
                     ▼
               [ CANCELLED ]  (false alarm confirmed)
```

---

## 4. Severity Classification

| Severity | SLA (Technician on-site) | Auto-Action |
|---|---|---|
| **Critical** | ≤ 15 minutes | Work Unit set to FAULTED; Production WO placed on HOLD |
| **Major** | ≤ 45 minutes | Alert dispatched to Maintenance team |
| **Minor** | ≤ 4 hours | Added to Technician queue |

---

## 5. Key Events

| Event | Description |
|---|---|
| `maintenance.request.raised` | A new Maintenance Request is created (auto or manual) |
| `maintenance.request.acknowledged` | A Technician acknowledges and accepts the request |
| `maintenance.request.work_order_created` | A Maintenance Work Order is generated from this request |
