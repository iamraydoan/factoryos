# Bounded Context: Maintenance (CMMS) & Asset Health

## 1. Overview
Orchestrates Computerized Maintenance Management System (CMMS) operations: preventive maintenance (PM) schedules, anomaly-triggered work orders, and spare parts management.

---

## 2. Core Entities & Aggregates

* **MaintenanceWorkOrder:** Ticket for repairing or inspecting machinery.
* **PMSchedule:** Recurring preventive maintenance plan based on runtime hours or calendar intervals.
* **AssetHealthMetric:** Real-time vibration, temperature, and wear indicators.

---

## 3. Key Events

* `maintenance.anomaly.detected`
* `maintenance.work_order.dispatched`
* `maintenance.pm_schedule.due`
